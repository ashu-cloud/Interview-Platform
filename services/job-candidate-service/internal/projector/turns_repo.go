package projector

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MaxOutboxRetries is the number of failed projections after which an outbox row is marked failed.
const MaxOutboxRetries = 5

const insertTurnSQL = `
	INSERT INTO session_turns (session_id, org_id, turn_number, role, content)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (session_id, turn_number, role) DO NOTHING`

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// TurnRepository persists projected turns and manages outbox_turns rows.
type TurnRepository struct {
	pool *pgxpool.Pool
}

// NewTurnRepository constructs a TurnRepository.
func NewTurnRepository(pool *pgxpool.Pool) *TurnRepository {
	return &TurnRepository{pool: pool}
}

// ProjectionError reports a failed outbox projection after the failure was recorded.
type ProjectionError struct {
	OutboxID   string
	SessionID  string
	RetryCount int
	Status     string // 'pending' (will retry) or 'failed'
	Err        error
}

func (e *ProjectionError) Error() string {
	return fmt.Sprintf("project outbox row %s (session %s): %v", e.OutboxID, e.SessionID, e.Err)
}

func (e *ProjectionError) Unwrap() error { return e.Err }

// InsertTurn writes a turn to session_turns. Idempotent: an existing turn is never overwritten.
func (r *TurnRepository) InsertTurn(ctx context.Context, turn SessionTurnEvent) error {
	return insertTurn(ctx, r.pool, turn)
}

func insertTurn(ctx context.Context, db execer, turn SessionTurnEvent) error {
	_, err := db.Exec(ctx, insertTurnSQL, turn.SessionID, turn.OrgID, turn.TurnNumber, turn.Role, turn.Content)
	return err
}

// ProcessOutboxBatch projects up to limit pending outbox_turns rows in ONE transaction:
// select (FOR UPDATE SKIP LOCKED), insert into session_turns, mark completed.
func (r *TurnRepository) ProcessOutboxBatch(ctx context.Context, limit int) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id::text, session_id::text, org_id::text, turn_number, role, content, created_at
		FROM outbox_turns
		WHERE status = 'pending'
		ORDER BY created_at ASC, id ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return 0, fmt.Errorf("select pending: %w", err)
	}
	var batch []SessionTurnEvent
	for rows.Next() {
		var t SessionTurnEvent
		if err := rows.Scan(&t.ID, &t.SessionID, &t.OrgID, &t.TurnNumber, &t.Role, &t.Content, &t.Timestamp); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan outbox row: %w", err)
		}
		batch = append(batch, t)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate outbox rows: %w", err)
	}

	for _, t := range batch {
		if err := insertTurn(ctx, tx, t); err != nil {
			_ = tx.Rollback(ctx)
			if isConcurrencyConflict(err) {
				return 0, fmt.Errorf("concurrency conflict, batch will be retried: %w", err)
			}
			retries, status, recErr := r.recordFailure(ctx, t.ID)
			if recErr != nil {
				return 0, fmt.Errorf("projection failed (%v) and recording failure failed: %w", err, recErr)
			}
			return 0, &ProjectionError{OutboxID: t.ID, SessionID: t.SessionID, RetryCount: retries, Status: status, Err: err}
		}
		if _, err := tx.Exec(ctx,
			`UPDATE outbox_turns SET status = 'completed', processed_at = NOW() WHERE id = $1`, t.ID); err != nil {
			return 0, fmt.Errorf("mark completed: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return len(batch), nil
}

func (r *TurnRepository) recordFailure(ctx context.Context, id string) (int, string, error) {
	var retries int
	var status string
	err := r.pool.QueryRow(ctx, `
		UPDATE outbox_turns
		SET retry_count = retry_count + 1,
		    status      = CASE WHEN retry_count + 1 >= $2 THEN 'failed' ELSE 'pending' END,
		    failed_at   = CASE WHEN retry_count + 1 >= $2 THEN NOW() ELSE NULL END
		WHERE id = $1
		RETURNING retry_count, status`, id, MaxOutboxRetries).Scan(&retries, &status)
	return retries, status, err
}

func isConcurrencyConflict(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40P01" || pgErr.Code == "40001"
	}
	return false
}
