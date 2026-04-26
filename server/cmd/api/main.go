// @title           PDF Search API
// @version         1.0
// @description     API для интеллектуального поиска по документам.

// @contact.name   arbon

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
	if err := application.RunAPI(); err != nil {
		slog.Error("API failed", "error", err)
		os.Exit(1)
	}
}
