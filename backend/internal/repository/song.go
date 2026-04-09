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

// GetExistingSongIDs returns the subset of the given song IDs that already exist in the DB.
// Use this to skip duplicate songs before fetching lyrics or running LLM analysis.
func (r *SongRepo) GetExistingSongIDs(ctx context.Context, ids []int64) (map[int64]bool, error) {
	if len(ids) == 0 {
		return map[int64]bool{}, nil
	}
	var existing []int64
	err := r.db.WithContext(ctx).
		Model(&model.Songs{}).
		Where("song_id IN ?", ids).
		Pluck("song_id", &existing).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]bool, len(existing))
	for _, id := range existing {
		result[id] = true
	}
	return result, nil
}

// GetSongsNeedingEmbedding returns up to `limit` songs that have LLM analysis
// but no embedding yet. Used by the embedding backfill job.
func (r *SongRepo) GetSongsNeedingEmbedding(ctx context.Context, limit int) ([]model.Songs, error) {
	var songs []model.Songs
	err := r.db.WithContext(ctx).
		Where("style IS NOT NULL AND (embedding IS NULL)").
		Limit(limit).
		Find(&songs).Error
	return songs, err
}

// UpdateEmbedding writes a vector embedding for the given song_id.
// Uses a raw Exec because GORM does not natively support the pgvector ::vector cast.
func (r *SongRepo) UpdateEmbedding(ctx context.Context, songID int64, embedding []float32) error {
	embStr := float32SliceToString(embedding)
	err := r.db.WithContext(ctx).Exec(
		`UPDATE songs SET embedding = ?::vector WHERE song_id = ?`,
		embStr, songID,
	).Error
	if err != nil {
		log.Printf("UpdateEmbedding failed for song %d: %v", songID, err)
	}
	return err
}

// HasSongLLMAnalyzed returns true if the song has already been analyzed by the LLM.
func (r *SongRepo) HasSongLLMAnalyzed(ctx context.Context, id int64) bool {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Songs{}).Where("song_id = ? AND description != ''", id).Count(&count).Error
	return err == nil && count > 0
}

// SearchSimilarSongs returns songs ranked by cosine similarity (<=> operator in pgvector)
// to the given query embedding. Songs without embeddings are excluded.
func (r *SongRepo) SearchSimilarSongs(ctx context.Context, embedding []float32, limit int) ([]model.Songs, error) {
	var songs []model.Songs
	embStr := float32SliceToString(embedding)
	err := r.db.WithContext(ctx).
		Where("embedding IS NOT NULL").
		Order(fmt.Sprintf("embedding <=> '%s'::vector", embStr)).
		Limit(limit).
		Find(&songs).Error
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
