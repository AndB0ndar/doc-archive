package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

// MockUserRepository implements domain.UserRepository for testing
type MockUserRepository struct {
	mu            sync.RWMutex
	users         map[string]*domain.User // key: user ID
	usersByEmail  map[string]*domain.User // key: email
	nextID        int
	createErr     error
	getByEmailErr error
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:        make(map[string]*domain.User),
		usersByEmail: make(map[string]*domain.User),
		nextID:       1,
	}
}

// Create implements domain.UserRepository.Create
func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createErr != nil {
		return "", m.createErr
	}

	// Generate a mock ID if not set
	if user.ID == "" {
		user.ID = fmt.Sprintf("mock-user-%d", m.nextID)
		m.nextID++
	}

	// Store the user
	m.users[user.ID] = user
	if user.Email != "" {
		m.usersByEmail[user.Email] = user
	}

	return user.ID, nil
}

// GetByEmail implements domain.UserRepository.GetByEmail
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}

	user, exists := m.usersByEmail[email]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Return a copy to prevent mutation
	return &domain.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}, nil
}

// SetCreateError sets an error to be returned by Create
func (m *MockUserRepository) SetCreateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createErr = err
}

// SetGetByEmailError sets an error to be returned by GetByEmail
func (m *MockUserRepository) SetGetByEmailError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getByEmailErr = err
}

// ClearErrors clears all configured errors
func (m *MockUserRepository) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createErr = nil
	m.getByEmailErr = nil
}

// GetUser returns a user by ID (test helper)
func (m *MockUserRepository) GetUser(id string) *domain.User {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.users[id]
}

// AddUser adds a user directly (test helper)
func (m *MockUserRepository) AddUser(user *domain.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	if user.Email != "" {
		m.usersByEmail[user.Email] = user
	}
}

// Clear clears all users (test helper)
func (m *MockUserRepository) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = make(map[string]*domain.User)
	m.usersByEmail = make(map[string]*domain.User)
	m.nextID = 1
	m.createErr = nil
	m.getByEmailErr = nil
}
