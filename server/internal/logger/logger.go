package logger

import (
	"log/slog"
	"os"
)

func New(env string) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	log := slog.New(handler)
	slog.SetDefault(log)
	return log
}

func Setup(env string) {
	New(env)
}
