package service

import (
	"fmt"
	"strings"

	"github.com/AndB0ndar/doc-archive/internal/config"
	"github.com/AndB0ndar/doc-archive/internal/models"
	"github.com/AndB0ndar/doc-archive/internal/repository"
)

type SearchService struct {
	cfg            *config.Config
	chunkRepo      *repository.ChunkRepository
	embedderClient *Embedder
}

func NewSearchService(
	cfg *config.Config,
	repo *repository.ChunkRepository,
	embedder *Embedder,
) *SearchService {
	return &SearchService{
		cfg:            cfg,
		chunkRepo:      repo,
		embedderClient: embedder,
	}
}

type SearchRequest struct {
	Query string
	Type  string
	Limit int
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

	switch req.Type {
	case "", "text":
		return s.chunkRepo.FullTextSearchChunks(req.Query, req.Limit)
	case "vector", "semantic":
		embedding, err := s.embedderClient.Embed(req.Query)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrEmbedding, err)
		}
		return s.chunkRepo.SemanticSearchChunks(embedding, req.Limit)
	default:
		return nil, ErrInvalidType
	}
}
