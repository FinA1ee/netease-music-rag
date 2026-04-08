package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"

	"netease-music-rag/backend/internal/config"
	"netease-music-rag/backend/internal/dal"
	"netease-music-rag/backend/internal/model"
	"netease-music-rag/backend/internal/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Unable to connect to database via GORM: %v", err)
	}

	// 0. 初始化LLM

	// Init LLM Client
	llmClient, err := service.NewLLMClient(cfg.GeminiAPIKey, "gemini-2.5-flash-lite")
	if err != nil {
		log.Fatalf("Gemini client init failed: %v", err)
	}

	// 测试
	// llmClient.DryRun(context.Background())

	// 1. 连接数据库

	// Auto migrate matching the new struct format
	db.Exec("CREATE EXTENSION IF NOT EXISTS vector;")
	// Setup vector column explicitly since AutoMigrate doesn't fully understand vector types sometimes
	db.AutoMigrate(&model.Songs{})

	log.Println("Connected to PostgreSQL successfully via GORM.")

	// Init Repo
	// repo := repository.NewSongRepo(db)

	// 2. 网易云服务
	neteaseClient := service.NewNeteaseClient(cfg)
	if neteaseClient == nil {
		log.Fatalf("Netease client init failed")
	}

	if cfg.NeteasePhone != "" {
		if err := neteaseClient.Login(cfg.NeteasePhone); err != nil {
			log.Printf("Netease login failed: %v", err)
		}
	}

	// fetch all playlist
	recommendPlaylists, err := neteaseClient.GetDailyRecommendPlaylist()
	if err != nil || recommendPlaylists == nil {
		log.Fatalf("Netease get recommendSongs failed: %v", err)
	}

	// 🔥 最终歌曲列表（全局去重）
	finalSongList := make([]*model.NeteaseSongDTO, 0)
	// 🔥 去重使用：记录已经加入的歌曲 ID
	existSongID := make(map[int64]bool)

	// for each playlist, search for detail based on id
	for idx, playlist := range *recommendPlaylists {
		log.Printf("Processing Playlist %d", idx)
		playlistDetail, err := neteaseClient.GetDetailPlaylist(playlist.ID)

		if err != nil || playlistDetail == nil {
			log.Printf("Netease get playlistDetail %d failed: %v", playlist.ID, err)
			continue
		}

		songs := playlistDetail.Tracks
		if len(songs) == 0 {
			continue
		}

		// 随机打乱
		rand.Shuffle(len(songs), func(i, j int) {
			songs[i], songs[j] = songs[j], songs[i]
		})

		// 每个歌单最多取 1 首
		count := 0
		for _, song := range songs {
			if count >= 1 {
				break
			}

			// ✅ 去重：不存在才加入
			if !existSongID[song.ID] {

				// get song related playlist
				song.Playlist = model.DetailPlaylistData{
					ID:          playlistDetail.ID,
					Name:        playlistDetail.Name,
					CoverImgUrl: playlistDetail.CoverImgUrl,
					Description: playlistDetail.Description,
					Tags:        playlistDetail.Tags,
					// AlgTags:         playlistDetail.AlgTags,
					SubscribedCount: playlistDetail.SubscribedCount,
				}

				// get song lyrics
				lyric, err := neteaseClient.GetSongLyrics(song.ID)
				if err != nil || lyric == nil {
					log.Printf("Netease get songLyrics failed: %v", err)
					return
				}
				song.Lyric = *lyric

				// get song llm analysis
				llmAnalysis, err := llmClient.AnalyzeSong(context.Background(), &song, *lyric)
				if err != nil || llmAnalysis == nil {
					log.Printf("LLM analyze song failed: %v", err)
					return
				}

				// 关键词
				kw, _ := json.Marshal(llmAnalysis.Keywords)
				s, _ := json.Marshal(llmAnalysis.Style)
				m, _ := json.Marshal(llmAnalysis.Mood)
				t, _ := json.Marshal(llmAnalysis.Theme)
				featuresBytes, _ := json.Marshal(llmAnalysis.Features)

				song.LlmData = &model.NeteaseSongLLMAnalysis{
					Keywords: string(kw),
					Style:    string(s),
					Mood:     string(m),
					Theme:    string(t),
					Features: string(featuresBytes),
				}

				finalSongList = append(finalSongList, &song)
				existSongID[song.ID] = true
				count++
			}
		}
	}

	log.Printf("✅ 去重完成，最终有效歌曲数量：%d\n", len(finalSongList))

	// 将60首歌落库
	err = dal.SaveSongsToDB(db, finalSongList)
	if err != nil {
		log.Fatalf("Save songs to DB failed: %v", err)
	}

	// workflowSvc := service.NewWorkflowService(neteaseClient, llmClient, repo)

	// // Start cron job
	// workflowSvc.StartCron()

	// // Init Router
	// r := chi.NewRouter()
	// r.Use(middleware.Logger)
	// r.Use(cors.Handler(cors.Options{
	// 	AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8080", "*"},
	// 	AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowedHeaders: []string{"Accept", "Content-Type", "X-CSRF-Token"},
	// }))

	// apiHandler := handler.NewAPIHandler(workflowSvc)
	// apiHandler.RegisterRoutes(r)

	// log.Printf("Server starting on port %s...", cfg.Port)
	// if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
	// 	log.Fatalf("Server stopped: %v", err)
	// }
}
