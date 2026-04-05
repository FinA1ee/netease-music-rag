package model

// 艺人信息（对应网易云 ar 字段）
type Artist struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// 专辑信息（对应网易云 al 字段）
type Album struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	PicUrl string `json:"picUrl"`
}

// 歌单信息
type Playlist struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	CoverImgUrl string   `json:"coverImgUrl"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`    // 曲风描述
	AlgTags     []string `json:"algTags"` // 算法标签
}

// 歌曲模型 Supabase
type Songs struct {
	ID         uint     `gorm:"primaryKey" json:"id"`
	SongID     int64    `gorm:"column:song_id;unique" json:"song_id"`
	Name       string   `gorm:"column:name" json:"name"`
	Lyric      string   `gorm:"column:lyric" json:"lyric"`
	Popularity float32  `gorm:"column:popularity" json:"popularity"`
	Duration   int64    `gorm:"column:duration" json:"duration"`                          // 时长 ms
	Artists    []Artist `gorm:"column:artists;type:jsonb;serializer:json" json:"artists"` // 👈 加了 serializer
	Album      Album    `gorm:"column:album;type:jsonb;serializer:json" json:"album"`     // 👈 加了 serializer

	// Playlist Data 日推歌单数据
	Playlist Playlist `gorm:"column:playlist;type:jsonb;serializer:json" json:"playlist"`

	// Tags 曲风 & 热度等数据
	Keywords []string `gorm:"column:keywords;type:text[]" json:"keywords"` // LLM 训练用
	Mood     []string `gorm:"column:mood;type:text[]" json:"mood"`         // LLM 生成
	Theme    []string `gorm:"column:theme;type:text[]" json:"theme"`       // LLM 生成

	// ✅ 修复 vector 空值报错
	Embedding string `gorm:"column:embedding;type:vector(1536);default:null" json:"-"`
}

type LLMAnalysisResult struct {
	Style       []string `json:"style"`
	Mood        []string `json:"mood"`
	Scene       []string `json:"scene"`
	Description string   `json:"description"`
}

type QueryResponse struct {
	Songs []Songs `json:"songs"`
}
