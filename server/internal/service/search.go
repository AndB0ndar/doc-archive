package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/models"
	"github.com/AndB0ndar/doc-archive/internal/repository"
)

type SearchService struct {
	cfg            *config.Config
	chunkRepo      *repository.ChunkRepository
	embedderClient *Embedder
	rerankerClient *Reranker
	readerClient   *Reader
}

func NewSearchService(
	cfg *config.Config,
	repo *repository.ChunkRepository,
	embedderClient *Embedder,
	rerankerClient *Reranker,
	readerClient *Reader,
) *SearchService {
	return &SearchService{
		cfg:            cfg,
		chunkRepo:      repo,
		embedderClient: embedderClient,
		rerankerClient: rerankerClient,
		readerClient:   readerClient,
	}
}

type SearchRequest struct {
	Query  string
	Type   string
	UserID int
	Limit  int
}

func (r *SearchRequest) Validate(defaultLimit, maxLimit int) error {
	r.Query = strings.TrimSpace(r.Query)
	if r.Query == "" {
		return ErrEmptyQuery
	}
	r.Type = strings.ToLower(r.Type)
	if r.Type != "" && r.Type != "text" && r.Type != "vector" && r.Type != "semantic" {
		return ErrInvalidType
	}
	if r.Type == "" {
		r.Type = "text"
	}
	if r.Limit <= 0 {
		r.Limit = defaultLimit
	}
	if r.Limit > maxLimit {
		r.Limit = maxLimit
	}
	return nil
}

var (
	ErrEmptyQuery  = fmt.Errorf("empty query")
	ErrInvalidType = fmt.Errorf("invalid search type, use 'text' or 'semantic'")
	ErrEmbedding   = fmt.Errorf("failed to get embedding")
)

func (s *SearchService) Search(
	req SearchRequest,
) ([]models.ChunkSearchResponse, error) {
	if err := req.Validate(s.cfg.SearchDefaultLimit, s.cfg.SearchMaxLimit); err != nil {
		return nil, err
	}

	var results []models.ChunkSearchResponse
	var err error

	switch req.Type {
	case "", "text":
		results, err = s.chunkRepo.FullTextSearchChunks(req.Query, req.UserID, req.Limit*3)
	case "vector", "semantic":
		embedding, err := s.embedderClient.Embed(req.Query)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrEmbedding, err)
		}
		results, err = s.chunkRepo.SemanticSearchChunks(embedding, req.UserID, req.Limit*3)
	default:
		return nil, ErrInvalidType
	}
	if err != nil {
		return results, err
	}

	if s.cfg.RerankerEnabled && len(results) > 0 {
		results = s.applyReranking(req, results)
	} else {
		if len(results) > req.Limit*2 {
			results = results[:req.Limit*2]
		}
	}

	if s.cfg.ReaderEnabled && len(results) > 0 {
		results = s.applyReader(req, results)
	} else {
		if len(results) > req.Limit {
			results = results[:req.Limit]
		}
	}

	return results, err
}

func (s *SearchService) applyReranking(
	req SearchRequest,
	results []models.ChunkSearchResponse,
) []models.ChunkSearchResponse {
	if len(results) == 0 {
		return results
	}

	texts := make([]string, len(results))
	for i, r := range results {
		texts[i] = r.Content
	}

	scores, err := s.rerankerClient.Rerank(req.Query, texts)
	if err != nil {
		if len(results) > req.Limit*2 {
			return results[:req.Limit*2]
		}
		return results
	}

	if len(scores) != len(results) {
		if len(results) > req.Limit*2 {
			return results[:req.Limit*2]
		}
		return results
	}

	for i := range results {
		results[i].Similarity = float64(scores[i])
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if len(results) > req.Limit*2 {
		results = results[:req.Limit*2]
	}
	return results
}

func (s *SearchService) applyReader(
	req SearchRequest,
	results []models.ChunkSearchResponse,
) []models.ChunkSearchResponse {
	readerTopK := 3
	if readerTopK > len(results) {
		readerTopK = len(results)
	}
	for i := 0; i < readerTopK; i++ {
		ans, err := s.readerClient.ExtractAnswer(req.Query, results[i].Content)
		if err != nil {
			continue
		}
		results[i].Answer = &ans.Answer
		results[i].AnswerConfidence = &ans.Confidence
	}
	if len(results) > req.Limit {
		return results[:req.Limit]
	}
	return results
}
