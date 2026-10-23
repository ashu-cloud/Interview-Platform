package config

import (
	"testing"
)

func setRequiredWorkerEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("PLATFORM_SERVICE_URL", "http://job-candidate-service:8001")
	t.Setenv("INTERNAL_API_KEY", "test-internal-key")
	t.Setenv("EVALUATION_LLM_MODE", "mock")
}

func TestLoadWorkerConfig(t *testing.T) {
	t.Run("outbox mode when KAFKA_BROKERS empty", func(t *testing.T) {
		setRequiredWorkerEnv(t)
		t.Setenv("KAFKA_BROKERS", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.KafkaBrokers != "" {
			t.Fatalf("expected empty KafkaBrokers, got %q", cfg.KafkaBrokers)
		}
	})

	t.Run("cloud deployment falls back to outbox when localhost broker configured", func(t *testing.T) {
		setRequiredWorkerEnv(t)
		t.Setenv("RENDER", "true")
		t.Setenv("KAFKA_BROKERS", "localhost:9092")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.KafkaBrokers != "" {
			t.Fatalf("expected KafkaBrokers to be cleared in cloud when pointing to localhost, got %q", cfg.KafkaBrokers)
		}
	})

	t.Run("USE_OUTBOX forces outbox mode", func(t *testing.T) {
		setRequiredWorkerEnv(t)
		t.Setenv("USE_OUTBOX", "true")
		t.Setenv("KAFKA_BROKERS", "external-kafka.domain.com:9092")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.KafkaBrokers != "" {
			t.Fatalf("expected KafkaBrokers to be cleared when USE_OUTBOX=true, got %q", cfg.KafkaBrokers)
		}
	})
}
