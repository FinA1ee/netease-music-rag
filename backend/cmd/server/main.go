package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"netease-music-rag/backend/internal/config"
	"netease-music-rag/backend/internal/model"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Unable to connect to database via GORM: %v", err)
	}

	// Auto migrate matching the new struct format
	db.Exec("CREATE EXTENSION IF NOT EXISTS vector;")
	// Setup vector column explicitly since AutoMigrate doesn't fully understand vector types sometimes
	db.AutoMigrate(&model.Songs{})

	log.Println("Connected to PostgreSQL successfully via GORM.")

	// // Init Repo
	// repo := repository.NewSongRepo(db)

	// // Init Services
	// neteaseClient := service.NewNeteaseClient(cfg)
	// llmClient, err := service.NewLLMClient(cfg.GeminiAPIKey, cfg.LLMModel, cfg.EmbeddingModel)
	// if err != nil {
	// 	log.Fatalf("Gemini client init failed: %v", err)
	// }

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
