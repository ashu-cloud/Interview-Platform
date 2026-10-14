package projector

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

const (
	outboxBatchSize = 50
)

// TurnProjector projects session turns from outbox_turns into session_turns.
type TurnProjector struct {
	repo         *TurnRepository
	logger       *slog.Logger
	pollInterval time.Duration
}

// NewTurnProjector constructs a TurnProjector for outbox processing.
func NewTurnProjector(repo *TurnRepository, logger *slog.Logger, pollInterval time.Duration) *TurnProjector {
	if logger == nil {
		logger = slog.Default()
	}
	if pollInterval <= 0 {
		pollInterval = 500 * time.Millisecond
	}
	return &TurnProjector{
		repo:         repo,
		logger:       logger,
		pollInterval: pollInterval,
	}
}

// RunOutboxMode runs the outbox polling loop until ctx is cancelled.
func (p *TurnProjector) RunOutboxMode(ctx context.Context) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	// Initial immediate run
	p.processOutboxBatch(ctx)

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
	n, err := p.repo.ProcessOutboxBatch(ctx, outboxBatchSize)
	if err == nil {
		if n > 0 {
			p.logger.Info("outbox turns projected", "count", n)
		}
		return
	}
	if ctx.Err() != nil {
		return
	}
	var pe *ProjectionError
	if errors.As(err, &pe) {
		p.logger.Error("outbox projection failed",
			"outbox_id", pe.OutboxID, "session_id", pe.SessionID,
			"retry_count", pe.RetryCount, "status", pe.Status, "error", pe.Err)
		return
	}
	p.logger.Error("outbox batch failed", "error", err)
}
