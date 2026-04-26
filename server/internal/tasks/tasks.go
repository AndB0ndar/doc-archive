package tasks

import (
	"encoding/json"
	"github.com/hibiken/asynq"
)

const TypeProcessDocument = "process:document"

type ProcessDocumentPayload struct {
	DocumentID string `json:"document_id"`
	ObjectKey  string `json:"object_key"`
}

func NewProcessDocumentTask(
	docID, objectKey string, opts ...asynq.Option,
) (*asynq.Task, error) {
	payload := ProcessDocumentPayload{
		DocumentID: docID,
		ObjectKey:  objectKey,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeProcessDocument, data, opts...), nil
}
