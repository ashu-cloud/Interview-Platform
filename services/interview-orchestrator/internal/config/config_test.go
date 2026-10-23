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

	t.Run("cloud deployment falls back from localhost kafka to outbox", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("RENDER", "true")
		t.Setenv("KAFKA_BROKERS", "localhost:9092")
		t.Setenv("DATABASE_URL", "postgres://db")

		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.KafkaBrokers != "" {
			t.Fatalf("expected KafkaBrokers to be cleared in cloud when pointing to localhost, got %q", cfg.KafkaBrokers)
		}
	})

	t.Run("CleanRedisURL normalizes upstash cli command", func(t *testing.T) {
		setRequiredEnv(t)
		t.Setenv("REDIS_URL", "redis-cli --tls -u redis://default:token@divine-ape-235380.upstash.io:6379")
		t.Setenv("DATABASE_URL", "postgres://db")
		t.Setenv("KAFKA_BROKERS", "")

		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		expected := "rediss://default:token@divine-ape-235380.upstash.io:6379"
		if cfg.RedisURL != expected {
			t.Fatalf("expected RedisURL %q, got %q", expected, cfg.RedisURL)
		}
	})
}

