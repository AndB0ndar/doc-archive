package mocks

import (
	"sync"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

// MockRerankerClient implements domain.RerankerClient for testing
type MockRerankerClient struct {
	mu        sync.RWMutex
	rerankErr error
	scores    []float32
	calls     []rerankerCallRecord // Track calls for verification
}

type rerankerCallRecord struct {
	query    string
	passages []string
}

// NewMockRerankerClient creates a new mock reranker client
func NewMockRerankerClient() *MockRerankerClient {
	return &MockRerankerClient{
		scores: []float32{0.8, 0.7, 0.6}, // Default mock scores
		calls:  make([]rerankerCallRecord, 0),
	}
}

// Rerank implements domain.RerankerClient.Rerank
func (m *MockRerankerClient) Rerank(query string, passages []string) (*domain.RerankResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track the call
	m.calls = append(m.calls, rerankerCallRecord{
		query:    query,
		passages: passages,
	})

	if m.rerankErr != nil {
		return nil, m.rerankErr
	}

	// Return scores for each passage
	// If we have fewer scores than passages, repeat the last score
	resultScores := make([]float32, len(passages))
	for i := range resultScores {
		if i < len(m.scores) {
			resultScores[i] = m.scores[i]
		} else if len(m.scores) > 0 {
			resultScores[i] = m.scores[len(m.scores)-1]
		} else {
			resultScores[i] = 0.5 // Default score
		}
	}

	return &domain.RerankResult{
		Scores: resultScores,
	}, nil
}

// SetRerankError sets an error to be returned by Rerank
func (m *MockRerankerClient) SetRerankError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rerankErr = err
}

// SetScores sets the scores to be returned by Rerank
func (m *MockRerankerClient) SetScores(scores []float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scores = scores
}

// GetCalls returns all rerank calls
func (m *MockRerankerClient) GetCalls() []rerankerCallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	calls := make([]rerankerCallRecord, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// ClearCalls clears the call history
func (m *MockRerankerClient) ClearCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make([]rerankerCallRecord, 0)
}

// ClearErrors clears all configured errors
func (m *MockRerankerClient) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rerankErr = nil
}

// Clear resets the mock to its initial state
func (m *MockRerankerClient) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rerankErr = nil
	m.scores = []float32{0.8, 0.7, 0.6}
	m.calls = make([]rerankerCallRecord, 0)
}
