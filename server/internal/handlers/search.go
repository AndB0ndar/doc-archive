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

// Search выполняет поиск документов/чанков.
// @Summary      Поиск документов
// @Description  Полнотекстовый или семантический поиск по содержимому.
// @Tags         search
// @Produce      json
// @Param        q query string true "Поисковый запрос"
// @Param        type query string false "Тип поиска: text (по умолчанию) или vector"
// @Param        limit query int false "Максимальное количество результатов (макс 100)"
// @Success      200  {array}   models.ChunkSearchResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Security     BearerAuth
// @Router       /search [get]
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
