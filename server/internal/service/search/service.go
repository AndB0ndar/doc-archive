package search

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/domain"
)

type Service struct {
	chunkRepo      domain.ChunkRepository
	embedderClient domain.EmbedderClient
	rerankerClient domain.RerankerClient
	readerClient   domain.ReaderClient
	cfg            config.SearchConfig
}

func New(
	chunkRepo domain.ChunkRepository,
	embedderClient domain.EmbedderClient,
	rerankerClient domain.RerankerClient,
	readerClient domain.ReaderClient,
	cfg config.SearchConfig,
) domain.SearchService {
	return &Service{
		chunkRepo:      chunkRepo,
		embedderClient: embedderClient,
		rerankerClient: rerankerClient,
		readerClient:   readerClient,
		cfg:            cfg,
	}
}

func (s *Service) Search(
	ctx context.Context,
	req domain.SearchQuery,
	UserID string,
	Limit int,
) ([]domain.SearchResult, error) {
	if err := s.validate(&req); err != nil {
		return nil, err
	}
	if Limit <= 0 {
		Limit = s.cfg.DefaultLimit
	}
	if Limit > s.cfg.MaxLimit {
		Limit = s.cfg.MaxLimit
	}
	fetchLimit := Limit * 3
	if fetchLimit > s.cfg.MaxLimit {
		fetchLimit = s.cfg.MaxLimit
	}

	var chunks []*domain.ChunkSearchResult
	var err error

	switch req.Type {
	case "", "text":
		chunks, err = s.chunkRepo.FullTextSearch(
			ctx, req.Query, UserID, fetchLimit,
		)
	case "vector", "semantic":
		embedResult, err := s.embedderClient.Embed(req.Query)
		if err != nil {
			return nil, fmt.Errorf("embedding failed: %w", err)
		}
		chunks, err = s.chunkRepo.SemanticSearch(
			ctx, embedResult.Embeddings, UserID, fetchLimit,
		)
	default:
		return nil, fmt.Errorf("unsupported search type: %s", req.Type)
	}
	if err != nil {
		return nil, err
	}

	results := s.toResults(chunks)

	if s.cfg.RerankerEnabled && len(results) > 0 {
		results = s.applyReranking(ctx, req.Query, results)
	}

	if len(results) > Limit*2 {
		results = results[:Limit*2]
	}

	if s.cfg.ReaderEnabled && len(results) > 0 {
		results = s.applyReader(ctx, req.Query, results)
	}

	if len(results) > Limit {
		results = results[:Limit]
	}

	return results, nil
}

func (s *Service) validate(req *domain.SearchQuery) error {
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		return fmt.Errorf("empty query")
	}
	return nil
}

func (s *Service) toResults(
	chunks []*domain.ChunkSearchResult,
) []domain.SearchResult {
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

func (s *Service) applyReranking(
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
		return results
	}
	if len(rerankResult.Scores) != len(results) {
		return results
	}

	for i := range results {
		score := float64(rerankResult.Scores[i]) // FIXME
		results[i].Confidence = &score
	}

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

func (s *Service) applyReader(
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
			continue
		}
		results[i].Answer = &ans.Answer
		results[i].Confidence = &ans.Confidence
	}
	return results
}
