package repository

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/cache"
)

type CachedDocumentRepository struct {
	repo  domain.DocumentRepository
	cache *cache.RedisCache
	ttl   time.Duration
	log   *slog.Logger
}

func NewCachedDocumentRepository(
	repo domain.DocumentRepository,
	cache *cache.RedisCache,
	ttl time.Duration,
	log *slog.Logger,
) domain.DocumentRepository {
	return &CachedDocumentRepository{
		repo:  repo,
		cache: cache,
		ttl:   ttl,
		log:   log,
	}
}

func (c *CachedDocumentRepository) GetByID(
	ctx context.Context, id string, userID string,
) (*domain.Document, error) {
	key := fmt.Sprintf("doc:%s:%s", userID, id)

	var doc domain.Document
	found, err := c.cache.Get(ctx, key, &doc)
	if err == nil && found {
		return &doc, nil
	}

	docPtr, err := c.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = c.cache.Set(context.Background(), key, *docPtr, c.ttl)
	}()
	return docPtr, nil
}

func (c *CachedDocumentRepository) GetByUserID(
	ctx context.Context, userID string, limit, offset int,
) ([]*domain.Document, error) {
	key := fmt.Sprintf(
		"user_docs:%s:limit:%d:offset:%d", userID, limit, offset,
	)

	var docs []*domain.Document
	found, err := c.cache.Get(ctx, key, &docs)
	if err == nil && found {
		return docs, nil
	}

	docs, err = c.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = c.cache.Set(context.Background(), key, docs, c.ttl)
	}()
	return docs, nil
}

func (c *CachedDocumentRepository) UpdateStatus(
	ctx context.Context, id string, status domain.DocumentStatus,
) error {
	err := c.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	if err := c.cache.DeleteByPattern(ctx, "doc:*"); err != nil {
		_ = err
	}
	return nil
}

func (c *CachedDocumentRepository) Create(
	ctx context.Context, doc *domain.Document,
) (string, error) {
	id, err := c.repo.Create(ctx, doc)
	if err != nil {
		return "", err
	}
	go func() {
		pattern := fmt.Sprintf("user_docs:%s:*", doc.UserID)
		_ = c.cache.DeleteByPattern(context.Background(), pattern)
	}()
	return id, nil
}

func (c *CachedDocumentRepository) Delete(
	ctx context.Context, id string, userID string,
) error {
	err := c.repo.Delete(ctx, id, userID)
	if err != nil {
		return err
	}
	go func() {
		docKey := fmt.Sprintf("doc:%s:%s", userID, id)
		_ = c.cache.Delete(context.Background(), docKey)
		pattern := fmt.Sprintf("user_docs:%s:*", userID)
		_ = c.cache.DeleteByPattern(context.Background(), pattern)
	}()
	return nil
}
