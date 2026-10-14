package repository

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/turn-projector/internal/model"
)

// Integration tests run against a migrated Postgres (migrations 000001-000003).
// They are opt-in; once enabled, configuration and connectivity errors are failures.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("set INTEGRATION_TEST=1 to run database integration tests")
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Fatal("DATABASE_URL is required when INTEGRATION_TEST=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("database unavailable: %v", err)
	}
	var reg *string
	if err := pool.QueryRow(ctx, `SELECT to_regclass('public.outbox_turns')::text`).Scan(&reg); err != nil {
		pool.Close()
		t.Fatalf("check integration schema: %v", err)
	}
	if reg == nil {
		pool.Close()
		t.Fatal("migrations not applied (outbox_turns missing)")
	}
	t.Cleanup(pool.Close)
	return pool
}

type fixture struct{ orgID, sessionID string }

// seedSession creates org → user → track → blueprint → participant → invite → session.
func seedSession(t *testing.T, pool *pgxpool.Pool) fixture {
	t.Helper()
	ctx := context.Background()
	f := fixture{orgID: uuid.NewString(), sessionID: uuid.NewString()}
	userID, trackID, bpID, partID, inviteID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	steps := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO organizations (id, name, slug) VALUES ($1, 'Projector Test', $2)`, []any{f.orgID, f.orgID}},
		{`INSERT INTO users (id, org_id, email, password_hash, full_name) VALUES ($1, $2, $3, 'h', 'T')`, []any{userID, f.orgID, userID + "@t.test"}},
		{`INSERT INTO assessment_tracks (id, org_id, created_by, title) VALUES ($1, $2, $3, 'T')`, []any{trackID, f.orgID, userID}},
		{`INSERT INTO assessment_blueprints (id, job_role_id, org_id, title) VALUES ($1, $2, $3, 'B')`, []any{bpID, trackID, f.orgID}},
		{`INSERT INTO participants (id, org_id, job_role_id, email, full_name) VALUES ($1, $2, $3, $4, 'P')`, []any{partID, f.orgID, trackID, partID + "@t.test"}},
		{`INSERT INTO assessment_invites (id, candidate_id, docket_id, org_id, token, expires_at) VALUES ($1, $2, $3, $4, $5, NOW() + INTERVAL '1 day')`, []any{inviteID, partID, bpID, f.orgID, uuid.NewString()}},
		{`INSERT INTO assessment_sessions (id, invite_id, candidate_id, docket_id, org_id, status) VALUES ($1, $2, $3, $4, $5, 'in_progress')`, []any{f.sessionID, inviteID, partID, bpID, f.orgID}},
	}
	for _, s := range steps {
		if _, err := pool.Exec(ctx, s.sql, s.args...); err != nil {
			t.Fatalf("seed failed: %v\n%s", err, s.sql)
		}
	}
	t.Cleanup(func() {
		c := context.Background()
		_, _ = pool.Exec(c, `DELETE FROM outbox_turns WHERE org_id = $1`, f.orgID)
		_, _ = pool.Exec(c, `DELETE FROM session_turns WHERE org_id = $1`, f.orgID)
		_, _ = pool.Exec(c, `DELETE FROM assessment_sessions WHERE org_id = $1`, f.orgID)
		_, _ = pool.Exec(c, `DELETE FROM organizations WHERE id = $1`, f.orgID)
	})
	return f
}

func enqueue(t *testing.T, pool *pgxpool.Pool, orgID, sessionID string, turn int, role, content string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`INSERT INTO outbox_turns (session_id, org_id, turn_number, role, content, created_at)
		 VALUES ($1, $2, $3, $4, $5, clock_timestamp())`, sessionID, orgID, turn, role, content)
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
}

func enqueueAt(t *testing.T, pool *pgxpool.Pool, orgID, sessionID string, turn int, role, content string, createdAt time.Time) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO outbox_turns (session_id, org_id, turn_number, role, content, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id::text`, sessionID, orgID, turn, role, content, createdAt).Scan(&id)
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	return id
}

