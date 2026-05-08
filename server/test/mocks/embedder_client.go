package mocks

import (
	"sync"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

// MockEmbedderClient implements domain.EmbedderClient for testing
type MockEmbedderClient struct {
	mu        sync.RWMutex
	embedErr  error
	embedding []float32
	calls     []string // Track calls for verification
}

// NewMockEmbedderClient creates a new mock embedder client
func NewMockEmbedderClient() *MockEmbedderClient {
	return &MockEmbedderClient{
		embedding: make([]float32, 384), // Default mock embedding size
		calls:     make([]string, 0),
	}
}

// Embed implements domain.EmbedderClient.Embed
func (m *MockEmbedderClient) Embed(text string) (*domain.EmbedResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track the call
	m.calls = append(m.calls, text)

	if m.embedErr != nil {
		return nil, m.embedErr
	}

	// Return a copy of the mock embedding
	embeddingCopy := make([]float32, len(m.embedding))
	copy(embeddingCopy, m.embedding)

	return &domain.EmbedResult{
		Embeddings: embeddingCopy,
	}, nil
}

// SetEmbedError sets an error to be returned by Embed
func (m *MockEmbedderClient) SetEmbedError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.embedErr = err
}

// SetEmbedding sets the embedding to be returned by Embed
func (m *MockEmbedderClient) SetEmbedding(embedding []float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.embedding = embedding
}

// GetCalls returns all texts that were embedded
func (m *MockEmbedderClient) GetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// ClearCalls clears the call history
func (m *MockEmbedderClient) ClearCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make([]string, 0)
}

// ClearErrors clears all configured errors
func (m *MockEmbedderClient) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.embedErr = nil
}

// Clear resets the mock to its initial state
func (m *MockEmbedderClient) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.embedErr = nil
	m.embedding = make([]float32, 384)
	m.calls = make([]string, 0)
}
