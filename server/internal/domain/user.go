package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
}

func NewUser(email, plainPassword string) (*User, error) {
	if email == "" {
		return nil, errors.New("email required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
	}, nil
}

func (u *User) ValidatePassword(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain)) == nil
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type AuthService interface {
	Register(ctx context.Context, email, password string) (string, *User, error) // returns token, user, error
	Login(ctx context.Context, email, password string) (string, *User, error)
	GenerateToken(userID string) (string, error)
}
