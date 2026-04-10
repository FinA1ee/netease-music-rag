package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"netease-music-rag/backend/internal/model"
	"netease-music-rag/backend/internal/repository"

	"github.com/robfig/cron/v3"
)

// withRetry calls fn up to maxAttempts times, sleeping delay between failures.
// It returns the first successful (value, nil-error) pair, or the last error.
func withRetry[T any](maxAttempts int, delay time.Duration, label string, fn func() (T, error)) (T, error) {
	var (
		result T
		err    error
	)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}
		log.Printf("%s: attempt %d/%d failed: %v", label, attempt, maxAttempts, err)
		if attempt < maxAttempts {
			time.Sleep(delay)
		}
	}
	return result, err
}

type WorkflowService struct {
	neteaseClient *NeteaseClient
	llmClient     *LLMClient
	repo          *repository.SongRepo
	cron          *cron.Cron
	phone         string
	bus           *EventBus
}

func NewWorkflowService(nc *NeteaseClient, lc *LLMClient, r *repository.SongRepo, phone string, bus *EventBus) *WorkflowService {
	c := cron.New()
	return &WorkflowService{
		neteaseClient: nc,
		llmClient:     lc,
		repo:          r,
		cron:          c,
		phone:         phone,
		bus:           bus,
	}
}

// StartCron schedules:
//   - RunDailyJob  at 00:05 — fetch playlists, analyse with LLM, persist songs
//   - RunEmbeddingJob at 01:05 — embed all songs that have analysis but no vector yet
func (s *WorkflowService) StartCron() {
	if _, err := s.cron.AddFunc("5 0 * * *", func() {
		if err := s.RunDailyJob(); err != nil {
			log.Printf("Cron daily job error: %v", err)
		}
	}); err != nil {
		log.Fatalf("Error adding daily cron job: %v", err)
	}

	if _, err := s.cron.AddFunc("5 1 * * *", func() {
		if err := s.RunEmbeddingJob(context.Background()); err != nil {
			log.Printf("Cron embedding job error: %v", err)
		}
	}); err != nil {
		log.Fatalf("Error adding embedding cron job: %v", err)
	}

	s.cron.Start()
	log.Println("Cron jobs scheduled: analysis @ 00:05, embedding @ 01:05")
}

