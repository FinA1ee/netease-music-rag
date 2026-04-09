package repository

import (
	"context"
	"fmt"
	"log"

	"netease-music-rag/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SongRepo struct {
	db *gorm.DB
}

func NewSongRepo(db *gorm.DB) *SongRepo {
	return &SongRepo{db: db}
}

// SaveSongs batch-upserts a list of NeteaseSongDTOs into the songs table.
func (r *SongRepo) SaveSongs(songList []*model.NeteaseSongDTO) error {
	if len(songList) == 0 {
		return nil
	}

	var batch []*model.Songs

	for _, dto := range songList {
		var artists []model.Artist
		for _, ar := range dto.Ar {
			artists = append(artists, model.Artist{ID: ar.ID, Name: ar.Name})
		}

		album := model.Album{
			ID:     dto.Al.ID,
			Name:   dto.Al.Name,
			PicUrl: dto.Al.PicUrl,
		}

		playlist := model.Playlist{
			ID:          dto.Playlist.ID,
			Name:        dto.Playlist.Name,
			CoverImgUrl: dto.Playlist.CoverImgUrl,
			Description: dto.Playlist.Description,
			Tags:        dto.Playlist.Tags,
		}

		batch = append(batch, &model.Songs{
			SongID:     dto.ID,
			Name:       dto.Name,
			Lyric:      dto.Lyric,
			Popularity: dto.Pop,
			Duration:   dto.Dt,
			Artists:    artists,
			Album:      album,
			Playlist:   playlist,
			Style:      dto.LlmData.Style,
			Keywords:   dto.LlmData.Keywords,
			Mood:       dto.LlmData.Mood,
			Theme:      dto.LlmData.Theme,
			Features:   dto.LlmData.Features,
		})
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "song_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "artists", "album", "duration", "lyric", "style", "keywords", "mood", "theme", "features"}),
	}).CreateInBatches(batch, 50).Error
}

// SaveSong upserts a single song with its vector embedding.
func (r *SongRepo) SaveSong(ctx context.Context, song *model.Songs, embedding []float32) error {
	song.Embedding = float32SliceToString(embedding)
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

// HasSongLLMAnalyzed returns true if the song has already been analyzed by the LLM.
func (r *SongRepo) HasSongLLMAnalyzed(ctx context.Context, id int64) bool {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Songs{}).Where("song_id = ? AND description != ''", id).Count(&count).Error
	return err == nil && count > 0
}

// SearchSimilarSongs returns songs ordered by vector similarity to the given embedding.
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
