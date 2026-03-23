package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/middleware"
	"github.com/AndB0ndar/doc-archive/internal/models"
)

type SearchHandler struct {
	searchService domain.SearchService
	logger        *slog.Logger
}

func NewSearchHandler(
	searchService domain.SearchService, logger *slog.Logger,
) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
		logger:        logger,
	}
}

// Search выполняет поиск документов/чанков.
// @Summary      Поиск документов
// @Description  Полнотекстовый или семантический поиск по содержимому.
// @Tags         search
// @Produce      json
// @Param        q query string true "Поисковый запрос"
// @Param        type query string false "Тип поиска: text (по умолчанию) или vector/semantic"
// @Param        limit query int false "Максимальное количество результатов (макс 100)"
// @Success      200  {object}  models.SearchResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Security     BearerAuth
// @Router       /search [get]
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := h.logger.With(
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	logger.Debug("Search request started")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		logger.Warn("Search: unauthorized - userID missing in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	logger = logger.With(slog.String("user_id", userID))

	query := r.URL.Query().Get("q")
	if query == "" {
		logger.Warn("Search: missing search query (q)")
		http.Error(w, "Missing search query (q)", http.StatusBadRequest)
		return
	}
	logger = logger.With(slog.String("query", query))

	searchType := r.URL.Query().Get("type")
	if searchType == "" {
		searchType = "text"
		logger.Debug(
			"Search: using default search type",
			slog.String("type", searchType),
		)
	}
	if searchType != "text" && searchType != "vector" && searchType != "semantic" {
		logger.Warn(
			"Search: invalid search type",
			slog.String("provided_type", searchType),
		)
		http.Error(
			w, "Invalid search type. Use 'text', 'vector', or 'semantic'",
			http.StatusBadRequest,
		)
		return
	}
	logger = logger.With(slog.String("search_type", searchType))

	limit := 10 // default
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 100 {
				l = 100
				logger.Debug(
					"Search: limit capped to 100",
					slog.Int("original", l),
					slog.Int("capped", 100),
				)
			}
			limit = l
		} else {
			logger.Warn(
				"Search: invalid limit parameter, using default",
				slog.String("limit_str", limitStr),
				slog.Any("error", err),
			)
		}
	}
	logger = logger.With(slog.Int("limit", limit))
	logger.Debug("Search: parameters validated")

	searchQuery := domain.SearchQuery{
		Query: query,
		Type:  searchType,
	}

	logger.Debug("Search: calling searchService.Search")
	results, err := h.searchService.Search(
		r.Context(), searchQuery, userID, limit,
	)
	if err != nil {
		logger.Error("Search: search failed", slog.Any("error", err))
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}
	logger.Debug(
		"Search: search completed", slog.Int("results_count", len(results)),
	)

	// mapping domain.SearchResult -> models.SearchResultItem
	items := make([]models.SearchResultItem, len(results))
	for i, r := range results {
		items[i] = models.SearchResultItem{
			ChunkID:    r.ChunkID,
			DocumentID: r.DocumentID,
			Content:    r.Content,
			Answer:     r.Answer,
			Confidence: r.Confidence,
		}
	}
	logger.Debug(
		"Search: transformed results", slog.Int("items_count", len(items)),
	)

	resp := models.SearchResponse{Results: items}
	w.Header().Set("Content-Type", "application/json")
	logger.Debug("Search: encoding response")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error(
			"Search: failed to encode search results", slog.Any("error", err),
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logger.Info(
		"Search: completed", slog.Duration("duration", time.Since(start)),
	)
}
