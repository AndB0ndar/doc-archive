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
	EmbedderURL string
	RerankerURL string
	ReaderURL   string
	Database    DatabaseConfig
	MinIO       MinIOConfig
	Queue       QueueConfig
	Search      SearchConfig
	Chunk       ChunkConfig
	JWTSecret   string
	JWTExpiry   int
	RedisURL    string
}

type ChunkConfig struct {
	Size             int
	Overlap          int
	Separators       []string
	SplitBySentences bool
}

type SearchConfig struct {
	DefaultLimit    int
	MaxLimit        int
	RerankerEnabled bool
	ReaderEnabled   bool
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

type MinIOConfig struct {
	URL       string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type QueueConfig struct {
	Concurrency int // for server
	MaxRetry    int // for client
	TimeoutSec  int // for client
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:        port,
		EmbedderURL: getEnv("EMBEDDER_URL", "http://embedder:5001"),
		RerankerURL: getEnv("RERANKER_URL", "http://embedder:5001"),
		ReaderURL:   getEnv("READER_URL", "http://embedder:5001"),
		Env:         getEnv("ENV", "development"),
		JWTSecret:   getEnv("SECRET_KEY", "default-secret-change-me"),
		JWTExpiry:   getEnvAsInt("JWTExpiry", 24), // hourse
		Search: SearchConfig{
			DefaultLimit:    20,
			MaxLimit:        100,
			RerankerEnabled: getEnv("RERANKER_ENABLED", "1") == "1",
			ReaderEnabled:   getEnv("READER_ENABLED", "1") == "1",
		},
		Chunk: ChunkConfig{
			Size:             500,
			Overlap:          1,
			Separators:       []string{"\n\n", "\n", ". ", "? ", "! "},
			SplitBySentences: true,
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
		MinIO: MinIOConfig{
			URL:       getEnv("MINIO_URL", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			Bucket:    getEnv("MINIO_BUCKET", "pdf-documents"),
			UseSSL:    getEnv("MINIO_USE_SSL", "0") == "1",
		},
		Queue: QueueConfig{
			Concurrency: getEnvAsInt("QUEUE_CONCURRENCY", 5),
			MaxRetry:    getEnvAsInt("QUEUE_MAX_RETRY", 3),
			TimeoutSec:  getEnvAsInt("QUEUE_TIMEOUT_SEC", 600),
		},
		RedisURL: getEnv("REDIS_URL", "redis:6379"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
