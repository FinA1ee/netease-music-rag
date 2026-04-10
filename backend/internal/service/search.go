package service

import (
	"context"
	"fmt"

	"netease-music-rag/backend/internal/model"
	"netease-music-rag/backend/internal/repository"
)

// SearchService handles natural-language song recommendation via vector similarity.
type SearchService struct {
	llmClient *LLMClient
	repo      *repository.SongRepo
}

func NewSearchService(lc *LLMClient, repo *repository.SongRepo) *SearchService {
	return &SearchService{
		llmClient: lc,
		repo:      repo,
	}
}

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

	// 1. Embed the query
	embedding, err := s.llmClient.GetEmbedding(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// 2. Vector similarity search (cosine via <=>)
	songs, err := s.repo.SearchSimilarSongs(ctx, embedding, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 3. Map to response DTOs
	results := make([]model.SongResult, 0, len(songs))
	for _, song := range songs {
		results = append(results, model.SongResult{
			SongID:   song.SongID,
			Name:     song.Name,
			Artists:  song.Artists,
			Album:    song.Album,
			Style:    song.Style,
			Mood:     song.Mood,
			Keywords: song.Keywords,
		})
	}

	return &model.SearchResponse{
		Query:   req.Query,
		Results: results,
	}, nil
}
