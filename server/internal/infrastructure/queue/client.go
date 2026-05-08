package queue

import (
	"errors"
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

// EnqueueAny implements domain.TaskQueue interface
func (c *Client) EnqueueAny(task interface{}) (string, error) {
	asynqTask, ok := task.(*asynq.Task)
	if !ok {
		return "", errors.New("task must be *asynq.Task")
	}
	return c.Enqueue(asynqTask)
}

func (c *Client) Close() error {
	return c.asynq.Close()
}
