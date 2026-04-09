package main

import (
	"log"

	"netease-music-rag/backend/internal/config"
	"netease-music-rag/backend/internal/model"
	"netease-music-rag/backend/internal/repository"
	"netease-music-rag/backend/internal/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db := initDB(cfg)

	llmClient, err := service.NewLLMClient(cfg.GeminiAPIKey, cfg.LLMModel)
	if err != nil {
		log.Fatalf("Gemini client init failed: %v", err)
	}

	neteaseClient := service.NewNeteaseClient(cfg)
	if neteaseClient == nil {
		log.Fatalf("Netease client init failed")
	}

	repo := repository.NewSongRepo(db)
	workflowSvc := service.NewWorkflowService(neteaseClient, llmClient, repo, cfg.NeteasePhone)

	if err := workflowSvc.RunDailyJob(); err != nil {
		log.Fatalf("Daily job failed: %v", err)
	}

	// TODO: uncomment when ready to run as a server
	// workflowSvc.StartCron()
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

func initDB(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Unable to connect to database via GORM: %v", err)
	}

	db.Exec("CREATE EXTENSION IF NOT EXISTS vector;")
	db.AutoMigrate(&model.Songs{})

	log.Println("Connected to PostgreSQL successfully via GORM.")
	return db
}
