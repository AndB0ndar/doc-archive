package handlers

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"github.com/AndB0ndar/doc-archive/internal/domain"
	"github.com/AndB0ndar/doc-archive/internal/infrastructure/logger"
	"github.com/AndB0ndar/doc-archive/internal/tasks"
)

type DocumentHandler struct {
	docService domain.DocumentService
	logger     *logger.Logger
}

func NewDocumentHandler(
	docService domain.DocumentService,
	logger *logger.Logger,
) *DocumentHandler {
	return &DocumentHandler{
		docService: docService,
		logger:     logger,
	}
}

func (h *DocumentHandler) ProcessDocument(
	ctx context.Context, task *asynq.Task,
) error {
	var payload tasks.ProcessDocumentPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}
	h.logger.Info(
		"processing document",
		"doc_id", payload.DocumentID, "object_key", payload.ObjectKey,
	)

	err := h.docService.ProcessDocument(
		ctx, payload.DocumentID, payload.ObjectKey,
	)
	if err != nil {
		h.logger.Error(
			"processing failed", "doc_id", payload.DocumentID, "error", err,
		)
		return err
	}
	h.logger.Info(
		"processing document completed",
		"doc_id", payload.DocumentID, "object_key", payload.ObjectKey,
	)

	return nil
}
