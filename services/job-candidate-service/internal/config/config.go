package config

import (
	"fmt"
	"os"
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
