package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        int
	Env         string
	UploadDir   string
	EmbedderURL string
	Database    DatabaseConfig
}

type DatabaseConfig struct {
	URL               string
	MigrationsPath    string
	MaxOpenConns      int
	MaxIdleConns      int
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:        port,
		UploadDir:   getEnv("UPLOAD_DIR", "uploads"),
		EmbedderURL: getEnv("EMBEDDER_URL", "http://localhost:5001"),
		Env:         getEnv("ENV", "development"),
		Database: DatabaseConfig{
			URL:               getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/docdb?sslmode=disable"),
			MigrationsPath:    getEnv("MIGRATIONS_PATH", "migrations"),
			MaxOpenConns:      20,
			MaxIdleConns:      10,
			MaxConnLifetime:   30 * time.Minute,
			MaxConnIdleTime:   5 * time.Minute,
			HealthCheckPeriod: 1 * time.Minute,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
