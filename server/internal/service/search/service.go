package search

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
)

type service struct {
	chunkRepo      domain.ChunkRepository
	embedderClient domain.EmbedderClient
	rerankerClient domain.RerankerClient
	readerClient   domain.ReaderClient
	cfg            config.SearchConfig
	logger         *logger.Logger
}

func New(
	chunkRepo domain.ChunkRepository,
	embedderClient domain.EmbedderClient,
	rerankerClient domain.RerankerClient,
	readerClient domain.ReaderClient,
	cfg config.SearchConfig,
	logger *logger.Logger,
) domain.SearchService {
	return &service{
		chunkRepo:      chunkRepo,
		embedderClient: embedderClient,
		rerankerClient: rerankerClient,
		readerClient:   readerClient,
		cfg:            cfg,
		logger:         logger,
	}
}

func (s *service) Search(
	ctx context.Context,
	req domain.SearchQuery,
	userID string,
	limit int,
) ([]domain.SearchResult, error) {
	// Log the incoming request
	s.logger.InfoContext(ctx, "search request started",
		slog.String("query", req.Query),
		slog.String("type", req.Type),
		slog.String("user_id", userID),
		slog.Int("requested_limit", limit),
	)

	// Validate input
	if err := s.validate(&req); err != nil {
		s.logger.WarnContext(ctx, "search validation failed", slog.String("error", err.Error()))
		return nil, err
	}

	// Apply limits
	if limit <= 0 {
		limit = s.cfg.DefaultLimit
		s.logger.DebugContext(ctx, "using default limit", slog.Int("limit", limit))
	}
	if limit > s.cfg.MaxLimit {
		limit = s.cfg.MaxLimit
		s.logger.DebugContext(ctx, "capped limit to max", slog.Int("limit", limit))
	}
	fetchLimit := limit * 3
	if fetchLimit > s.cfg.MaxLimit {
		fetchLimit = s.cfg.MaxLimit
	}
	s.logger.DebugContext(ctx, "computed fetch limit",
		slog.Int("limit", limit),
		slog.Int("fetch_limit", fetchLimit),
	)

	var chunks []*domain.ChunkSearchResult
	var err error

	// Perform search based on type
	switch req.Type {
	case "", "text":
		s.logger.InfoContext(ctx, "performing full-text search",
			slog.String("query", req.Query),
			slog.String("user_id", userID),
			slog.Int("fetch_limit", fetchLimit),
		)
		chunks, err = s.chunkRepo.FullTextSearch(ctx, req.Query, userID, fetchLimit)
		if err != nil {
			s.logger.ErrorContext(ctx, "full-text search failed", slog.String("error", err.Error()))
			return nil, fmt.Errorf("full-text search failed: %w", err)
		}
	case "vector", "semantic":
		s.logger.InfoContext(ctx, "performing semantic search",
			slog.String("query", req.Query),
			slog.String("user_id", userID),
			slog.Int("fetch_limit", fetchLimit),
		)
		embedResult, err := s.embedderClient.Embed(req.Query)
		if err != nil {
			s.logger.ErrorContext(ctx, "embedding failed", slog.String("error", err.Error()))
			return nil, fmt.Errorf("embedding failed: %w", err)
		}
		chunks, err = s.chunkRepo.SemanticSearch(ctx, embedResult.Embeddings, userID, fetchLimit)
		if err != nil {
			s.logger.ErrorContext(ctx, "semantic search failed", slog.String("error", err.Error()))
			return nil, fmt.Errorf("semantic search failed: %w", err)
		}
	default:
		errMsg := fmt.Sprintf("unsupported search type: %s", req.Type)
		s.logger.WarnContext(ctx, errMsg)
		return nil, fmt.Errorf("unsupported search type: %s", req.Type)
	}

	s.logger.InfoContext(ctx, "retrieved chunks", slog.Int("count", len(chunks)))
	if len(chunks) == 0 {
		s.logger.InfoContext(ctx, "no chunks found, returning empty result")
		return []domain.SearchResult{}, nil
	}

	// Convert to results
	results := s.toResults(chunks)
	s.logger.DebugContext(ctx, "converted chunks to results", slog.Int("count", len(results)))

	// Apply reranking if enabled
	if s.cfg.RerankerEnabled && len(results) > 0 {
		s.logger.InfoContext(ctx, "applying reranking", slog.Int("input_count", len(results)))
		beforeCount := len(results)
		results = s.applyReranking(ctx, req.Query, results)
		s.logger.InfoContext(ctx, "reranking applied",
			slog.Int("input_count", beforeCount),
			slog.Int("output_count", len(results)),
		)
	}

	// Trim to 2x limit before reader
	if len(results) > limit*2 {
		results = results[:limit*2]
		s.logger.DebugContext(ctx, "trimmed results before reader", slog.Int("trimmed_count", len(results)))
	}

	// Apply reader if enabled
	if s.cfg.ReaderEnabled && len(results) > 0 {
		s.logger.InfoContext(ctx, "applying reader", slog.Int("input_count", len(results)))
		beforeCount := len(results)
		results = s.applyReader(ctx, req.Query, results)
		s.logger.InfoContext(ctx, "reader applied",
			slog.Int("input_count", beforeCount),
			slog.Int("output_count", len(results)),
			slog.Int("results_with_answer", countWithAnswer(results)),
		)
	}

	// Final trim to limit
	if len(results) > limit {
		results = results[:limit]
	}

	s.logger.InfoContext(ctx, "search completed successfully",
		slog.String("query", req.Query),
		slog.Int("final_results", len(results)),
	)

	return results, nil
}

