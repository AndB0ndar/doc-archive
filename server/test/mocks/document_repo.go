package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/domain"
)

// MockDocumentRepository implements domain.DocumentRepository for testing
type MockDocumentRepository struct {
	mu              sync.RWMutex
	documents       map[string]*domain.Document // key: document ID
	userDocuments   map[string][]string         // key: user ID, value: list of document IDs
	nextID          int
	createErr       error
	getByIDErr      error
	getByUserIDErr  error
	updateStatusErr error
	deleteErr       error
}

// NewMockDocumentRepository creates a new mock document repository
func NewMockDocumentRepository() *MockDocumentRepository {
	return &MockDocumentRepository{
		documents:     make(map[string]*domain.Document),
		userDocuments: make(map[string][]string),
		nextID:        1,
	}
}

// Create implements domain.DocumentRepository.Create
func (m *MockDocumentRepository) Create(ctx context.Context, doc *domain.Document) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.createErr != nil {
		return "", m.createErr
	}

	// Generate a mock ID if not set
	if doc.ID == "" {
		doc.ID = fmt.Sprintf("mock-doc-%d", m.nextID)
		m.nextID++
	}

	// Set default status if empty
	if doc.Status == "" {
		doc.Status = domain.DocumentStatusPending
	}

	// Set created at if empty
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}

	// Store the document
	m.documents[doc.ID] = doc
	m.userDocuments[doc.UserID] = append(m.userDocuments[doc.UserID], doc.ID)

	return doc.ID, nil
}

// GetByID implements domain.DocumentRepository.GetByID
func (m *MockDocumentRepository) GetByID(ctx context.Context, id string, userID string) (*domain.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}

	doc, exists := m.documents[id]
	if !exists {
		return nil, fmt.Errorf("document not found")
	}

	// Check user authorization
	if doc.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Return a copy to prevent mutation
	return &domain.Document{
		ID:        doc.ID,
		UserID:    doc.UserID,
		Title:     doc.Title,
		Authors:   doc.Authors,
		Year:      doc.Year,
		Category:  doc.Category,
		FilePath:  doc.FilePath,
		FileSize:  doc.FileSize,
		Status:    doc.Status,
		CreatedAt: doc.CreatedAt,
	}, nil
}

// GetByUserID implements domain.DocumentRepository.GetByUserID
func (m *MockDocumentRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.getByUserIDErr != nil {
		return nil, m.getByUserIDErr
	}

	docIDs, exists := m.userDocuments[userID]
	if !exists {
		return []*domain.Document{}, nil
	}

	// Apply offset and limit
	start := offset
	if start > len(docIDs) {
		start = len(docIDs)
	}
	end := start + limit
	if end > len(docIDs) {
		end = len(docIDs)
	}

	// Get documents
	docs := make([]*domain.Document, 0, end-start)
	for _, id := range docIDs[start:end] {
		if doc, ok := m.documents[id]; ok {
			// Return copies
			docs = append(docs, &domain.Document{
				ID:        doc.ID,
				UserID:    doc.UserID,
				Title:     doc.Title,
				Authors:   doc.Authors,
				Year:      doc.Year,
				Category:  doc.Category,
				FilePath:  doc.FilePath,
				FileSize:  doc.FileSize,
				Status:    doc.Status,
				CreatedAt: doc.CreatedAt,
			})
		}
	}

	return docs, nil
}

// UpdateStatus implements domain.DocumentRepository.UpdateStatus
func (m *MockDocumentRepository) UpdateStatus(ctx context.Context, docID string, status domain.DocumentStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}

	doc, exists := m.documents[docID]
	if !exists {
		return fmt.Errorf("document not found")
	}

	doc.Status = status
	return nil
}

// Delete implements domain.DocumentRepository.Delete
func (m *MockDocumentRepository) Delete(ctx context.Context, id string, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.deleteErr != nil {
		return m.deleteErr
	}

	doc, exists := m.documents[id]
	if !exists {
		return fmt.Errorf("document not found")
	}

	// Check user authorization
	if doc.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Remove from documents map
	delete(m.documents, id)

	// Remove from user documents list
	if docIDs, ok := m.userDocuments[userID]; ok {
		for i, docID := range docIDs {
			if docID == id {
				m.userDocuments[userID] = append(docIDs[:i], docIDs[i+1:]...)
				break
			}
		}
	}

	return nil
}

// SetCreateError sets an error to be returned by Create
func (m *MockDocumentRepository) SetCreateError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createErr = err
}

// SetGetByIDError sets an error to be returned by GetByID
func (m *MockDocumentRepository) SetGetByIDError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getByIDErr = err
}

// SetGetByUserIDError sets an error to be returned by GetByUserID
func (m *MockDocumentRepository) SetGetByUserIDError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getByUserIDErr = err
}

// SetUpdateStatusError sets an error to be returned by UpdateStatus
func (m *MockDocumentRepository) SetUpdateStatusError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateStatusErr = err
}

// SetDeleteError sets an error to be returned by Delete
func (m *MockDocumentRepository) SetDeleteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteErr = err
}

// ClearErrors clears all configured errors
func (m *MockDocumentRepository) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createErr = nil
	m.getByIDErr = nil
	m.getByUserIDErr = nil
	m.updateStatusErr = nil
	m.deleteErr = nil
}

// GetDocument returns a document by ID (test helper)
func (m *MockDocumentRepository) GetDocument(id string) *domain.Document {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.documents[id]
}

// AddDocument adds a document directly (test helper)
func (m *MockDocumentRepository) AddDocument(doc *domain.Document) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.documents[doc.ID] = doc
	m.userDocuments[doc.UserID] = append(m.userDocuments[doc.UserID], doc.ID)
}

// Clear clears all documents (test helper)
func (m *MockDocumentRepository) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.documents = make(map[string]*domain.Document)
	m.userDocuments = make(map[string][]string)
	m.nextID = 1
	m.createErr = nil
	m.getByIDErr = nil
	m.getByUserIDErr = nil
	m.updateStatusErr = nil
	m.deleteErr = nil
}
