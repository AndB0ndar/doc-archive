package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/cache"
	"github.com/AndB0ndar/doc-archive/internal/domain"
)

type CachedUserRepository struct {
	repo  domain.UserRepository
	cache *cache.RedisCache
	ttl   time.Duration
}

func NewCachedUserRepository(
	repo domain.UserRepository,
	cache *cache.RedisCache,
	ttl time.Duration,
) domain.UserRepository {
	return &CachedUserRepository{
		repo:  repo,
		cache: cache,
		ttl:   ttl,
	}
}

func (c *CachedUserRepository) GetByEmail(
	ctx context.Context, email string,
) (*domain.User, error) {
	key := fmt.Sprintf("user:email:%s", email)

	var user domain.User
	found, err := c.cache.Get(ctx, key, &user)
	if err == nil && found {
		return &user, nil
	}

	userPtr, err := c.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = c.cache.Set(context.Background(), key, *userPtr, c.ttl)
	}()
	return userPtr, nil
}

func (c *CachedUserRepository) Create(
	ctx context.Context, user *domain.User,
) (string, error) {
	id, err := c.repo.Create(ctx, user)
	if err != nil {
		return "", err
	}
	go func() {
		key := fmt.Sprintf("user:email:%s", user.Email)
		_ = c.cache.Delete(context.Background(), key)
	}()
	return id, nil
}
