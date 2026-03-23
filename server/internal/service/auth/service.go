package auth

import (
	"context"
	"errors"
	"time"

	"github.com/AndB0ndar/doc-archive/pkg/jwt"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

type Service struct {
	userRepo  domain.UserRepository
	jwtSecret string
	tokenTTL  time.Duration
}

func New(
	userRepo domain.UserRepository, jwtSecret string, tokenTTL time.Duration,
) domain.AuthService {
	return &Service{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		tokenTTL:  tokenTTL,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (string, *domain.User, error) {
	user, err := domain.NewUser(email, password)
	if err != nil {
		return "", nil, err
	}
	if _, err := s.userRepo.Create(ctx, user); err != nil {
		return "", nil, err
	}
	token, err := jwt.GenerateToken(user.ID, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	if !user.ValidatePassword(password) {
		return "", nil, errors.New("invalid credentials")
	}
	token, err := jwt.GenerateToken(user.ID, s.jwtSecret, s.tokenTTL)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *Service) GenerateToken(userID string) (string, error) {
	return jwt.GenerateToken(userID, s.jwtSecret, s.tokenTTL)
}
