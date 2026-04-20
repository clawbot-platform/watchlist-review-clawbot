package retrievalgateway

import (
	"encoding/json"
	"net/http"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/retrieval"
)

type Handler struct {
	searcher Searcher
}

func NewHandler(searcher Searcher) *Handler {
	return &Handler{searcher: searcher}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("/v1/search", h.search)
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	if h == nil || h.searcher == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "retrieval searcher is not configured"})
		return
	}

	var req retrieval.Query
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	resp, err := h.searcher.Search(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
