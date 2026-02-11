package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        int
	BaseURL     string
	UploadDir   string
	EmbedderURL string
	DatabaseURL string
	Env         string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:        port,
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		UploadDir:   getEnv("UPLOAD_DIR", "uploads"),
		EmbedderURL: getEnv("EMBEDDER_URL", "http://localhost:5001"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:pass@localhost:5432/docdb?sslmode=disable"),
		Env:         getEnv("ENV", "development"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
