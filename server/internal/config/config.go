package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               int
	Env                string
	UploadDir          string
	EmbedderURL        string
	RerankerURL        string
	RerankerEnabled    bool
	ReaderURL          string
	ReaderEnabled      bool
	Database           DatabaseConfig
	SearchDefaultLimit int
	SearchMaxLimit     int
	Chunks             ChunkConfig
	JWTSecret          string
}

type ChunkConfig struct {
	Size       int
	Overlap    int
	Separators []string
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
		Port:               port,
		UploadDir:          getEnv("UPLOAD_DIR", "uploads"),
		EmbedderURL:        getEnv("EMBEDDER_URL", "http://localhost:5001"),
		RerankerURL:        getEnv("RERANKER_URL", "http://localhost:5001"),
		RerankerEnabled:    getEnv("RERANKER_ENABLED", "1") == "1",
		ReaderURL:          getEnv("READER_URL", "http://embedder:5001"),
		ReaderEnabled:      getEnv("READER_ENABLED", "1") == "1",
		Env:                getEnv("ENV", "development"),
		JWTSecret:          getEnv("SECRET_KEY", "default-secret-change-me"),
		SearchDefaultLimit: 20,
		SearchMaxLimit:     100,
		Chunks: ChunkConfig{
			Size:       500,
			Overlap:    1,
			Separators: []string{"\n\n", "\n", ". ", "? ", "! "},
		},
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
