package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"

	"netease-music-rag/backend/internal/model"
	"netease-music-rag/backend/internal/repository"

	"github.com/robfig/cron/v3"
)

type WorkflowService struct {
	neteaseClient *NeteaseClient
	llmClient     *LLMClient
	repo          *repository.SongRepo
	cron          *cron.Cron
	phone         string
}

func NewWorkflowService(nc *NeteaseClient, lc *LLMClient, r *repository.SongRepo, phone string) *WorkflowService {
	c := cron.New()
	return &WorkflowService{
		neteaseClient: nc,
		llmClient:     lc,
		repo:          r,
		cron:          c,
		phone:         phone,
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

	if s.phone != "" {
		if err := s.neteaseClient.Login(s.phone); err != nil {
			log.Printf("Netease login failed (continuing): %v", err)
		}
	}

	recommendPlaylists, err := s.neteaseClient.GetDailyRecommendPlaylist()
	if err != nil || recommendPlaylists == nil {
		return fmt.Errorf("failed to fetch daily recommend playlists: %w", err)
	}

	finalSongList := make([]*model.NeteaseSongDTO, 0)
	existSongID := make(map[int64]bool)

	for idx, playlist := range *recommendPlaylists {
		log.Printf("Processing playlist %d/%d (id=%d)", idx+1, len(*recommendPlaylists), playlist.ID)

		playlistDetail, err := s.neteaseClient.GetDetailPlaylist(playlist.ID)
		if err != nil || playlistDetail == nil {
			log.Printf("Skipping playlist %d: %v", playlist.ID, err)
			continue
		}

		songs := playlistDetail.Tracks
		if len(songs) == 0 {
			continue
		}

		// Shuffle so we don't always pick the first track
		rand.Shuffle(len(songs), func(i, j int) {
			songs[i], songs[j] = songs[j], songs[i]
		})

		for _, song := range songs {
			if existSongID[song.ID] {
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

			lyric, err := s.neteaseClient.GetSongLyrics(song.ID)
			if err != nil || lyric == nil {
				log.Printf("Skipping song %d (lyrics error): %v", song.ID, err)
				continue
			}

			llmAnalysis, err := s.llmClient.AnalyzeSong(context.Background(), &song, *lyric)
			if err != nil || llmAnalysis == nil {
				log.Printf("Skipping song %d (LLM error): %v", song.ID, err)
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

			// One song per playlist
			break
		}

		break
	}

	log.Printf("Collected %d songs, saving to DB...", len(finalSongList))

	if err := s.repo.SaveSongs(finalSongList); err != nil {
		return err
	}

	log.Println("Daily recommendation job finished.")
	return nil
}

// RunEmbeddingJob loads songs that have LLM analysis but no vector embedding,
// generates embeddings in batches, and writes them back to the DB.
// Safe to run repeatedly — already-embedded songs are skipped by the query.
func (s *WorkflowService) RunEmbeddingJob(ctx context.Context) error {
	const batchSize = 50
	log.Println("Starting embedding job...")

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
			}
		}
	}

	log.Println("Embedding job finished.")
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