func TestConcurrentWorkersSkipLockedRow(t *testing.T) {
	pool := testPool(t)
	f := seedSession(t, pool)
	repo := NewTurnRepository(pool)
	ctx := context.Background()

	firstID := enqueueAt(t, pool, f.orgID, f.sessionID, 1, model.RoleParticipant, "first", time.Unix(1, 0))
	secondID := enqueueAt(t, pool, f.orgID, f.sessionID, 2, model.RoleAssessor, "second", time.Unix(2, 0))

	// Worker A selects and locks the oldest pending row but does not complete it.
	workerA, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin worker A: %v", err)
	}
	defer func() { _ = workerA.Rollback(context.Background()) }()
	var lockedID string
	if err := workerA.QueryRow(ctx, `
		SELECT id::text
		FROM outbox_turns
		WHERE status = 'pending' AND org_id = $1
		ORDER BY created_at ASC, id ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED`, f.orgID).Scan(&lockedID); err != nil {
		t.Fatalf("worker A lock: %v", err)
	}
	if lockedID != firstID {
		t.Fatalf("worker A locked %s, want first row %s", lockedID, firstID)
	}

	// Worker B must skip A's lock and complete the next pending row.
	workerBCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if n, err := repo.ProcessOutboxBatch(workerBCtx, 1); err != nil || n != 1 {
		t.Fatalf("worker B batch n=%d err=%v, want 1/nil", n, err)
	}

	var firstStatus, secondStatus string
	if err := pool.QueryRow(ctx, `
		SELECT
			(SELECT status FROM outbox_turns WHERE id = $1),
			(SELECT status FROM outbox_turns WHERE id = $2)`, firstID, secondID).Scan(&firstStatus, &secondStatus); err != nil {
		t.Fatalf("read outbox statuses: %v", err)
	}
	if firstStatus != "pending" || secondStatus != "completed" {
		t.Fatalf("statuses first=%q second=%q, want pending/completed", firstStatus, secondStatus)
	}
	var projectedTurn int
	if err := pool.QueryRow(ctx, `SELECT turn_number FROM session_turns WHERE session_id = $1`, f.sessionID).Scan(&projectedTurn); err != nil {
		t.Fatalf("read projected turn: %v", err)
	}
	if projectedTurn != 2 {
		t.Fatalf("projected turn=%d, want worker B's second turn", projectedTurn)
	}

	if err := workerA.Rollback(ctx); err != nil {
		t.Fatalf("release worker A lock: %v", err)
	}
	if n, err := repo.ProcessOutboxBatch(ctx, 1); err != nil || n != 1 {
		t.Fatalf("process released first row n=%d err=%v, want 1/nil", n, err)
	}
}

