package document

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
	"github.com/AndB0ndar/doc-archive/test/helpers"
	"github.com/AndB0ndar/doc-archive/test/mocks"
)

// mockFileStorage implements domain.FileStorage for testing
type mockFileStorage struct {
	uploadErr   error
	downloadErr error
	deleteErr   error
	urlErr      error
	deletedKeys []string
}

func (m *mockFileStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	return m.uploadErr
}

func (m *mockFileStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	return io.NopCloser(bytes.NewReader([]byte("test content"))), nil
}

func (m *mockFileStorage) Delete(ctx context.Context, objectKey string) error {
	m.deletedKeys = append(m.deletedKeys, objectKey)
	return m.deleteErr
}

func (m *mockFileStorage) GetPresignedURL(ctx context.Context, objectKey string, expires time.Duration) (string, error) {
	if m.urlErr != nil {
		return "", m.urlErr
	}
	return "http://example.com/presigned-url", nil
}

// mockTaskQueue implements domain.TaskQueue for testing
type mockTaskQueue struct {
	enqueueErr error
}

func (m *mockTaskQueue) EnqueueAny(task interface{}) (string, error) {
	if m.enqueueErr != nil {
		return "", m.enqueueErr
	}
	return "task-id", nil
}

// mockChunkerService implements domain.ChunkerService for testing
type mockChunkerService struct {
	splitErr error
	chunks   []string
}

func (m *mockChunkerService) Split(text string) ([]string, error) {
	if m.splitErr != nil {
		return nil, m.splitErr
	}
	if m.chunks != nil {
		return m.chunks, nil
	}
	return []string{"chunk1", "chunk2"}, nil
}

// mockMultipartFile implements a simple file for testing
type mockMultipartFile struct {
	*bytes.Reader
}

func (m *mockMultipartFile) Close() error {
	return nil
}

func newMockMultipartFile(content []byte) *mockMultipartFile {
	return &mockMultipartFile{bytes.NewReader(content)}
}

// setupTestService creates a document service with mocked dependencies
func setupTestService(t *testing.T) (*service, *mocks.MockDocumentRepository, *mocks.MockChunkRepository) {
	t.Helper()

	docRepo := mocks.NewMockDocumentRepository()
	chunkRepo := mocks.NewMockChunkRepository()
	embedder := mocks.NewMockEmbedderClient()
	chunker := &mockChunkerService{}
	fileStorage := &mockFileStorage{}
	taskQueue := &mockTaskQueue{}
	log := logger.New("test")

	svc := &service{
		docRepo:     docRepo,
		chunkRepo:   chunkRepo,
		embedder:    embedder,
		chunker:     chunker,
		fileStorage: fileStorage,
		taskQueue:   taskQueue,
		logger:      log,
	}

	return svc, docRepo, chunkRepo
}

