package repository

import (
	"context"
	"fmt"
	"log"

	"netease-music-rag/backend/internal/model"

	"gorm.io/gorm"
)

type SongRepo struct {
	db *gorm.DB
}

func NewSongRepo(db *gorm.DB) *SongRepo {
	return &SongRepo{db: db}
}

func (r *SongRepo) SaveSong(ctx context.Context, song *model.Songs, embedding []float32) error {
	song.Embedding = float32SliceToString(embedding)
	// Upsert based on SongID
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO songs (
			song_id, name, duration, artists, album, album_cover_url, 
			song_tag, lyric, style, mood, description, embedding
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?::vector
		) ON CONFLICT (song_id) DO UPDATE SET
			lyric = EXCLUDED.lyric,
			style = EXCLUDED.style,
			mood = EXCLUDED.mood,
			description = EXCLUDED.description,
			embedding = EXCLUDED.embedding::vector;
	`, song.SongID, song.Name, song.Duration, song.Artists, song.Album,
		song.Keywords, song.Lyric, song.Theme, song.Mood, song.Embedding,
	).Error

	if err != nil {
		log.Printf("Failed to insert/update song %d: %v", song.SongID, err)
	}
	return err
}

func (r *SongRepo) HasSongLLMAnalyzed(ctx context.Context, id int64) bool {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Songs{}).Where("song_id = ? AND description != ''", id).Count(&count).Error
	return err == nil && count > 0
}

func (r *SongRepo) SearchSimilarSongs(ctx context.Context, embedding []float32, limit int) ([]model.Songs, error) {
	var songs []model.Songs
	embStr := float32SliceToString(embedding)
	err := r.db.WithContext(ctx).Order(fmt.Sprintf("embedding <-> '%s'", embStr)).Limit(limit).Find(&songs).Error
	return songs, err
}

func float32SliceToString(v []float32) string {
	str := "["
	for i, f := range v {
		if i > 0 {
			str += ","
		}
		str += fmt.Sprintf("%f", f)
	}
	str += "]"
	return str
}
