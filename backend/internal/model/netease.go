package model

// RecommendPlaylistData 每日推荐歌单 完整DTO
type RecommendPlaylistData struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	PicUrl string `json:"picUrl"`
}

// DetailPlaylistData 网易云 - 歌单详情模型（含曲目列表，仅用于 API 响应解析）
type DetailPlaylistData struct {
	ID              int64            `json:"id"`
	Name            string           `json:"name"`
	CoverImgUrl     string           `json:"coverImgUrl"`
	Description     string           `json:"description"`
	Tags            []string         `json:"tags"`            // 曲风描述
	SubscribedCount int64            `json:"subscribedCount"` // 订阅人数
	Tracks          []NeteaseSongDTO `json:"tracks"`          // 歌曲列表（解析时使用，不存入 DB）
}

// NeteaseSongLLMAnalysis 存储 LLM 分析结果（JSON 序列化后的字符串，便于写入 DB jsonb 列）
type NeteaseSongLLMAnalysis struct {
	Keywords string `json:"keywords"` // LLM 训练用
	Style    string `json:"style"`    // LLM 生成
	Mood     string `json:"mood"`     // LLM 生成
	Theme    string `json:"theme"`    // LLM 生成
	Features string `json:"features"` // LLM 生成
}

// NeteaseSongDTO 网易云 - 歌曲模型（API DTO，使用共享的 Artist / Album / Playlist 类型）
type NeteaseSongDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Dt   int64  `json:"dt"` // duration (ms)
	Ar   []Artist `json:"ar"` // 艺人列表，复用 types.go Artist
	Al   Album    `json:"al"` // 专辑信息，复用 types.go Album

	Lyric    string   `json:"lyric"`    // 歌词
	Pop      float32  `json:"pop"`      // 热度
	Playlist Playlist `json:"playlist"` // 归属歌单元数据（不含 Tracks，避免循环引用）

	LlmData *NeteaseSongLLMAnalysis `json:"llm_data"` // LLM 分析结果
}
