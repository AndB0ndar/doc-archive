package mocks

import (
	"sync"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

// MockReaderClient implements domain.ReaderClient for testing
type MockReaderClient struct {
	mu        sync.RWMutex
	answerErr error
	answer    *domain.Answer
	calls     []readerCallRecord // Track calls for verification
}

type readerCallRecord struct {
	question string
	context  string
}

// NewMockReaderClient creates a new mock reader client
func NewMockReaderClient() *MockReaderClient {
	return &MockReaderClient{
		answer: &domain.Answer{
			Answer:     "This is a mock answer.",
			Confidence: 0.85,
			Start:      0,
			End:        23,
		},
		calls: make([]readerCallRecord, 0),
	}
}

// Answer implements domain.ReaderClient.Answer
func (m *MockReaderClient) Answer(question string, context string) (*domain.Answer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track the call
	m.calls = append(m.calls, readerCallRecord{
		question: question,
		context:  context,
	})

	if m.answerErr != nil {
		return nil, m.answerErr
	}

	// Return a copy of the mock answer
	return &domain.Answer{
		Answer:     m.answer.Answer,
		Confidence: m.answer.Confidence,
		Start:      m.answer.Start,
		End:        m.answer.End,
	}, nil
}

// SetAnswerError sets an error to be returned by Answer
func (m *MockReaderClient) SetAnswerError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.answerErr = err
}

// SetAnswer sets the answer to be returned by Answer
func (m *MockReaderClient) SetAnswer(answer *domain.Answer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.answer = answer
}

// GetCalls returns all answer calls
func (m *MockReaderClient) GetCalls() []readerCallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	calls := make([]readerCallRecord, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// ClearCalls clears the call history
func (m *MockReaderClient) ClearCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make([]readerCallRecord, 0)
}

// ClearErrors clears all configured errors
func (m *MockReaderClient) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.answerErr = nil
}

// Clear resets the mock to its initial state
func (m *MockReaderClient) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.answerErr = nil
	m.answer = &domain.Answer{
		Answer:     "This is a mock answer.",
		Confidence: 0.85,
		Start:      0,
		End:        23,
	}
	m.calls = make([]readerCallRecord, 0)
}
