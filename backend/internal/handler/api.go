package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"netease-music-rag/backend/internal/service"
)

type APIHandler struct {
	workflowService *service.WorkflowService
}

func NewAPIHandler(ws *service.WorkflowService) *APIHandler {
	return &APIHandler{workflowService: ws}
}

func (h *APIHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/search", h.Search)
	r.Post("/api/trigger-job", h.TriggerJob) // Manual trigger for testing
}

func (h *APIHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("l")
	limit := 5
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	if query == "" {
		http.Error(w, "query 'q' is required", http.StatusBadRequest)
		return
	}

	songs, err := h.workflowService.Search(r.Context(), query, limit)
	if err != nil {
		http.Error(w, "Search failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query": query,
		"songs": songs,
	})
}

func (h *APIHandler) TriggerJob(w http.ResponseWriter, r *http.Request) {
	go h.workflowService.RunDailyJob()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Job started in background",
	})
}
