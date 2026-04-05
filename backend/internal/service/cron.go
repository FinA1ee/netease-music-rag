package service

import (
	"netease-music-rag/backend/internal/repository"

	"github.com/robfig/cron/v3"
)

type WorkflowService struct {
	neteaseClient *NeteaseClient
	llmClient     *LLMClient
	repo          *repository.SongRepo
	cron          *cron.Cron
}

func NewWorkflowService(nc *NeteaseClient, lc *LLMClient, r *repository.SongRepo) *WorkflowService {
	// Schedule everyday at 00:05
	c := cron.New()
	return &WorkflowService{
		neteaseClient: nc,
		llmClient:     lc,
		repo:          r,
		cron:          c,
	}
}

// func (s *WorkflowService) StartCron() {
// 	_, err := s.cron.AddFunc("5 0 * * *", s.RunDailyJob)
// 	if err != nil {
// 		log.Fatalf("Error adding cron job: %v", err)
// 	}
// 	s.cron.Start()
// 	log.Println("Cron job scheduled at 00:05 daily.")
// }

// func (s *WorkflowService) RunDailyJob() {
// 	log.Println("Starting daily recommendation job...")
// 	ctx := context.Background()

// 	songs, err := s.neteaseClient.GetDailyRecommendations()
// 	if err != nil {
// 		log.Printf("Failed to fetch daily recommendations: %v", err)
// 		return
// 	}
// 	log.Printf("Fetched %d daily recommended songs.", len(songs))

// 	for _, nSong := range songs {
// 		if s.repo.HasSongLLMAnalyzed(ctx, nSong.ID) {
// 			log.Printf("Skip song %d: already analyzed.", nSong.ID)
// 			continue
// 		}

// 		lyrics, err := s.neteaseClient.GetLyric(nSong.ID)
// 		if err != nil {
// 			log.Printf("Failed to get lyric for %d: %v", nSong.ID, err)
// 			continue
// 		}

// 		// Simplify lyric length if excessively long to save LLM tokens
// 		lines := strings.Split(lyrics, "\n")
// 		clearLyrics := ""
// 		for _, line := range lines {
// 			idx := strings.Index(line, "]")
// 			if idx != -1 {
// 				clearLyrics += line[idx+1:] + " "
// 			}
// 		}
// 		if len(clearLyrics) > 2000 {
// 			clearLyrics = clearLyrics[:2000]
// 		}

// 		artistName := ""
// 		var artists []model.Artist
// 		if len(nSong.Ar) > 0 {
// 			artistName = nSong.Ar[0].Name
// 			for _, ar := range nSong.Ar {
// 				artists = append(artists, model.Artist{ID: ar.ID, Name: ar.Name})
// 			}
// 		}

// 		// Analysis
// 		analysis, err := s.llmClient.AnalyzeSong(ctx, nSong.Name, artistName, clearLyrics)
// 		if err != nil {
// 			log.Printf("LLM analysis failed for %d: %v", nSong.ID, err)
// 			continue
// 		}

// 		// Embedding
// 		embedStr := fmt.Sprintf("Song: %s. Artist: %s. Description: %s", nSong.Name, artistName, analysis.Description)
// 		embVec, err := s.llmClient.GetEmbedding(ctx, embedStr)
// 		if err != nil {
// 			log.Printf("Embedding failed for %d: %v", nSong.ID, err)
// 			continue
// 		}

// 		songData := &model.Songs{
// 			SongID:        nSong.ID,
// 			Name:          nSong.Name,
// 			Duration:      nSong.Dt,
// 			Artists:       artists,
// 			Album:         model.Album{ID: nSong.Al.ID, Name: nSong.Al.Name},
// 			AlbumCoverURL: nSong.Al.PicUrl,
// 			Lyric:         clearLyrics,
// 			Description:   analysis.Description,
// 			Style:         analysis.Style,
// 			Mood:          analysis.Mood,
// 		}

// 		err = s.repo.SaveSong(ctx, songData, embVec)
// 		if err != nil {
// 			log.Printf("Failed to save song %d: %v", nSong.ID, err)
// 		} else {
// 			log.Printf("Successfully analyzed and saved song: %s", nSong.Name)
// 		}
// 	}
// 	log.Println("Daily recommendation job finished.")
// }

// func (s *WorkflowService) Search(ctx context.Context, query string, limit int) ([]model.Song, error) {
// 	embVec, err := s.llmClient.GetEmbedding(ctx, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return s.repo.SearchSimilarSongs(ctx, embVec, limit)
// }
