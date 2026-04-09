package model

// SongFeatures holds structured per-song audio and cultural characteristics
// produced by LLM analysis. JSON keys match the prompt's required English field names.
type SongFeatures struct {
	Language     []string `json:"language"`     // e.g. ["普通话"]
	SingerNation []string `json:"singerNation"` // e.g. ["中国"]
	SingerGender []string `json:"singerGender"` // e.g. ["男"]
	SingerVoice  []string `json:"singerVoice"`  // e.g. ["温柔"]
	VocalType    []string `json:"vocalType"`    // e.g. ["独唱"]
	Instruments  []string `json:"instruments"`  // e.g. ["钢琴", "吉他"]
	Speed        []string `json:"speed"`        // e.g. ["中速"]
	Intensity    []string `json:"intensity"`    // e.g. ["温和"]
	Scene        []string `json:"scene"`        // e.g. ["睡前"]
	Arrangement  []string `json:"arrangement"`  // e.g. ["清淡"]
	EraStyle     []string `json:"eraStyle"`     // e.g. ["现代"]
	EmoColor     string   `json:"emoColor"`     // e.g. "暖色调" (single value per prompt spec)
}

// LLMAnalysisResult is the structured output returned by the LLM for a song.
type LLMAnalysisResult struct {
	Keywords []string     `json:"keywords"` // representative lyric keywords
	Style    []string     `json:"style"`    // music genre/style tags
	Mood     []string     `json:"mood"`     // emotional tags
	Theme    []string     `json:"theme"`    // lyric theme tags
	Features SongFeatures `json:"features"` // audio & cultural characteristics
}

// QueryResponse wraps the song search results returned by the API.
type QueryResponse struct {
	Songs []Songs `json:"songs"`
}
