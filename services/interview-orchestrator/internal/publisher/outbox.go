package publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func nowUTC() time.Time { return time.Now().UTC() }

type postgresDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Close()
}

type postgresPublisher struct {
	db postgresDB
}

func newPostgresPublisher(ctx context.Context, databaseURL string) (Publisher, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return newOutboxPublisher(pool), nil
}

func newOutboxPublisher(db postgresDB) Publisher {
	return &postgresPublisher{db: db}
}

// NewPostgresOutboxPublisher creates a PostgresOutboxPublisher wrapping pgxpool.Pool.
func NewPostgresOutboxPublisher(pool *pgxpool.Pool) Publisher {
	return newOutboxPublisher(pool)
}

func (p *postgresPublisher) PublishTurn(ctx context.Context, sessionID, orgID string, turn int, role, content string) error {
	event := turnEvent{
		EventType:  "turn",
		SessionID:  sessionID,
		OrgID:      orgID,
		TurnNumber: turn,
		Role:       role,
		Content:    content,
		Timestamp:  nowUTC(),
	}
	if err := validateTurn(event); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `
		INSERT INTO outbox_turns (session_id, org_id, turn_number, role, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		event.SessionID, event.OrgID, event.TurnNumber, event.Role, event.Content, event.Timestamp)
	if err != nil {
		return fmt.Errorf("insert outbox_turns: %w", err)
	}
	return nil
}

func (p *postgresPublisher) PublishCompleted(ctx context.Context, sessionID, orgID, participantID, blueprintID string, turnCount int) error {
	event := completedEvent{
		EventType:     "completed",
		SessionID:     sessionID,
		OrgID:         orgID,
		ParticipantID: participantID,
		BlueprintID:   blueprintID,
		TurnCount:     turnCount,
		CompletedAt:   nowUTC(),
	}
	if err := validateCompleted(event); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `
		INSERT INTO outbox_completed (session_id, org_id, participant_id, blueprint_id, turn_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		event.SessionID, event.OrgID, event.ParticipantID, event.BlueprintID, event.TurnCount, event.CompletedAt)
	if err != nil {
		return fmt.Errorf("insert outbox_completed: %w", err)
	}
	return nil
}

func (p *postgresPublisher) Close() error {
	p.db.Close()
	return nil
}
