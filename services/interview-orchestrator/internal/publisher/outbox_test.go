package publisher

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type execCall struct {
	sql  string
	args []any
}

type fakeDB struct {
	calls  []execCall
	err    error
	closed bool
}

func (f *fakeDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	f.calls = append(f.calls, execCall{sql: sql, args: args})
	return pgconn.CommandTag{}, f.err
}

func (f *fakeDB) Close() { f.closed = true }

func TestOutboxPublisherInsertsCanonicalEvents(t *testing.T) {
	db := &fakeDB{}
	pub := newOutboxPublisher(db)
	before := time.Now().UTC()

	if err := pub.PublishTurn(context.Background(), "session", "org", 3, RoleParticipant, "answer"); err != nil {
		t.Fatal(err)
	}
	if err := pub.PublishCompleted(context.Background(), "session", "org", "participant", "blueprint", 3); err != nil {
		t.Fatal(err)
	}

	if len(db.calls) != 2 {
		t.Fatalf("got %d inserts, want 2", len(db.calls))
	}
	if !strings.Contains(db.calls[0].sql, "INSERT INTO outbox_turns") {
		t.Fatalf("unexpected turn SQL: %s", db.calls[0].sql)
	}
	if got := db.calls[0].args; len(got) != 6 || got[0] != "session" || got[1] != "org" || got[2] != 3 || got[3] != RoleParticipant || got[4] != "answer" {
		t.Fatalf("unexpected turn args: %#v", got)
	}
	if !strings.Contains(db.calls[1].sql, "INSERT INTO outbox_completed") {
		t.Fatalf("unexpected completed SQL: %s", db.calls[1].sql)
	}
	if got := db.calls[1].args; len(got) != 6 || got[0] != "session" || got[1] != "org" || got[2] != "participant" || got[3] != "blueprint" || got[4] != 3 {
		t.Fatalf("unexpected completed args: %#v", got)
	}
	after := time.Now().UTC()
	for _, call := range db.calls {
		at, ok := call.args[5].(time.Time)
		if !ok || at.Before(before) || at.After(after) || at.Location() != time.UTC {
			t.Fatalf("publisher did not generate a current UTC timestamp: %#v", call.args[5])
		}
	}
}

func TestOutboxPublisherValidatesAndWrapsErrors(t *testing.T) {
	db := &fakeDB{err: errors.New("db down")}
	pub := newOutboxPublisher(db)

	if err := pub.PublishTurn(context.Background(), "session", "org", 0, "user", "answer"); err == nil {
		t.Fatal("expected invalid turn error")
	}
	err := pub.PublishCompleted(context.Background(), "session", "org", "participant", "blueprint", 0)
	if err == nil || !strings.Contains(err.Error(), "insert outbox_completed: db down") {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pub.Close(); err != nil || !db.closed {
		t.Fatalf("Close() = %v, closed = %v", err, db.closed)
	}
}
