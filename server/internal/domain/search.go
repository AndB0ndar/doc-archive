package domain

import "context"

type SearchResult struct {
	ChunkID    string
	DocumentID string
	Content    string
	Answer     *string
	Confidence *float64
}

type SearchQuery struct {
	Query string
	Type  string
}

type SearchService interface {
	Search(ctx context.Context, query SearchQuery, UserID string, Limit int) ([]SearchResult, error)
}
