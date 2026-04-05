package dal

import (
	"netease-music-rag/backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SaveSongsToDB 批量保存 NeteaseSongDTO 到 songs 表
// 自动去重、存在则更新、不存在则插入
func SaveSongsToDB(db *gorm.DB, songList []*model.NeteaseSongDTO) error {
	if len(songList) == 0 {
		return nil
	}

	var songsBatch []*model.Songs

	for _, dto := range songList {
		// 1. 转换 Artists
		var artists []model.Artist
		for _, ar := range dto.Ar {
			artists = append(artists, model.Artist{
				ID:   ar.ID,
				Name: ar.Name,
			})
		}

		// 2. 转换 Album
		album := model.Album{
			ID:     dto.Al.ID,
			Name:   dto.Al.Name,
			PicUrl: dto.Al.PicUrl,
		}

		// 3. 转换 Playlist
		playlist := model.Playlist{
			ID:          dto.Playlist.ID,
			Name:        dto.Playlist.Name,
			CoverImgUrl: dto.Playlist.CoverImgUrl,
			Description: dto.Playlist.Description,
			Tags:        dto.Playlist.Tags,
			AlgTags:     dto.Playlist.AlgTags,
		}

		// 3. 构建数据库模型
		song := &model.Songs{
			SongID:     dto.ID,
			Name:       dto.Name,
			Lyric:      dto.Lyric,
			Popularity: dto.Pop,
			Duration:   dto.Dt,
			Artists:    artists,
			Album:      album,
			Playlist:   playlist,
			Keywords:   []string{},
			Mood:       []string{},
			Theme:      []string{},
			// Embedding:     "",
		}

		songsBatch = append(songsBatch, song)
	}

	// 批量插入 + 主键冲突更新
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"song_id", "name", "artists", "album", "duration"}),
	}).CreateInBatches(songsBatch, 50).Error
}
