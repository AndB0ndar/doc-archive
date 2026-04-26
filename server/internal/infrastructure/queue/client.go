package queue

import (
	"time"

	"github.com/hibiken/asynq"
)

type Client struct {
	asynq       *asynq.Client
	MaxRetry    int
	TaskTimeout time.Duration
}

func NewClient(
	RedisAddr string,
	RedisDB int,
	MaxRetry int,
	TaskTimeout time.Duration,
) (*Client, error) {
	redisOpt := asynq.RedisClientOpt{Addr: RedisAddr, DB: RedisDB}
	return &Client{
		asynq:       asynq.NewClient(redisOpt),
		MaxRetry:    MaxRetry,
		TaskTimeout: TaskTimeout,
	}, nil
}

func (c *Client) Enqueue(task *asynq.Task) (string, error) {
	info, err := c.asynq.Enqueue(task,
		asynq.MaxRetry(c.MaxRetry),
		asynq.Timeout(c.TaskTimeout),
	)
	if err != nil {
		return "", err
	}
	return info.ID, nil
}

func (c *Client) Close() error {
	return c.asynq.Close()
}