func TestConcurrentWorkersOverlappingRows(t *testing.T) {
	pool := testPool(t)
	f := seedSession(t, pool)
	repo := NewTurnRepository(pool)
	ctx := context.Background()

	// 10 turns x 2 roles = 20 distinct keys, each enqueued twice (overlapping events).
	for dup := 0; dup < 2; dup++ {
		for turn := 1; turn <= 10; turn++ {
			for _, role := range []string{model.RoleParticipant, model.RoleAssessor} {
				enqueue(t, pool, f.orgID, f.sessionID, turn, role, "content")
			}
		}
	}

	var wg sync.WaitGroup
	for w := 0; w < 2; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				n, err := repo.ProcessOutboxBatch(ctx, 5)
				if err != nil {
					var pe *ProjectionError
					if errors.As(err, &pe) {
						t.Errorf("unexpected projection failure: %v", err)
						return
					}
					continue // concurrency conflict (deadlock); rows stay pending, retry
				}
				if n == 0 {
					var pending int
					_ = pool.QueryRow(ctx, `SELECT count(*) FROM outbox_turns WHERE org_id=$1 AND status='pending'`, f.orgID).Scan(&pending)
					if pending == 0 {
						return
					}
				}
			}
		}()
	}
	wg.Wait()

	var turns, dupKeys, completed, retried int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM session_turns WHERE session_id=$1`, f.sessionID).Scan(&turns)
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM (SELECT 1 FROM session_turns WHERE session_id=$1
		GROUP BY session_id, turn_number, role HAVING count(*) > 1) d`, f.sessionID).Scan(&dupKeys)
	_ = pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE status='completed'), count(*) FILTER (WHERE retry_count>0)
		FROM outbox_turns WHERE org_id=$1`, f.orgID).Scan(&completed, &retried)
	if turns != 20 || dupKeys != 0 {
		t.Fatalf("session_turns rows=%d dupKeys=%d, want 20/0", turns, dupKeys)
	}
	if completed != 40 || retried != 0 {
		t.Fatalf("outbox completed=%d retried=%d, want 40/0", completed, retried)
	}
}

func TestRepeatedEventDoesNotDuplicateOrOverwrite(t *testing.T) {
	pool := testPool(t)
	f := seedSession(t, pool)
	repo := NewTurnRepository(pool)
	ctx := context.Background()

	enqueue(t, pool, f.orgID, f.sessionID, 1, model.RoleParticipant, "first")
	if n, err := repo.ProcessOutboxBatch(ctx, 10); err != nil || n != 1 {
		t.Fatalf("first batch n=%d err=%v", n, err)
	}
	enqueue(t, pool, f.orgID, f.sessionID, 1, model.RoleParticipant, "second")
	if n, err := repo.ProcessOutboxBatch(ctx, 10); err != nil || n != 1 {
		t.Fatalf("second batch n=%d err=%v", n, err)
	}
	// Kafka-path insert of the same key must also be a no-op.
	if err := repo.InsertTurn(ctx, model.SessionTurnEvent{SessionID: f.sessionID, OrgID: f.orgID, TurnNumber: 1, Role: model.RoleParticipant, Content: "third"}); err != nil {
		t.Fatal(err)
	}

	var count int
	var content string
	_ = pool.QueryRow(ctx, `SELECT count(*), min(content) FROM session_turns WHERE session_id=$1`, f.sessionID).Scan(&count, &content)
	if count != 1 || content != "first" {
		t.Fatalf("count=%d content=%q, want 1/first", count, content)
	}
}

func TestFailingProjectionIncrementsRetryThenFails(t *testing.T) {
	pool := testPool(t)
	f := seedSession(t, pool)
	repo := NewTurnRepository(pool)
	ctx := context.Background()

	// org_id that does not exist violates session_turns.org_id FK → projection fails.
	badOrg := uuid.NewString()
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM outbox_turns WHERE org_id=$1`, badOrg) })
	enqueue(t, pool, badOrg, f.sessionID, 1, model.RoleParticipant, "x")

	for attempt := 1; attempt <= MaxOutboxRetries; attempt++ {
		_, err := repo.ProcessOutboxBatch(ctx, 10)
		var pe *ProjectionError
		if !errors.As(err, &pe) {
			t.Fatalf("attempt %d: want ProjectionError, got %v", attempt, err)
		}
		var retries int
		var status string
		var failedAt *time.Time
		_ = pool.QueryRow(ctx, `SELECT retry_count, status, failed_at FROM outbox_turns WHERE id=$1`, pe.OutboxID).Scan(&retries, &status, &failedAt)
		wantStatus := "pending"
		if attempt == MaxOutboxRetries {
			wantStatus = "failed"
		}
		if retries != attempt || status != wantStatus || (status == "failed") != (failedAt != nil) {
			t.Fatalf("attempt %d: retries=%d status=%s failedAt=%v", attempt, retries, status, failedAt)
		}
	}

	// A failed row is no longer picked up.
	if n, err := repo.ProcessOutboxBatch(ctx, 10); err != nil || n != 0 {
		t.Fatalf("after failure n=%d err=%v, want 0/nil", n, err)
	}
	var turns int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM session_turns WHERE session_id=$1`, f.sessionID).Scan(&turns)
	if turns != 0 {
		t.Fatalf("session_turns rows = %d, want 0 (rolled back)", turns)
	}
}

func TestOutboxEndToEndVerification(t *testing.T) {
	pool := testPool(t)
	f := seedSession(t, pool)
	repo := NewTurnRepository(pool)
	ctx := context.Background()

	// 1. Enqueue participant and assessor turns into outbox_turns (status = 'pending')
	enqueue(t, pool, f.orgID, f.sessionID, 1, model.RoleParticipant, "Hello from candidate")
	enqueue(t, pool, f.orgID, f.sessionID, 1, model.RoleAssessor, "Welcome to the assessment")

	// 2. Process outbox batch (turn-projector consumption)
	n, err := repo.ProcessOutboxBatch(ctx, 10)
	if err != nil {
		t.Fatalf("ProcessOutboxBatch failed: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 processed rows, got %d", n)
	}

	// 3. Verify outbox_turns status is 'completed'
	var completedCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM outbox_turns WHERE session_id = $1 AND status = 'completed'`, f.sessionID).Scan(&completedCount); err != nil {
		t.Fatalf("query outbox_turns status: %v", err)
	}
	if completedCount != 2 {
		t.Fatalf("expected 2 completed outbox rows, got %d", completedCount)
	}

	// 4. Verify session_turns has both turns
	var turnCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM session_turns WHERE session_id = $1`, f.sessionID).Scan(&turnCount); err != nil {
		t.Fatalf("query session_turns count: %v", err)
	}
	if turnCount != 2 {
		t.Fatalf("expected 2 session_turns rows, got %d", turnCount)
	}

	// 5. Replay verification: enqueue duplicate turn 1 participant again
	enqueue(t, pool, f.orgID, f.sessionID, 1, model.RoleParticipant, "Duplicate hello")
	n2, err := repo.ProcessOutboxBatch(ctx, 10)
	if err != nil {
		t.Fatalf("ProcessOutboxBatch replay failed: %v", err)
	}
	if n2 != 1 {
		t.Fatalf("expected 1 processed replay row, got %d", n2)
	}

	// Ensure session_turns still only has 1 row for (turn 1, participant) and content is not overwritten
	var finalTurnCount int
	var turn1Content string
	if err := pool.QueryRow(ctx, `SELECT count(*), min(content) FROM session_turns WHERE session_id = $1 AND turn_number = 1 AND role = $2`, f.sessionID, model.RoleParticipant).Scan(&finalTurnCount, &turn1Content); err != nil {
		t.Fatalf("query replay session_turns: %v", err)
	}
	if finalTurnCount != 1 || turn1Content != "Hello from candidate" {
		t.Fatalf("replay failed: count=%d, content=%q", finalTurnCount, turn1Content)
	}
}

