package main

import (
	"log"
	"net/http"

	"netease-music-rag/backend/internal/config"
	"netease-music-rag/backend/internal/handler"
	"netease-music-rag/backend/internal/model"
	"netease-music-rag/backend/internal/repository"
	"netease-music-rag/backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db := initDB(cfg)

	llmClient, err := service.NewLLMClient(cfg.GeminiAPIKey, cfg.LLMModel, cfg.EmbeddingModel)
	if err != nil {
		log.Fatalf("Gemini client init failed: %v", err)
	}

	neteaseClient := service.NewNeteaseClient(cfg)
	if neteaseClient == nil {
		log.Fatalf("Netease client init failed")
	}

	repo := repository.NewSongRepo(db)
	eventBus := service.NewEventBus()
	workflowSvc := service.NewWorkflowService(neteaseClient, llmClient, repo, cfg.NeteasePhone, eventBus)
	searchSvc := service.NewSearchService(llmClient, repo)

	// workflowSvc.StartCron()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	apiHandler := handler.NewAPIHandler(workflowSvc, searchSvc, neteaseClient, eventBus)
	apiHandler.RegisterRoutes(r)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
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
