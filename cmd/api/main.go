package main

import (
	"log/slog"
	"os"

	"github.com/AndB0ndar/doc-archive/internal/app"
	"github.com/AndB0ndar/doc-archive/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	application := app.New(cfg)
	if err := application.Run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}
