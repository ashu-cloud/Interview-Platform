package config

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Mode identifies how the projector receives turn events.
type Mode string

const (
	ModeKafka  Mode = "kafka"
	ModeOutbox Mode = "outbox"
)

// Config holds turn-projector settings.
type Config struct {
	KafkaBrokers string        // optional — empty selects outbox mode
	KafkaGroupID string        // default: "turn-projectors"
	KafkaTopic   string        // default: "session.turns"
	DatabaseURL  string        // required
	PollInterval time.Duration // default: 500ms
}

// Load reads configuration using the supplied lookup function (os.Getenv in production).
func Load(getenv func(string) string) (*Config, error) {
	cfg := &Config{
		KafkaBrokers: strings.TrimSpace(getenv("KAFKA_BROKERS")),
		KafkaGroupID: strings.TrimSpace(getenv("KAFKA_GROUP_ID")),
		KafkaTopic:   strings.TrimSpace(getenv("KAFKA_TOPIC")),
		DatabaseURL:  strings.TrimSpace(getenv("DATABASE_URL")),
		PollInterval: 500 * time.Millisecond,
	}
	if cfg.KafkaGroupID == "" {
		cfg.KafkaGroupID = "turn-projectors"
	}
	if cfg.KafkaTopic == "" {
		cfg.KafkaTopic = "session.turns"
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if raw := strings.TrimSpace(getenv("POLL_INTERVAL")); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			return nil, fmt.Errorf("POLL_INTERVAL must be a positive duration, got %q", raw)
		}
		cfg.PollInterval = d
	}
	return cfg, nil
}

// Mode returns Kafka mode when KAFKA_BROKERS is set, otherwise Postgres outbox mode.
func (c *Config) Mode() Mode {
	if c.KafkaBrokers != "" {
		return ModeKafka
	}
	return ModeOutbox
}
