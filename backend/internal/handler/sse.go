package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"netease-music-rag/backend/internal/service"
)

// StreamEvents godoc
// GET /api/jobs/stream
//
// Opens a Server-Sent Events stream. The client receives events published by
// the daily job and embedding job as they execute.
func (h *APIHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {

	// SSE headers — disable buffering at every proxy layer
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")        // disable nginx/proxy buffering
	w.Header().Set("Transfer-Encoding", "chunked")    // force chunked streaming

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported by server", http.StatusInternalServerError)
		return
	}

	// Subscribe and ensure cleanup when the client disconnects
	ch := h.eventBus.Subscribe()
	defer h.eventBus.Unsubscribe(ch)

	// Send a "connected" heartbeat so the browser knows we're live
	fmt.Fprintf(w, "data: %s\n\n", mustJSON(service.JobEvent{Type: "connected"}))
	flusher.Flush()

	// Replay recently missed events to this new subscriber
	for _, ev := range h.eventBus.RecentEvents() {
		fmt.Fprintf(w, "data: %s\n\n", mustJSON(ev))
	}
	flusher.Flush()

	// Heartbeat every 15s — keeps webpack-dev-server proxy and nginx from
	// closing idle SSE connections. SSE comment lines are invisible to EventSource.
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", mustJSON(ev))
			flusher.Flush()

		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			// Client disconnected — stop the goroutine
			return
		}
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
