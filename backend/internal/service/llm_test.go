package service

import (
	"testing"

	"google.golang.org/genai"
)

func TestCleanJSONWrapper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "json fenced block",
			in:   "```json\n{\"a\":1}\n```",
			want: "{\"a\":1}",
		},
		{
			name: "plain fenced block",
			in:   "```\n{\"a\":1}\n```",
			want: "{\"a\":1}",
		},
		{
			name: "no fence",
			in:   "{\"a\":1}",
			want: "{\"a\":1}",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := cleanJSONWrapper(tc.in)
			if got != tc.want {
				t.Fatalf("cleanJSONWrapper(%q) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestExtractTextFromResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result *genai.GenerateContentResponse
		want   string
	}{
		{
			name: "valid candidate text",
			result: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{{Text: "hello"}},
						},
					},
				},
			},
			want: "hello",
		},
		{name: "nil result", result: nil, want: ""},
		{name: "empty candidates", result: &genai.GenerateContentResponse{}, want: ""},
		{
			name: "nil content",
			result: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{{Content: nil}},
			},
			want: "",
		},
		{
			name: "empty parts",
			result: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{Content: &genai.Content{Parts: nil}},
				},
			},
			want: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := extractTextFromResult(tc.result)
			if got != tc.want {
				t.Fatalf("extractTextFromResult() = %q; want %q", got, tc.want)
			}
		})
	}
}
