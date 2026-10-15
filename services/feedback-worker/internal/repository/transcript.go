package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionTurn struct {
	ID         string
	TurnNumber int
	Role       string // "participant" or "assessor"
	Content    string
	CreatedAt  time.Time
}

type TranscriptRepository struct {
	pool *pgxpool.Pool
}

func NewTranscriptRepository(pool *pgxpool.Pool) *TranscriptRepository {
	return &TranscriptRepository{pool: pool}
}

func (r *TranscriptRepository) GetTranscript(ctx context.Context, sessionID string) ([]SessionTurn, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, turn_number, role, content, created_at
		FROM session_turns
		WHERE session_id = $1
		ORDER BY turn_number ASC,
		         CASE WHEN role = 'participant' THEN 0 ELSE 1 END ASC,
		         created_at ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("query session turns: %w", err)
	}
	defer rows.Close()

	var turns []SessionTurn
	for rows.Next() {
		var t SessionTurn
		if err := rows.Scan(&t.ID, &t.TurnNumber, &t.Role, &t.Content, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan session turn: %w", err)
		}
		turns = append(turns, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("session turns rows: %w", err)
	}
	return turns, nil
}
