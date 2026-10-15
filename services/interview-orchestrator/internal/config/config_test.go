package config

import (
	"strings"
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JOB_CANDIDATE_SERVICE_URL", "http://job-candidate-service")
	t.Setenv("INTERNAL_API_KEY", "test-key")
	t.Setenv("INTERVIEW_DEMO_MODE", "true")
	t.Setenv("KAFKA_TURNS_TOPIC", "")
	t.Setenv("KAFKA_COMPLETED_TOPIC", "")
}

func TestLoadSelectsPublisherConfiguration(t *testing.T) {
	t.Run("Kafka does not require database URL", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("KAFKA_BROKERS", " kafka:9092 ")
		t.Setenv("DATABASE_URL", "")

		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.KafkaBrokers != "kafka:9092" {
			t.Fatalf("KafkaBrokers = %q", cfg.KafkaBrokers)
		}
		if cfg.KafkaTurnsTopic != "session.turns" || cfg.KafkaCompletedTopic != "session.completed" {
			t.Fatalf("unexpected topic defaults: %q, %q", cfg.KafkaTurnsTopic, cfg.KafkaCompletedTopic)
		}
	})

	t.Run("outbox requires database URL", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("KAFKA_BROKERS", "")
		t.Setenv("DATABASE_URL", "")

		_, err := Load()
		if err == nil || !strings.Contains(err.Error(), "DATABASE_URL is required") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("outbox accepts database URL", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("KAFKA_BROKERS", "")
		t.Setenv("DATABASE_URL", " postgres://db ")

		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.DatabaseURL != "postgres://db" {
			t.Fatalf("DatabaseURL = %q", cfg.DatabaseURL)
		}
	})
}
