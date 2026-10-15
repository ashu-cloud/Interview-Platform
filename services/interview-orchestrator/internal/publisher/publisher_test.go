package publisher

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type stubPublisher struct{}

func (stubPublisher) PublishTurn(context.Context, string, string, int, string, string) error {
	return nil
}
func (stubPublisher) PublishCompleted(context.Context, string, string, string, string, int) error {
	return nil
}
func (stubPublisher) Close() error { return nil }

func TestNewPublisherSelection(t *testing.T) {
	tests := []struct {
		name       string
		cfg        Config
		want       string
		wantErr    bool
		wantTopics []string
	}{
		{name: "Kafka when brokers set", cfg: Config{KafkaBrokers: []string{" broker:9092 "}}, want: "kafka", wantTopics: []string{DefaultTurnsTopic, DefaultCompletedTopic}},
		{name: "outbox without brokers", cfg: Config{DatabaseURL: "postgres://db"}, want: "outbox"},
		{name: "database required in outbox mode", cfg: Config{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := ""
			var gotTopics []string
			_, err := newPublisher(context.Background(), tt.cfg,
				func(cfg Config) Publisher {
					called = "kafka"
					gotTopics = []string{cfg.TurnsTopic, cfg.CompletedTopic}
					if !reflect.DeepEqual(cfg.KafkaBrokers, []string{"broker:9092"}) {
						t.Fatalf("brokers = %#v", cfg.KafkaBrokers)
					}
					return stubPublisher{}
				},
				func(context.Context, string) (Publisher, error) {
					called = "outbox"
					return stubPublisher{}, nil
				},
			)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if called != tt.want {
				t.Fatalf("selected %q, want %q", called, tt.want)
			}
			if tt.wantTopics != nil && !reflect.DeepEqual(gotTopics, tt.wantTopics) {
				t.Fatalf("topics = %#v, want %#v", gotTopics, tt.wantTopics)
			}
		})
	}
}

func TestNewPublisherWrapsOutboxFailure(t *testing.T) {
	_, err := newPublisher(context.Background(), Config{DatabaseURL: "postgres://db"},
		func(Config) Publisher { return stubPublisher{} },
		func(context.Context, string) (Publisher, error) { return nil, errors.New("unavailable") },
	)
	if err == nil || err.Error() != "initialize Postgres outbox publisher: unavailable" {
		t.Fatalf("unexpected error: %v", err)
	}
}
