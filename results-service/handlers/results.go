package handlers

import (
	"encoding/json"
	"net/http"
	"results-service/store"
)

// ResultsHandler handles HTTP requests for vote results
type ResultsHandler struct {
	store *store.Store
}

// NewResultsHandler creates a new ResultsHandler instance
func NewResultsHandler(s *store.Store) *ResultsHandler {
	return &ResultsHandler{store: s}
}

// ServeHTTP implements the http.Handler interface
func (h *ResultsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	counts, err := h.store.GetVoteCounts(r.Context())
	if err != nil {
		http.Error(w, "Failed to get results", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}
