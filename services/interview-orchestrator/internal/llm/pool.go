package llm

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

type Provider interface {
	GenerateResponse(ctx context.Context, systemPrompt string, messages []model.Message, streamChan chan<- string) error
}

type Pool struct {
	primary   Provider
	secondary Provider
	cbPrimary *CircuitBreaker
	cbSecond  *CircuitBreaker
}

func NewPool(primary, secondary Provider) *Pool {
	return &Pool{
		primary:   primary,
		secondary: secondary,
		// Threshold: 5 failures, Open duration: 30 seconds (from spec)
		cbPrimary: NewCircuitBreaker(5, 30*time.Second),
		cbSecond:  NewCircuitBreaker(5, 30*time.Second),
	}
}

func (p *Pool) GenerateResponse(ctx context.Context, systemPrompt string, messages []model.Message, streamChan chan<- string) error {
	var streamed atomic.Bool
	primaryChan := make(chan string, 100)
	done := make(chan struct{})

	go func() {
		for chunk := range primaryChan {
			streamed.Store(true)
			streamChan <- chunk
		}
		close(done)
	}()

	// 1. Try primary
	err := p.cbPrimary.Execute(ctx, func() error {
		return p.primary.GenerateResponse(ctx, systemPrompt, messages, primaryChan)
	})
	
	close(primaryChan)
	<-done // ensure all chunks are flushed to streamChan

	if err == nil {
		return nil
	}

	// If primary emitted tokens before failing, do not failover to avoid duplicating/appending responses.
	if streamed.Load() {
		return fmt.Errorf("primary LLM failed after partial stream: %w", err)
	}

	// 2. Try secondary if primary fails or circuit is open, and no tokens were emitted
	errSec := p.cbSecond.Execute(ctx, func() error {
		return p.secondary.GenerateResponse(ctx, systemPrompt, messages, streamChan)
	})

	if errSec == nil {
		return nil
	}

	return fmt.Errorf("all LLM providers failed: primary=%v, secondary=%v", err, errSec)
}
