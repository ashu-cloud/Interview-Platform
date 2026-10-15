package events

import (
	"log/slog"
	"testing"
)

func TestNewPublisher(t *testing.T) {
	logger := slog.Default()

	// Kafka mode when brokers specified
	pubKafka := NewPublisher("localhost:9092", nil, logger)
	if pubKafka == nil {
		t.Fatal("expected non-nil publisher in Kafka mode")
	}

	// Outbox mode when brokers empty
	pubOutbox := NewPublisher("", nil, logger)
	if pubOutbox == nil {
		t.Fatal("expected non-nil publisher in Outbox mode")
	}
}
