package queue

import (
	"github.com/hibiken/asynq"

	"github.com/AndB0ndar/doc-archive/internal/tasks"
	"github.com/AndB0ndar/doc-archive/internal/transport/queue/handlers"
)

func NewRouter(docHandler *handlers.DocumentHandler) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeProcessDocument, docHandler.ProcessDocument)
	return mux
}
