package config

import (
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	if _, err := Load(env(map[string]string{})); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.KafkaGroupID != "turn-projectors" || cfg.KafkaTopic != "session.turns" || cfg.PollInterval != 500*time.Millisecond {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

func TestModeSelection(t *testing.T) {
	cases := []struct {
		name    string
		brokers string
		want    Mode
	}{
		{"set selects kafka", "kafka:9092", ModeKafka},
		{"unset selects outbox", "", ModeOutbox},
		{"blank selects outbox", "   ", ModeOutbox},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x", "KAFKA_BROKERS": tc.brokers}))
			if err != nil {
				t.Fatal(err)
			}
			if got := cfg.Mode(); got != tc.want {
				t.Fatalf("Mode() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestInvalidPollInterval(t *testing.T) {
	_, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x", "POLL_INTERVAL": "nope"}))
	if err == nil {
		t.Fatal("expected error for invalid POLL_INTERVAL")
	}
}
