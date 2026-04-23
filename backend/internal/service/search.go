package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"netease-music-rag/backend/internal/model"
	"netease-music-rag/backend/internal/repository"
)

// SearchService handles natural-language song recommendation via vector similarity.
type SearchService struct {
	llmClient         *LLMClient
	repo              *repository.SongRepo
	minPopularity     float32
	distanceThreshold float32
	vectorWeight      float32
	lexicalWeight     float32
}

func NewSearchService(
	lc *LLMClient,
	repo *repository.SongRepo,
	minPopularity float32,
	distanceThreshold float32,
	vectorWeight float32,
	lexicalWeight float32,
) *SearchService {
	return &SearchService{
		llmClient:         lc,
		repo:              repo,
		minPopularity:     minPopularity,
		distanceThreshold: distanceThreshold,
		vectorWeight:      vectorWeight,
		lexicalWeight:     lexicalWeight,
	}
}

var (
	artistDirectiveRe = regexp.MustCompile(`(?i)(?:歌手|artist|singer)\s*[:：]\s*([^\n,，。;；]+)`)
	albumDirectiveRe  = regexp.MustCompile(`(?i)(?:专辑|album)\s*[:：]\s*([^\n,，。;；]+)`)
)

// Search embeds the natural-language query and returns the top-k most similar songs
// using cosine similarity against the pgvector embedding column.
//
// Flow:
//  1. Embed the raw query text → []float32 via Gemini embedding model
//  2. Query the DB: ORDER BY embedding <=> query_vector LIMIT k
//  3. Map DB rows → []model.SongResult for the API response
func (s *SearchService) Search(ctx context.Context, req model.SearchRequest) (*model.SearchResponse, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query must not be empty")
	}
	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.MinPopularity <= 0 {
		req.MinPopularity = s.minPopularity
	}
	if req.ExactArtist == "" || req.ExactAlbum == "" {
		artist, album := parsePromptFilters(req.Query)
		if req.ExactArtist == "" {
			req.ExactArtist = artist
		}
		if req.ExactAlbum == "" {
			req.ExactAlbum = album
		}
	}

	// 1. Embed the query
	embedding, err := s.llmClient.GetEmbedding(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// 2. Vector similarity search (cosine via <=>)
	songs, err := s.repo.SearchSimilarSongs(
		ctx,
		embedding,
		req.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 3. Map to response DTOs
	results := make([]model.SongResult, 0, len(songs))
	for _, song := range songs {
		results = append(results, model.SongResult{
			SongID:     song.SongID,
			Name:       song.Name,
			Artists:    song.Artists,
			Album:      song.Album,
			Style:      song.Style,
			Mood:       song.Mood,
			Keywords:   song.Keywords,
			Popularity: song.Popularity,
		})
	}

	return &model.SearchResponse{
		Query:   req.Query,
		Results: results,
	}, nil
}

func parsePromptFilters(query string) (artist string, album string) {
	if m := artistDirectiveRe.FindStringSubmatch(query); len(m) == 2 {
		artist = cleanDirectiveValue(m[1], []string{
			"专辑:", "专辑：", "album:", "album：",
		})
	}
	if m := albumDirectiveRe.FindStringSubmatch(query); len(m) == 2 {
		album = cleanDirectiveValue(m[1], []string{
			"歌手:", "歌手：", "artist:", "artist：", "singer:", "singer：",
		})
	}
	return artist, album
}

func cleanDirectiveValue(raw string, stopTokens []string) string {
	v := strings.TrimSpace(raw)
	lower := strings.ToLower(v)

	cut := len(v)
	for _, tok := range stopTokens {
		if idx := strings.Index(lower, strings.ToLower(tok)); idx >= 0 && idx < cut {
			cut = idx
		}
	}
	v = strings.TrimSpace(v[:cut])
	return v
}
