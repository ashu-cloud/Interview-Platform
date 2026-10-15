package publisher

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresPublisherIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("INTEGRATION_TEST is not 1")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL must be set when INTEGRATION_TEST=1")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("database unavailable: %v", err)
	}

	sessionID := uuid.NewString()
	orgID := uuid.NewString()
	participantID := uuid.NewString()
	blueprintID := uuid.NewString()
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, table := range []string{"outbox_turns", "outbox_completed"} {
			if _, err := pool.Exec(cleanupCtx, "DELETE FROM "+table+" WHERE session_id = $1", sessionID); err != nil {
				t.Errorf("cleanup %s: %v", table, err)
			}
		}
	})

	pub := newOutboxPublisher(pool)
	if err := pub.PublishTurn(ctx, sessionID, orgID, 1, RoleParticipant, "integration answer"); err != nil {
		t.Fatalf("publish turn: %v", err)
	}
	if err := pub.PublishCompleted(ctx, sessionID, orgID, participantID, blueprintID, 1); err != nil {
		t.Fatalf("publish completed: %v", err)
	}

	var role, content string
	if err := pool.QueryRow(ctx,
		"SELECT role, content FROM outbox_turns WHERE session_id = $1", sessionID,
	).Scan(&role, &content); err != nil {
		t.Fatalf("read outbox_turns: %v", err)
	}
	if role != RoleParticipant || content != "integration answer" {
		t.Fatalf("unexpected turn row: role=%q content=%q", role, content)
	}

	var gotParticipantID, gotBlueprintID string
	if err := pool.QueryRow(ctx,
		"SELECT participant_id::text, blueprint_id::text FROM outbox_completed WHERE session_id = $1", sessionID,
	).Scan(&gotParticipantID, &gotBlueprintID); err != nil {
		t.Fatalf("read outbox_completed: %v", err)
	}
	if gotParticipantID != participantID || gotBlueprintID != blueprintID {
		t.Fatalf("unexpected completed row: participant=%q blueprint=%q", gotParticipantID, gotBlueprintID)
	}
}
