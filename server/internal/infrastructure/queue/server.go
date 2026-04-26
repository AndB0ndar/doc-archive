package queue

import (
	"time"

	"github.com/hibiken/asynq"
)

type Server struct {
	asynq *asynq.Server
}

func NewServer(
	RedisAddr string,
	RedisDB int,
	Concurrency int,
) *Server {
	redisOpt := asynq.RedisClientOpt{Addr: RedisAddr, DB: RedisDB}
	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: Concurrency,
		Queues:      map[string]int{"default": 1},
		RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
			return time.Duration(1<<uint(n)) * time.Second
		},
	})
	return &Server{asynq: srv}
}

func (s *Server) Run(mux *asynq.ServeMux) error {
	return s.asynq.Run(mux)
}

func (s *Server) Stop() {
	s.asynq.Shutdown()
}
