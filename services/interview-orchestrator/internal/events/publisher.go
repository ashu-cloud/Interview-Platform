package events

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/publisher"
)

// Publisher defines the contract for publishing session turn and completion events.
type Publisher interface {
	PublishTurn(ctx context.Context, sessionID, orgID string, turn int, role, content string) error
	PublishCompleted(ctx context.Context, sessionID, orgID, participantID, blueprintID string, turnCount int) error
}

// SessionTurnMessage is the canonical session.turns payload shape.
type SessionTurnMessage = publisher.SessionTurnMessage

// SessionCompletedMessage is the canonical session.completed payload shape.
type SessionCompletedMessage = publisher.SessionCompletedMessage

// NewKafkaPublisher wraps the Kafka publisher implementation.
func NewKafkaPublisher(kafkaBrokers string) Publisher {
	return publisher.NewKafkaPublisher(kafkaBrokers)
}

// NewPostgresOutboxPublisher wraps the Postgres outbox publisher implementation.
func NewPostgresOutboxPublisher(pool *pgxpool.Pool) Publisher {
	return publisher.NewPostgresOutboxPublisher(pool)
}

// NewPublisher selects mode based on KAFKA_BROKERS env var.
func NewPublisher(kafkaBrokers string, pool *pgxpool.Pool, logger *slog.Logger) Publisher {
	if kafkaBrokers != "" {
		if logger != nil {
			logger.Info("event publisher: Kafka mode", "brokers", kafkaBrokers)
		}
		return NewKafkaPublisher(kafkaBrokers)
	}
	if logger != nil {
		logger.Info("event publisher: Postgres outbox mode")
	}
	return NewPostgresOutboxPublisher(pool)
}
