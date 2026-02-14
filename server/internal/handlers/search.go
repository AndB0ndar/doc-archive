package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AndB0ndar/doc-archive/internal/middleware"
	"github.com/AndB0ndar/doc-archive/internal/service"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req := service.SearchRequest{
		Query:  r.URL.Query().Get("q"),
		Type:   r.URL.Query().Get("type"),
		UserID: userID,
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = l
		}
	}

	results, err := h.searchService.Search(req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		slog.Error("failed to encode search results", "error", err)
	}
}

func (h *SearchHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrEmptyQuery):
		http.Error(w, "Missing search query (q)", http.StatusBadRequest)
	case errors.Is(err, service.ErrInvalidType):
		http.Error(
			w,
			"Invalid search type. Use 'text' or 'semantic'",
			http.StatusBadRequest,
		)
	case errors.Is(err, service.ErrEmbedding):
		slog.Error("embedding failed", "error", err)
		http.Error(
			w,
			"Search service unavailable",
			http.StatusServiceUnavailable,
		)
	default:
		slog.Error("search failed", "error", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
	}
}
