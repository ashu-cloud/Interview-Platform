package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/segmentio/kafka-go"
)

type fakeKafkaWriter struct {
	messages []kafka.Message
	err      error
}

func (w *fakeKafkaWriter) WriteMessages(_ context.Context, messages ...kafka.Message) error {
	w.messages = append(w.messages, messages...)
	return w.err
}
func (*fakeKafkaWriter) Close() error { return nil }

func TestKafkaPublisherCanonicalPayloads(t *testing.T) {
	w := &fakeKafkaWriter{}
	p := &kafkaPublisher{writer: w, turnsTopic: DefaultTurnsTopic, completedTopic: DefaultCompletedTopic}
	if err := p.PublishTurn(context.Background(), "session", "org", 2, RoleAssessor, "question"); err != nil {
		t.Fatal(err)
	}
	if err := p.PublishCompleted(context.Background(), "session", "org", "participant", "blueprint", 2); err != nil {
		t.Fatal(err)
	}
	if len(w.messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(w.messages))
	}
	for i, want := range []struct {
		topic     string
		fields    map[string]any
		timeField string
	}{
		{DefaultTurnsTopic, map[string]any{"event_type": "turn", "session_id": "session", "org_id": "org", "turn_number": float64(2), "role": RoleAssessor, "content": "question"}, "timestamp"},
		{DefaultCompletedTopic, map[string]any{"event_type": "completed", "session_id": "session", "org_id": "org", "participant_id": "participant", "blueprint_id": "blueprint", "turn_count": float64(2)}, "completed_at"},
	} {
		msg := w.messages[i]
		if msg.Topic != want.topic || string(msg.Key) != "session" {
			t.Fatalf("unexpected Kafka routing: topic=%q key=%q", msg.Topic, msg.Key)
		}
		var fields map[string]any
		if err := json.Unmarshal(msg.Value, &fields); err != nil {
			t.Fatal(err)
		}
		if fields[want.timeField] == nil || fields[want.timeField] == "0001-01-01T00:00:00Z" {
			t.Fatalf("missing event time: %v", fields)
		}
		delete(fields, want.timeField)
		if !reflect.DeepEqual(fields, want.fields) {
			t.Fatalf("fields = %v, want %v", fields, want.fields)
		}
	}
}

func TestKafkaPublisherRejectsLegacyRoleAndReturnsWriteError(t *testing.T) {
	w := &fakeKafkaWriter{err: errors.New("broker unavailable")}
	p := &kafkaPublisher{writer: w, turnsTopic: DefaultTurnsTopic}
	if err := p.PublishTurn(context.Background(), "session", "org", 0, "assistant", "question"); err == nil || len(w.messages) != 0 {
		t.Fatalf("legacy role should not be published: error=%v", err)
	}
	if err := p.PublishTurn(context.Background(), "session", "org", 0, RoleAssessor, "question"); !errors.Is(err, w.err) {
		t.Fatalf("write error not propagated: %v", err)
	}
}
