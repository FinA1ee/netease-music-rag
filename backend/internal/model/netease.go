package model

// DailyRecommendPlaylist 每日推荐歌单 完整DTO
type RecommendPlaylistData struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	PicUrl string `json:"picUrl"`
}

// 网易云 - 歌单详情模型
type DetailPlaylistData struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	CoverImgUrl string   `json:"coverImgUrl"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"` // 曲风描述
	// AlgTags         []string         `json:"algTags"`         // 算法标签
	SubscribedCount int64            `json:"subscribedCount"` // 订阅人数
	Tracks          []NeteaseSongDTO `json:"tracks"`          // 歌曲
}

type NeteaseSongLLMAnalysis struct {
	Keywords any `json:"keywords"` // LLM 训练用
	Style    any `json:"style"`    // LLM 生成
	Mood     any `json:"mood"`     // LLM 生成
	Theme    any `json:"theme"`    // LLM 生成
	Features any `json:"features"` // LLM 生成
}

// 网易云 - 歌曲模型
type NeteaseSongDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Dt   int64  `json:"dt"` // duration
	Ar   []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"ar"`
	Al struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		PicUrl string `json:"picUrl"`
	} `json:"al"`

	Lyric    string             `json:"lyric"`    // 歌词
	Pop      float32            `json:"pop"`      // 热度
	Playlist DetailPlaylistData `json:"playlist"` // 歌单映射

	LlmData *NeteaseSongLLMAnalysis `json:"llm_data"` // LLM 分析结果
}
