package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration read from environment at startup.
// Fail fast: Load returns an error if any required variable is missing.
type Config struct {
	Port           string
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	InternalAPIKey string
	AppDomain      string

	// Embedding generation for blueprints (RAG + adaptive engine).
	OpenAIEmbeddingAPIKey string // empty (and not demo) disables embedding generation
	EmbeddingModel        string
	EmbeddingDims         int
	EmbeddingBaseURL      string
	DemoMode              bool // ASSESSMENT_DEMO_MODE: use random demo embeddings
}

// Load reads all required env vars and returns Config.
// It returns an error if any required variable is empty.
func Load() (*Config, error) {
	cfg := &Config{
		Port:           getEnv("PORT", "8001"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		RedisURL:       os.Getenv("REDIS_URL"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		InternalAPIKey: os.Getenv("INTERNAL_API_KEY"),
		AppDomain:      getEnv("APP_DOMAIN", "http://localhost:3000"),

		OpenAIEmbeddingAPIKey: strings.TrimSpace(os.Getenv("OPENAI_EMBEDDING_API_KEY")),
		EmbeddingModel:        getEnv("EMBEDDING_MODEL", "text-embedding-3-small"),
		EmbeddingBaseURL:      strings.TrimSpace(os.Getenv("EMBEDDING_BASE_URL")),
		DemoMode:              strings.EqualFold(os.Getenv("ASSESSMENT_DEMO_MODE"), "true") || strings.EqualFold(os.Getenv("INTERVIEW_DEMO_MODE"), "true"),
	}

	cfg.EmbeddingDims = 768
	if raw := strings.TrimSpace(os.Getenv("EMBEDDING_DIMS")); raw != "" {
		dims, err := strconv.Atoi(raw)
		if err != nil || dims <= 0 {
			return nil, fmt.Errorf("config: EMBEDDING_DIMS must be a positive integer, got %q", raw)
		}
		cfg.EmbeddingDims = dims
	}

	required := []struct {
		name string
		val  string
	}{
		{"DATABASE_URL", cfg.DatabaseURL},
		{"REDIS_URL", cfg.RedisURL},
		{"JWT_SECRET", cfg.JWTSecret},
		{"INTERNAL_API_KEY", cfg.InternalAPIKey},
	}

	for _, r := range required {
		if r.val == "" {
			return nil, fmt.Errorf("config: required environment variable %s is not set", r.name)
		}
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
