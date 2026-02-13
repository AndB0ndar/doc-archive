package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/repository"
	"github.com/AndB0ndar/doc-archive/internal/vectorizer"
)

type SearchHandler struct {
	cfg            *config.Config
	docRepo        *repository.DocumentRepository
	chunkRepo      *repository.ChunkRepository
	embedderClient *vectorizer.Client
}

func NewSearchHandler(
	cfg *config.Config,
	docRepo *repository.DocumentRepository,
	chunkRepo *repository.ChunkRepository,
	embedderClient *vectorizer.Client,
) *SearchHandler {
	return &SearchHandler{
		cfg:            cfg,
		docRepo:        docRepo,
		chunkRepo:      chunkRepo,
		embedderClient: embedderClient,
	}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	searchType := strings.ToLower(r.URL.Query().Get("type"))
	limitStr := r.URL.Query().Get("limit")

	if query == "" {
		http.Error(w, "Missing search query (q)", http.StatusBadRequest)
		return
	}

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var results []repository.ChunkSearchResult
	var err error

	switch searchType {
	case "", "text":
		slog.Debug("full-text search by chunks", "query", query, "limit", limit)
		results, err = h.chunkRepo.FullTextSearchChunks(query, limit)

	case "vector", "semantic":
		embedding, err := h.embedderClient.Embed(query)
		if err != nil {
			slog.Error("failed to get embedding for query", "error", err)
			http.Error(w, "Search service unavailable", http.StatusServiceUnavailable)
			return
		}
		results, err = h.chunkRepo.SemanticSearchChunks(embedding, limit)

	default:
		http.Error(w, "Invalid search type. Use 'text' or 'vector'", http.StatusBadRequest)
		return
	}

	if err != nil {
		slog.Error("search failed", "type", searchType, "error", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(results); err != nil {
		slog.Error("failed to encode search results", "error", err)
	}
}
