package publisher

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultTurnsTopic     = "session.turns"
	DefaultCompletedTopic = "session.completed"

	RoleParticipant = "participant"
	RoleAssessor    = "assessor"
)

// SessionTurnMessage is the canonical session.turns event payload.
type SessionTurnMessage struct {
	EventType  string    `json:"event_type"` // "turn"
	SessionID  string    `json:"session_id"`
	OrgID      string    `json:"org_id"`
	TurnNumber int       `json:"turn_number"`
	Role       string    `json:"role"` // "participant" or "assessor"
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
}

type turnEvent = SessionTurnMessage

// SessionCompletedMessage is the canonical session.completed event payload.
type SessionCompletedMessage struct {
	EventType     string    `json:"event_type"` // "completed"
	SessionID     string    `json:"session_id"`
	OrgID         string    `json:"org_id"`
	ParticipantID string    `json:"participant_id"`
	BlueprintID   string    `json:"blueprint_id"`
	CompletedAt   time.Time `json:"completed_at"`
	TurnCount     int       `json:"turn_count"`
}

type completedEvent = SessionCompletedMessage

// Publisher publishes durable interview session events.
type Publisher interface {
	PublishTurn(ctx context.Context, sessionID, orgID string, turn int, role, content string) error
	PublishCompleted(ctx context.Context, sessionID, orgID, participantID, blueprintID string, turnCount int) error
	Close() error
}

type Config struct {
	KafkaBrokers   []string
	DatabaseURL    string
	TurnsTopic     string
	CompletedTopic string
}

type kafkaFactory func(Config) Publisher
type outboxFactory func(context.Context, string) (Publisher, error)

// New selects Kafka when at least one broker is configured and Postgres outbox otherwise.
func New(ctx context.Context, cfg Config) (Publisher, error) {
	return newPublisher(ctx, cfg, newKafkaPublisher, newPostgresPublisher)
}

func newPublisher(ctx context.Context, cfg Config, newKafka kafkaFactory, newOutbox outboxFactory) (Publisher, error) {
	cfg.KafkaBrokers = cleanBrokers(cfg.KafkaBrokers)
	if len(cfg.KafkaBrokers) > 0 {
		if cfg.TurnsTopic == "" {
			cfg.TurnsTopic = DefaultTurnsTopic
		}
		if cfg.CompletedTopic == "" {
			cfg.CompletedTopic = DefaultCompletedTopic
		}
		return newKafka(cfg), nil
	}
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, errors.New("DATABASE_URL is required when KAFKA_BROKERS is not set")
	}
	pub, err := newOutbox(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("initialize Postgres outbox publisher: %w", err)
	}
	return pub, nil
}

func cleanBrokers(brokers []string) []string {
	clean := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		if broker = strings.TrimSpace(broker); broker != "" {
			clean = append(clean, broker)
		}
	}
	return clean
}

func validateTurn(event turnEvent) error {
	if event.SessionID == "" || event.OrgID == "" {
		return errors.New("session_id and org_id are required")
	}
	if event.TurnNumber < 0 {
		return errors.New("turn_number must not be negative")
	}
	if event.Role != RoleParticipant && event.Role != RoleAssessor {
		return fmt.Errorf("unsupported role %q", event.Role)
	}
	if event.Content == "" {
		return errors.New("content is required")
	}
	return nil
}

func validateCompleted(event completedEvent) error {
	if event.SessionID == "" || event.OrgID == "" || event.ParticipantID == "" || event.BlueprintID == "" {
		return errors.New("session_id, org_id, participant_id, and blueprint_id are required")
	}
	if event.TurnCount < 0 {
		return errors.New("turn_count must not be negative")
	}
	return nil
}