// RunDailyJob fetches recommended playlists, picks one song per playlist,
// enriches each with lyrics and LLM analysis, then persists them to the DB.
func (s *WorkflowService) RunDailyJob() error {
	log.Println("Starting daily recommendation job...")
	s.bus.emit(EvJobStarted, map[string]any{"message": "Daily recommendation job started"})

	if s.phone != "" {
		if err := s.neteaseClient.Login(s.phone); err != nil {
			log.Printf("Netease login failed (continuing): %v", err)
		}
	}

	recommendPlaylists, err := s.neteaseClient.GetDailyRecommendPlaylist()
	if err != nil || recommendPlaylists == nil {
		return fmt.Errorf("failed to fetch daily recommend playlists: %w", err)
	}

	const (
		maxPlaylists        = 10 // process at most this many playlists per run
		maxSongsPerPlaylist = 5  // collect at most this many songs per playlist
	)

	finalSongList := make([]*model.NeteaseSongDTO, 0)
	existSongID := make(map[int64]bool)

	for idx, playlist := range *recommendPlaylists {
		if idx >= maxPlaylists {
			break
		}
		log.Printf("Processing playlist %d/%d (id=%d)", idx+1, len(*recommendPlaylists), playlist.ID)

		playlistDetail, err := withRetry(5, 3*time.Second,
			fmt.Sprintf("playlist %d detail", playlist.ID),
			func() (*model.DetailPlaylistData, error) {
				return s.neteaseClient.GetDetailPlaylist(playlist.ID)
			},
		)
		if err != nil || playlistDetail == nil {
			log.Printf("Skipping playlist %d after retries: %v", playlist.ID, err)
			continue
		}

		s.bus.emit(EvPlaylistProcessing, map[string]any{
			"id":          playlistDetail.ID,
			"name":        playlistDetail.Name,
			"trackCount":  len(playlistDetail.Tracks),
			"coverImgUrl": playlistDetail.CoverImgUrl,
			"index":       idx + 1,
			"total":       min(len(*recommendPlaylists), 10),
		})

		songs := playlistDetail.Tracks
		if len(songs) == 0 {
			continue
		}

		// Shuffle so we don't always pick the first track
		rand.Shuffle(len(songs), func(i, j int) {
			songs[i], songs[j] = songs[j], songs[i]
		})

		// Batch-check which songs from this playlist are already in the DB
		trackIDs := make([]int64, len(songs))
		for i, s := range songs {
			trackIDs[i] = s.ID
		}
		dbExisting, err := s.repo.GetExistingSongIDs(context.Background(), trackIDs)
		if err != nil {
			log.Printf("Failed to check existing songs for playlist %d: %v", playlist.ID, err)
			dbExisting = map[int64]bool{} // non-fatal: process all songs
		}

		songsCollected := 0
		for _, song := range songs {
			if songsCollected >= maxSongsPerPlaylist {
				break
			}
			// Skip songs already processed in this run
			if existSongID[song.ID] {
				continue
			}
			// Skip songs already stored in the DB (no need to re-analyse)
			if dbExisting[song.ID] {
				log.Printf("Song %d (%s) already in DB, skipping", song.ID, song.Name)
				continue
			}

			song.Playlist = model.Playlist{
				ID:              playlistDetail.ID,
				Name:            playlistDetail.Name,
				CoverImgUrl:     playlistDetail.CoverImgUrl,
				Description:     playlistDetail.Description,
				Tags:            playlistDetail.Tags,
				SubscribedCount: playlistDetail.SubscribedCount,
			}

			lyric, lErr := withRetry(5, 3*time.Second,
				fmt.Sprintf("song %d lyrics", song.ID),
				func() (*string, error) {
					return s.neteaseClient.GetSongLyrics(song.ID)
				},
			)
			if lErr != nil || lyric == nil {
				log.Printf("Song %d: all lyric retries exhausted, skipping", song.ID)
				s.bus.emit(EvSongSkipped, map[string]any{
					"songId": song.ID,
					"name":   song.Name,
					"reason": "lyrics unavailable",
				})
				continue
			}
			song.Lyric = *lyric

			llmAnalysis, err := s.llmClient.AnalyzeSong(context.Background(), &song, *lyric)
			if err != nil || llmAnalysis == nil {
				log.Printf("Skipping song %d (LLM error): %v", song.ID, err)
				s.bus.emit(EvSongSkipped, map[string]any{
					"songId": song.ID,
					"name":   song.Name,
					"reason": "LLM analysis failed",
				})
				continue
			}

			kw, _ := json.Marshal(llmAnalysis.Keywords)
			st, _ := json.Marshal(llmAnalysis.Style)
			mo, _ := json.Marshal(llmAnalysis.Mood)
			th, _ := json.Marshal(llmAnalysis.Theme)
			fe, _ := json.Marshal(llmAnalysis.Features)

			song.LlmData = &model.NeteaseSongLLMAnalysis{
				Keywords: string(kw),
				Style:    string(st),
				Mood:     string(mo),
				Theme:    string(th),
				Features: string(fe),
			}

			finalSongList = append(finalSongList, &song)
			existSongID[song.ID] = true
			songsCollected++

			artistNames := make([]string, 0, len(song.Ar))
			for _, a := range song.Ar {
				artistNames = append(artistNames, a.Name)
			}
			s.bus.emit(EvSongAnalysed, map[string]any{
				"songId":  song.ID,
				"name":    song.Name,
				"artists": artistNames,
				"style":   llmAnalysis.Style,
				"mood":    llmAnalysis.Mood,
			})

			// Rate-limit guard: give the LLM API breathing room between calls
			log.Printf("Sleeping 10s before next LLM call...")
			time.Sleep(10 * time.Second)
		}
	}

	log.Printf("Collected %d songs, saving to DB...", len(finalSongList))

	if err := s.repo.SaveSongs(finalSongList); err != nil {
		return err
	}

	log.Println("Daily recommendation job finished.")
	s.bus.emit(EvJobCompleted, map[string]any{
		"jobType": "daily",
		"saved":   len(finalSongList),
		"message": fmt.Sprintf("Daily job complete — %d songs saved", len(finalSongList)),
	})

	// Run embedding strictly after the fetch+analyse phase — same goroutine, sequential.
	log.Println("Starting embedding job after daily job completion...")
	if err := s.RunEmbeddingJob(context.Background()); err != nil {
		log.Printf("Embedding job error: %v", err)
	}

	return nil
}

// RunEmbeddingJob loads songs that have LLM analysis but no vector embedding,
// generates embeddings in batches, and writes them back to the DB.
// Safe to run repeatedly — already-embedded songs are skipped by the query.
func (s *WorkflowService) RunEmbeddingJob(ctx context.Context) error {
	const batchSize = 50
	log.Println("Starting embedding job...")
	s.bus.emit(EvEmbeddingStarted, map[string]any{"message": "Embedding backfill job started"})
	totalEmbedded := 0

	for {
		songs, err := s.repo.GetSongsNeedingEmbedding(ctx, batchSize)
		if err != nil {
			return fmt.Errorf("GetSongsNeedingEmbedding: %w", err)
		}
		if len(songs) == 0 {
			break
		}

		log.Printf("Embedding batch of %d songs...", len(songs))
		for i := range songs {
			song := &songs[i]
			text := BuildEmbeddingText(song)
			embedding, err := s.llmClient.GetEmbedding(ctx, text)
			if err != nil {
				log.Printf("Song %d (%s): embedding failed: %v", song.SongID, song.Name, err)
				continue
			}
			if err := s.repo.UpdateEmbedding(ctx, song.SongID, embedding); err != nil {
				log.Printf("Song %d (%s): DB update failed: %v", song.SongID, song.Name, err)
			} else {
				totalEmbedded++
			}
		}
	}


	log.Println("Embedding job finished.")
	s.bus.emit(EvEmbeddingDone, map[string]any{
		"jobType": "embedding",
		"total":   totalEmbedded,
		"message": fmt.Sprintf("Embedding complete — %d songs embedded", totalEmbedded),
	})
	return nil
}

// Search finds songs semantically similar to the query string.
func (s *WorkflowService) Search(ctx context.Context, query string, limit int) ([]model.Songs, error) {
	embedding, err := s.llmClient.GetEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}
	return s.repo.SearchSimilarSongs(ctx, embedding, limit)
}