func TestDocumentService_Upload_Validation(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		fileContent []byte
		wantErr     bool
	}{
		{
			name:        "valid upload with title",
			title:       "Test Document",
			fileContent: []byte("PDF content"),
			wantErr:     false,
		},
		{
			name:        "upload without title returns error",
			title:       "",
			fileContent: []byte("PDF content"),
			wantErr:     true,
		},
		{
			name:        "upload with empty file",
			title:       "Test Document",
			fileContent: []byte(""),
			wantErr:     false, // Empty file is allowed (though unusual)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := helpers.ContextWithTimeout(t)
			svc, _, _ := setupTestService(t)

			file := newMockMultipartFile(tt.fileContent)
			_, err := svc.Upload(ctx, file, tt.title, "Author", "2024", "Category", "user-id")

			if tt.wantErr {
				helpers.AssertError(t, err)
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestDocumentService_Upload_YearValidation(t *testing.T) {
	tests := []struct {
		name    string
		year    string
		wantErr bool
	}{
		{
			name:    "valid year",
			year:    "2024",
			wantErr: false,
		},
		{
			name:    "empty year (optional)",
			year:    "",
			wantErr: false,
		},
		{
			name:    "invalid year string",
			year:    "not-a-year",
			wantErr: false, // Year is optional, invalid is ignored
		},
		{
			name:    "year too far in future",
			year:    "3000",
			wantErr: false, // Year is optional, invalid is ignored
		},
		{
			name:    "negative year",
			year:    "-2024",
			wantErr: false, // Year is optional, invalid is ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := helpers.ContextWithTimeout(t)
			svc, _, _ := setupTestService(t)

			fileContent := []byte("PDF content")
			file := newMockMultipartFile(fileContent)
			_, err := svc.Upload(ctx, file, "Test Document", "Author", tt.year, "Category", "user-id")

			if tt.wantErr {
				helpers.AssertError(t, err)
			} else {
				helpers.AssertNoError(t, err)
			}
		})
	}
}

func TestDocumentService_Upload_StorageFailure(t *testing.T) {
	t.Run("file storage upload failure returns error", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, _, _ := setupTestService(t)

		// Override file storage with one that returns error
		svc.fileStorage = &mockFileStorage{uploadErr: errors.New("storage upload failed")}

		fileContent := []byte("PDF content")
		file := newMockMultipartFile(fileContent)
		_, err := svc.Upload(ctx, file, "Test Document", "Author", "2024", "Category", "user-id")

		helpers.AssertError(t, err)
	})
}

func TestDocumentService_GetDocumentByID_Authorization(t *testing.T) {
	t.Run("user can retrieve their own document", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, docRepo, _ := setupTestService(t)

		// Create a document
		doc := &domain.Document{
			ID:        "test-doc-id",
			UserID:    "user1",
			Title:     "Test Document",
			FilePath:  "/test/path/document.pdf",
			FileSize:  1024,
			Status:    domain.DocumentStatusPending,
			CreatedAt: time.Now(),
		}
		_, err := docRepo.Create(ctx, doc)
		helpers.AssertNoError(t, err)

		// Retrieve with correct user
		retrievedDoc, err := svc.GetDocumentByID(ctx, "test-doc-id", "user1")

		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, retrievedDoc.ID, "test-doc-id")
		helpers.AssertEqual(t, retrievedDoc.UserID, "user1")
	})

	t.Run("user cannot retrieve another user's document", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, docRepo, _ := setupTestService(t)

		// Create a document with user1
		doc := &domain.Document{
			ID:        "test-doc-id",
			UserID:    "user1",
			Title:     "Test Document",
			FilePath:  "/test/path/document.pdf",
			FileSize:  1024,
			Status:    domain.DocumentStatusPending,
			CreatedAt: time.Now(),
		}
		_, err := docRepo.Create(ctx, doc)
		helpers.AssertNoError(t, err)

		// Try to retrieve with user2
		_, err = svc.GetDocumentByID(ctx, "test-doc-id", "user2")

		helpers.AssertError(t, err)
	})
}

func TestDocumentService_GetUserDocuments_Filtering(t *testing.T) {
	t.Run("returns only documents for specified user", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, docRepo, _ := setupTestService(t)

		// Create documents for user1
		for i := 1; i <= 3; i++ {
			doc := &domain.Document{
				ID:        fmt.Sprintf("user1-doc-%d", i),
				UserID:    "user1",
				Title:     fmt.Sprintf("User1 Document %d", i),
				FilePath:  fmt.Sprintf("/test/path/doc%d.pdf", i),
				FileSize:  int64(i * 1024),
				Status:    domain.DocumentStatusPending,
				CreatedAt: time.Now(),
			}
			_, err := docRepo.Create(ctx, doc)
			helpers.AssertNoError(t, err)
		}

		// Create documents for user2
		for i := 1; i <= 2; i++ {
			doc := &domain.Document{
				ID:        fmt.Sprintf("user2-doc-%d", i),
				UserID:    "user2",
				Title:     fmt.Sprintf("User2 Document %d", i),
				FilePath:  fmt.Sprintf("/test/path/doc%d.pdf", i),
				FileSize:  int64(i * 1024),
				Status:    domain.DocumentStatusPending,
				CreatedAt: time.Now(),
			}
			_, err := docRepo.Create(ctx, doc)
			helpers.AssertNoError(t, err)
		}

		// Retrieve user1's documents
		user1Docs, err := svc.GetUserDocuments(ctx, "user1", 10, 0)
		helpers.AssertNoError(t, err)

		// Should only have user1's documents
		helpers.AssertEqual(t, len(user1Docs), 3)
		for _, doc := range user1Docs {
			helpers.AssertEqual(t, doc.UserID, "user1")
		}

		// Retrieve user2's documents
		user2Docs, err := svc.GetUserDocuments(ctx, "user2", 10, 0)
		helpers.AssertNoError(t, err)

		// Should only have user2's documents
		helpers.AssertEqual(t, len(user2Docs), 2)
		for _, doc := range user2Docs {
			helpers.AssertEqual(t, doc.UserID, "user2")
		}
	})
}

