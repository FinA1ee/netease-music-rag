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

// StartCron schedules RunDailyJob to run every day at 00:05.
func (s *WorkflowService) StartCron() {
	_, err := s.cron.AddFunc("5 0 * * *", func() {
		if err := s.RunDailyJob(); err != nil {
			log.Printf("Cron daily job error: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Error adding cron job: %v", err)
	}
	s.cron.Start()
	log.Println("Cron job scheduled at 00:05 daily.")
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

// Search finds songs semantically similar to the query string.
// NOTE: requires LLMClient.GetEmbedding to be implemented.
func (s *WorkflowService) Search(ctx context.Context, query string, limit int) ([]model.Songs, error) {
	return nil, fmt.Errorf("Search not yet implemented: LLMClient.GetEmbedding is missing")
}
