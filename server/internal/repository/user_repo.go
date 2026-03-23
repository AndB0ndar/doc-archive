package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(
	ctx context.Context, user *domain.User,
) (string, error) {
	var id string
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		user.Email, user.PasswordHash,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert user: %w", err)
	}
	return id, nil
}

func (r *UserRepository) GetByEmail(
	ctx context.Context, email string,
) (*domain.User, error) {
	var dbUser models.UserDB
	err := r.db.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`,
		email,
	).Scan(&dbUser.ID, &dbUser.Email, &dbUser.PasswordHash, &dbUser.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &domain.User{
		ID:           dbUser.ID,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
	}, nil
}
