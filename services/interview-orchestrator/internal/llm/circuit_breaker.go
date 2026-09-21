package llm

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failures     int
	threshold    int
	openDuration time.Duration
	lastFailure  time.Time
}

func NewCircuitBreaker(threshold int, openDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		threshold:    threshold,
		openDuration: openDuration,
	}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, action func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.openDuration {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	err := action()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()
		if cb.failures >= cb.threshold || cb.state == StateHalfOpen {
			cb.state = StateOpen
		}
		return err
	}

	cb.failures = 0
	cb.state = StateClosed
	return nil
}
