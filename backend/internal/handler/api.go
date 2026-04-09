package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"netease-music-rag/backend/internal/model"
	"netease-music-rag/backend/internal/service"
)

type APIHandler struct {
	workflowService *service.WorkflowService
	searchService   *service.SearchService
}

func NewAPIHandler(ws *service.WorkflowService, ss *service.SearchService) *APIHandler {
	return &APIHandler{
		workflowService: ws,
		searchService:   ss,
	}
}

func (h *APIHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/search", h.Search)
	r.Post("/api/trigger-job", h.TriggerJob)
	r.Post("/api/trigger-embedding", h.TriggerEmbedding)
}

// Search godoc
// GET /api/search?q=我想听快节奏男声&l=5
//
// Embeds the query text, runs a cosine similarity search against the pgvector
// embedding column, and returns the top-l matching songs.
func (h *APIHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("l")

	if query == "" {
		writeError(w, http.StatusBadRequest, "query param 'q' is required")
		return
	}

	limit := 5
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	resp, err := h.searchService.Search(r.Context(), model.SearchRequest{
		Query: query,
		Limit: limit,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// TriggerJob godoc
// POST /api/trigger-job
//
// Manually kicks off the daily playlist fetch + LLM analysis job in the background.
func (h *APIHandler) TriggerJob(w http.ResponseWriter, r *http.Request) {
	go h.workflowService.RunDailyJob() //nolint:errcheck
	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "daily job started in background",
	})
}

// TriggerEmbedding godoc
// POST /api/trigger-embedding
//
// Manually kicks off the embedding backfill job in the background.
func (h *APIHandler) TriggerEmbedding(w http.ResponseWriter, r *http.Request) {
	go h.workflowService.RunEmbeddingJob(context.Background()) //nolint:errcheck
	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "embedding job started in background",
	})
}

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
