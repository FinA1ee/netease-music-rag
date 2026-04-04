package model

// 艺人信息（对应网易云 ar 字段）
type Artist struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// 专辑信息（对应网易云 al 字段）
type Album struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// 歌曲模型（完整适配网易云 API + Supabase）
type Songs struct {
	ID            uint     `gorm:"primaryKey" json:"id"`
	SongID        int64    `gorm:"column:song_id;unique" json:"song_id"`
	Name          string   `gorm:"column:name" json:"name"`
	Duration      int64    `gorm:"column:duration" json:"duration"`          // 时长 ms
	Artists       []Artist `gorm:"column:artists;type:jsonb" json:"artists"` // 艺人数组
	Album         Album    `gorm:"column:album;type:jsonb" json:"album"`     // 专辑对象
	AlbumCoverURL string   `gorm:"column:album_cover_url" json:"album_cover_url"`
	SongTag       []string `gorm:"column:song_tag;type:text[]" json:"song_tag"` // 官方标签
	Lyric         string   `gorm:"column:lyric" json:"lyric"`
	Style         []string `gorm:"column:style;type:text[]" json:"style"`
	Mood          []string `gorm:"column:mood;type:text[]" json:"mood"`
	Description   string   `gorm:"column:description" json:"description"`
	Embedding     string   `gorm:"column:embedding;type:vector(1536)" json:"-"`
}

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
