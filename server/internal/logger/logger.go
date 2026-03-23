package logger

import (
	"log/slog"
	"os"
)

func New(env string) *slog.Logger {
	var handler slog.Handler
	if env == "production" {
		opts := &slog.HandlerOptions{Level: slog.LevelInfo}
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts := &slog.HandlerOptions{Level: slog.LevelDebug}
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

func Setup(env string) {
	slog.SetDefault(New(env))
}
