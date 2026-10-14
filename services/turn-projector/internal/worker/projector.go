package worker

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yourusername/ai-interview/services/turn-projector/internal/config"
	"github.com/yourusername/ai-interview/services/turn-projector/internal/model"
	"github.com/yourusername/ai-interview/services/turn-projector/internal/repository"
)

const (
	outboxBatchSize     = 50
	kafkaMaxAttempts    = 5
	kafkaInitialBackoff = 200 * time.Millisecond
	kafkaMaxBackoff     = 5 * time.Second
	maxLoggedPayload    = 512
)

// turnStore is the persistence surface the projector needs in Kafka mode.
type turnStore interface {
	InsertTurn(ctx context.Context, turn model.SessionTurnEvent) error
}

// outboxStore is the persistence surface the projector needs in outbox mode.
type outboxStore interface {
	ProcessOutboxBatch(ctx context.Context, limit int) (int, error)
}

// committer commits Kafka messages (satisfied by *kafka.Reader).
type committer interface {
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}

// TurnProjector projects session turns into session_turns from Kafka or the Postgres outbox.
type TurnProjector struct {
	cfg    *config.Config
	turns  turnStore
	outbox outboxStore
	logger *slog.Logger

	// Overridable for tests.
	runKafka  func(ctx context.Context)
	runOutbox func(ctx context.Context)
	backoff   func(attempt int) time.Duration
}

// NewTurnProjector constructs a TurnProjector.
func NewTurnProjector(cfg *config.Config, repo *repository.TurnRepository, logger *slog.Logger) *TurnProjector {
	p := &TurnProjector{cfg: cfg, turns: repo, outbox: repo, logger: logger}
	p.runKafka = p.runKafkaMode
	p.runOutbox = p.runOutboxMode
	p.backoff = defaultBackoff
	return p
}

// Run blocks until ctx is cancelled. Mode is chosen from configuration.
func (p *TurnProjector) Run(ctx context.Context) {
	switch p.cfg.Mode() {
	case config.ModeKafka:
		p.logger.Info("turn-projector: Kafka mode", "topic", p.cfg.KafkaTopic, "group_id", p.cfg.KafkaGroupID)
		p.runKafka(ctx)
	default:
		p.logger.Info("turn-projector: Postgres outbox mode", "poll_interval", p.cfg.PollInterval.String())
		p.runOutbox(ctx)
	}
}

func (p *TurnProjector) runKafkaMode(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(p.cfg.KafkaBrokers, ","),
		Topic:    p.cfg.KafkaTopic,
		GroupID:  p.cfg.KafkaGroupID,
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  1 * time.Second,
	})
	defer r.Close()

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			p.logger.Error("kafka fetch failed", "error", err)
			if !sleepCtx(ctx, time.Second) {
				return
			}
			continue
		}
		p.handleKafkaMessage(ctx, r, m)
	}
}

// handleKafkaMessage projects one message with bounded retry/backoff. Malformed or
// permanently failing messages are logged with their coordinates and then committed
// so one bad message cannot block the partition forever.
func (p *TurnProjector) handleKafkaMessage(ctx context.Context, c committer, m kafka.Message) {
	attrs := []any{"topic", m.Topic, "partition", m.Partition, "offset", m.Offset}

	event, err := model.ParseTurnEvent(m.Value)
	if err != nil {
		p.logger.Error("malformed turn message skipped",
			append(attrs, "key", string(m.Key), "payload", truncate(m.Value), "error", err)...)
		p.commit(ctx, c, m)
		return
	}

	var lastErr error
	for attempt := 1; attempt <= kafkaMaxAttempts; attempt++ {
		if lastErr = p.turns.InsertTurn(ctx, event); lastErr == nil {
			p.commit(ctx, c, m)
			return
		}
		if ctx.Err() != nil {
			return // shutting down; leave uncommitted so the message is redelivered
		}
		p.logger.Warn("turn projection failed",
			append(attrs, "session_id", event.SessionID, "turn_number", event.TurnNumber,
				"attempt", attempt, "max_attempts", kafkaMaxAttempts, "error", lastErr)...)
		if attempt < kafkaMaxAttempts && !sleepCtx(ctx, p.backoff(attempt)) {
			return
		}
	}
	p.logger.Error("turn projection failed permanently, message skipped",
		append(attrs, "session_id", event.SessionID, "turn_number", event.TurnNumber,
			"role", event.Role, "error", lastErr)...)
	p.commit(ctx, c, m)
}

func (p *TurnProjector) commit(ctx context.Context, c committer, m kafka.Message) {
	if err := c.CommitMessages(ctx, m); err != nil && ctx.Err() == nil {
		p.logger.Error("kafka commit failed", "partition", m.Partition, "offset", m.Offset, "error", err)
	}
}

func (p *TurnProjector) runOutboxMode(ctx context.Context) {
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.processOutboxBatch(ctx)
		}
	}
}

func (p *TurnProjector) processOutboxBatch(ctx context.Context) {
	n, err := p.outbox.ProcessOutboxBatch(ctx, outboxBatchSize)
	if err == nil {
		if n > 0 {
			p.logger.Info("outbox turns projected", "count", n)
		}
		return
	}
	if ctx.Err() != nil {
		return
	}
	var pe *repository.ProjectionError
	if errors.As(err, &pe) {
		p.logger.Error("outbox projection failed",
			"outbox_id", pe.OutboxID, "session_id", pe.SessionID,
			"retry_count", pe.RetryCount, "status", pe.Status, "error", pe.Err)
		return
	}
	p.logger.Error("outbox batch failed", "error", err)
}

func defaultBackoff(attempt int) time.Duration {
	d := kafkaInitialBackoff << (attempt - 1)
	if d > kafkaMaxBackoff {
		return kafkaMaxBackoff
	}
	return d
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func truncate(b []byte) string {
	if len(b) > maxLoggedPayload {
		return string(b[:maxLoggedPayload]) + "…"
	}
	return string(b)
}
