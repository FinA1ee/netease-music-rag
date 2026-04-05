package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"netease-music-rag/backend/internal/model"

	"google.golang.org/genai"
)

type LLMClient struct {
	client         *genai.Client
	llmModel       string
	embeddingModel string
}

func NewLLMClient(apiKey, llmModel, embeddingModel string) (*LLMClient, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	return &LLMClient{
		client:         client,
		llmModel:       llmModel,
		embeddingModel: embeddingModel,
	}, nil
}

func (l *LLMClient) AnalyzeSong(ctx context.Context, songName, artist, lyrics string) (*model.LLMAnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze the following song and provide its musical style, mood, scene, and a poetic description text.
Respond ONLY in valid JSON format. Example: {"style":["Pop","Rock"], "mood":["Happy"], "scene":["Driving"], "description":"A wonderful song"}
Song Name: %s
Artist: %s
Lyrics:
%s
`, songName, artist, lyrics)

	result, err := l.client.Models.GenerateContent(ctx, l.llmModel, []*genai.Content{
		{Parts: []*genai.Part{{Text: prompt}}},
	}, nil)

	if err != nil {
		return nil, err
	}

	text := extractTextFromResult(result)
	text = cleanJSONWrapper(text)

	var analysis model.LLMAnalysisResult
	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		log.Printf("Failed to unmarshal JSON: %s", text)
		return nil, err
	}

	return &analysis, nil
}

// func (l *LLMClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
// 	resp, err := l.client.Models.EmbedContent(ctx, l.embeddingModel, &genai.EmbedContentRequest{
// 		Contents: []*genai.Content{
// 			{Parts: []*genai.Part{{Text: text}}},
// 		},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(resp.Embeddings) == 0 {
// 		return nil, fmt.Errorf("no embeddings returned")
// 	}
// 	return resp.Embeddings[0].Values, nil
// }

func extractTextFromResult(result *genai.GenerateContentResponse) string {
	if result != nil && len(result.Candidates) > 0 {
		cand := result.Candidates[0]
		if cand.Content != nil && len(cand.Content.Parts) > 0 {
			return cand.Content.Parts[0].Text
		}
	}
	return ""
}

func cleanJSONWrapper(js string) string {
	re := regexp.MustCompile(`(?s)^\s*` + "```" + `(?:json)?\s*(.*?)\s*` + "```" + `\s*$`)
	if match := re.FindStringSubmatch(js); len(match) == 2 {
		return match[1]
	}
	return js
}