func (s *service) validate(req *domain.SearchQuery) error {
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		return fmt.Errorf("empty query")
	}
	return nil
}

func (s *service) toResults(chunks []*domain.ChunkSearchResult) []domain.SearchResult {
	results := make([]domain.SearchResult, len(chunks))
	for i, ch := range chunks {
		results[i] = domain.SearchResult{
			ChunkID:    ch.ID,
			DocumentID: ch.DocumentID,
			Content:    ch.Content,
			Confidence: &ch.Similarity,
		}
	}
	return results
}

func (s *service) applyReranking(
	ctx context.Context,
	query string,
	results []domain.SearchResult,
) []domain.SearchResult {
	if len(results) == 0 {
		return results
	}

	texts := make([]string, len(results))
	for i, r := range results {
		texts[i] = r.Content
	}

	rerankResult, err := s.rerankerClient.Rerank(query, texts)
	if err != nil {
		s.logger.ErrorContext(ctx, "reranking failed, skipping",
			slog.String("error", err.Error()),
			slog.Int("result_count", len(results)),
		)
		return results
	}
	if len(rerankResult.Scores) != len(results) {
		s.logger.WarnContext(ctx, "reranking returned mismatched scores, skipping",
			slog.Int("expected", len(results)),
			slog.Int("got", len(rerankResult.Scores)),
		)
		return results
	}

	// Update confidence scores
	for i := range results {
		score := float64(rerankResult.Scores[i]) // FIXME: ensure conversion is correct
		results[i].Confidence = &score
	}

	// Re-sort by new confidence
	sort.Slice(results, func(i, j int) bool {
		ci := 0.0
		if results[i].Confidence != nil {
			ci = *results[i].Confidence
		}
		cj := 0.0
		if results[j].Confidence != nil {
			cj = *results[j].Confidence
		}
		return ci > cj
	})

	return results
}

func (s *service) applyReader(
	ctx context.Context,
	query string,
	results []domain.SearchResult,
) []domain.SearchResult {
	topK := 3
	if topK > len(results) {
		topK = len(results)
	}
	for i := 0; i < topK; i++ {
		ans, err := s.readerClient.Answer(query, results[i].Content)
		if err != nil {
			s.logger.WarnContext(ctx, "reader failed for chunk",
				slog.String("chunk_id", results[i].ChunkID),
				slog.String("document_id", results[i].DocumentID),
				slog.String("error", err.Error()),
			)
			continue
		}
		results[i].Answer = &ans.Answer
		results[i].Confidence = &ans.Confidence
		s.logger.DebugContext(ctx, "reader succeeded for chunk",
			slog.String("chunk_id", results[i].ChunkID),
			slog.String("document_id", results[i].DocumentID),
			slog.Float64("confidence", ans.Confidence),
		)
	}
	return results
}

// Helper function to count results that have an Answer
func countWithAnswer(results []domain.SearchResult) int {
	count := 0
	for _, r := range results {
		if r.Answer != nil {
			count++
		}
	}
	return count
}
