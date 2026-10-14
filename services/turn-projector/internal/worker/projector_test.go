package worker

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yourusername/ai-interview/services/turn-projector/internal/config"
	"github.com/yourusername/ai-interview/services/turn-projector/internal/model"
)

type fakeStore struct {
	calls int
	err   error
}

func (f *fakeStore) InsertTurn(context.Context, model.SessionTurnEvent) error {
	f.calls++
	return f.err
}

type fakeCommitter struct{ commits int }

func (f *fakeCommitter) CommitMessages(context.Context, ...kafka.Message) error {
	f.commits++
	return nil
}

func newTestProjector(store *fakeStore, buf *bytes.Buffer) *TurnProjector {
	return &TurnProjector{
		cfg:     &config.Config{},
		turns:   store,
		logger:  slog.New(slog.NewJSONHandler(buf, nil)),
		backoff: func(int) time.Duration { return time.Millisecond },
	}
}

const validPayload = `{"interview_id":"11111111-1111-1111-1111-111111111111","org_id":"22222222-2222-2222-2222-222222222222","turn_number":1,"role":"user","content":"hi"}`

func TestRunSelectsModeFromConfig(t *testing.T) {
	cases := []struct {
		brokers   string
		wantMode  string
		wantLog   string
		wantOther string
	}{
		{"kafka:9092", "kafka", "turn-projector: Kafka mode", "outbox"},
		{"", "outbox", "turn-projector: Postgres outbox mode", "kafka"},
	}
	for _, tc := range cases {
		t.Run(tc.wantMode, func(t *testing.T) {
			var buf bytes.Buffer
			ran := ""
			p := newTestProjector(&fakeStore{}, &buf)
			p.cfg = &config.Config{KafkaBrokers: tc.brokers, PollInterval: time.Millisecond}
			p.runKafka = func(context.Context) { ran = "kafka" }
			p.runOutbox = func(context.Context) { ran = "outbox" }
			p.Run(context.Background())
			if ran != tc.wantMode {
				t.Fatalf("ran %q, want %q", ran, tc.wantMode)
			}
			if !strings.Contains(buf.String(), tc.wantLog) {
				t.Fatalf("log missing %q: %s", tc.wantLog, buf.String())
			}
		})
	}
}

func TestKafkaSuccessCommitsOnce(t *testing.T) {
	var buf bytes.Buffer
	store, c := &fakeStore{}, &fakeCommitter{}
	newTestProjector(store, &buf).handleKafkaMessage(context.Background(), c, kafka.Message{Value: []byte(validPayload)})
	if store.calls != 1 || c.commits != 1 {
		t.Fatalf("calls=%d commits=%d", store.calls, c.commits)
	}
}

func TestKafkaRepeatedFailureIsBoundedAndLogged(t *testing.T) {
	var buf bytes.Buffer
	store, c := &fakeStore{err: errors.New("db down")}, &fakeCommitter{}
	newTestProjector(store, &buf).handleKafkaMessage(context.Background(), c,
		kafka.Message{Topic: "session.turns", Partition: 1, Offset: 42, Value: []byte(validPayload)})
	if store.calls != kafkaMaxAttempts {
		t.Fatalf("attempts = %d, want %d", store.calls, kafkaMaxAttempts)
	}
	if c.commits != 1 {
		t.Fatalf("commits = %d, want 1 (skip after bounded retries)", c.commits)
	}
	if !strings.Contains(buf.String(), "failed permanently") || !strings.Contains(buf.String(), "11111111-1111-1111-1111-111111111111") {
		t.Fatalf("expected permanent-failure log with session_id: %s", buf.String())
	}
}

func TestKafkaMalformedMessageLoggedNotSilentlyDropped(t *testing.T) {
	var buf bytes.Buffer
	store, c := &fakeStore{}, &fakeCommitter{}
	newTestProjector(store, &buf).handleKafkaMessage(context.Background(), c, kafka.Message{Offset: 7, Value: []byte(`{bad`)})
	if store.calls != 0 {
		t.Fatalf("store called for malformed message")
	}
	if c.commits != 1 {
		t.Fatalf("commits = %d, want 1", c.commits)
	}
	if !strings.Contains(buf.String(), "malformed turn message skipped") || !strings.Contains(buf.String(), "{bad") {
		t.Fatalf("expected malformed log with payload: %s", buf.String())
	}
}

func TestDefaultBackoffGrowsAndCaps(t *testing.T) {
	if defaultBackoff(1) != kafkaInitialBackoff || defaultBackoff(2) != 2*kafkaInitialBackoff {
		t.Fatal("backoff should double")
	}
	if defaultBackoff(20) != kafkaMaxBackoff {
		t.Fatal("backoff should cap")
	}
}