func TestDocumentService_DeleteDocument_Authorization(t *testing.T) {
	t.Run("owner can delete their document", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, docRepo, _ := setupTestService(t)
		fileStorage := svc.fileStorage.(*mockFileStorage)

		// Create a document
		doc := &domain.Document{
			ID:        "test-doc-id",
			UserID:    "user1",
			Title:     "Test Document",
			FilePath:  "documents/test-doc-id.pdf",
			FileSize:  1024,
			Status:    domain.DocumentStatusPending,
			CreatedAt: time.Now(),
		}
		_, err := docRepo.Create(ctx, doc)
		helpers.AssertNoError(t, err)

		// Delete with correct user
		err = svc.DeleteDocument(ctx, "test-doc-id", "user1")

		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, len(fileStorage.deletedKeys), 1)
		helpers.AssertEqual(t, fileStorage.deletedKeys[0], "documents/test-doc-id.pdf")

		// Verify document is deleted
		_, err = svc.GetDocumentByID(ctx, "test-doc-id", "user1")
		helpers.AssertError(t, err)
	})

	t.Run("non-owner cannot delete document", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, docRepo, _ := setupTestService(t)

		// Create a document with user1
		doc := &domain.Document{
			ID:        "test-doc-id",
			UserID:    "user1",
			Title:     "Test Document",
			FilePath:  "documents/test-doc-id.pdf",
			FileSize:  1024,
			Status:    domain.DocumentStatusPending,
			CreatedAt: time.Now(),
		}
		_, err := docRepo.Create(ctx, doc)
		helpers.AssertNoError(t, err)

		// Try to delete with user2
		err = svc.DeleteDocument(ctx, "test-doc-id", "user2")

		helpers.AssertError(t, err)

		// Verify document still exists
		retrievedDoc, err := svc.GetDocumentByID(ctx, "test-doc-id", "user1")
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, retrievedDoc.ID, "test-doc-id")
	})

	t.Run("storage delete failure returns error and keeps document", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, docRepo, _ := setupTestService(t)
		svc.fileStorage = &mockFileStorage{deleteErr: errors.New("storage delete failed")}

		doc := &domain.Document{
			ID:        "test-doc-id",
			UserID:    "user1",
			Title:     "Test Document",
			FilePath:  "documents/test-doc-id.pdf",
			FileSize:  1024,
			Status:    domain.DocumentStatusPending,
			CreatedAt: time.Now(),
		}
		_, err := docRepo.Create(ctx, doc)
		helpers.AssertNoError(t, err)

		err = svc.DeleteDocument(ctx, "test-doc-id", "user1")
		helpers.AssertError(t, err)

		retrievedDoc, err := svc.GetDocumentByID(ctx, "test-doc-id", "user1")
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, retrievedDoc.ID, "test-doc-id")
	})
}

func TestDocumentService_ProcessDocument_Validation(t *testing.T) {
	t.Run("process document with empty docID returns error", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, _, _ := setupTestService(t)

		err := svc.ProcessDocument(ctx, "", "object-key")

		helpers.AssertError(t, err)
	})

	t.Run("process document with empty objectKey returns error", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, _, _ := setupTestService(t)

		err := svc.ProcessDocument(ctx, "doc-id", "")

		helpers.AssertError(t, err)
	})

	t.Run("process document with storage download failure", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, _, _ := setupTestService(t)

		// Override file storage with one that returns error on download
		svc.fileStorage = &mockFileStorage{downloadErr: errors.New("download failed")}

		err := svc.ProcessDocument(ctx, "doc-id", "object-key")

		helpers.AssertError(t, err)
	})

	t.Run("process document with text extraction failure", func(t *testing.T) {
		// We need to mock the parser.ExtractFromPDF function
		// Since we can't easily mock a package function, we'll test the integration
		// by creating a document and trying to process it with real dependencies
		// For now, we'll skip this test or implement it differently
		t.Skip("Would require mocking parser.ExtractFromPDF which is a package function")
	})

	t.Run("process document with chunking failure", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, _, _ := setupTestService(t)

		// Override chunker with one that returns error
		svc.chunker = &mockChunkerService{splitErr: errors.New("chunking failed")}

		err := svc.ProcessDocument(ctx, "doc-id", "object-key")

		helpers.AssertError(t, err)
	})

	t.Run("process document with repository failure", func(t *testing.T) {
		ctx := helpers.ContextWithTimeout(t)
		svc, _, chunkRepo := setupTestService(t)

		// Set up chunk repo to return error
		chunkRepo.SetCreateBatchError(errors.New("repository error"))

		err := svc.ProcessDocument(ctx, "doc-id", "object-key")

		helpers.AssertError(t, err)
	})
}
