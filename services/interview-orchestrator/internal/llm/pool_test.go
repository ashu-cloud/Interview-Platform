package llm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

type mockProvider struct {
	err         error
	chunksToEmit []string
	delay       time.Duration
}

func (m *mockProvider) GenerateResponse(ctx context.Context, systemPrompt string, messages []model.Message, streamChan chan<- string) error {
	for _, chunk := range m.chunksToEmit {
		if m.delay > 0 {
			time.Sleep(m.delay)
		}
		streamChan <- chunk
	}
	return m.err
}

func TestPoolFailover_PrimaryFailsBeforeOutput(t *testing.T) {
	primary := &mockProvider{
		err:         errors.New("primary failed"),
		chunksToEmit: []string{}, // fails before output
	}
	secondary := &mockProvider{
		err:         nil,
		chunksToEmit: []string{"secondary", " ", "works"},
	}

	pool := NewPool(primary, secondary)
	streamChan := make(chan string, 10)

	var output string
	go func() {
		for chunk := range streamChan {
			output += chunk
		}
	}()

	err := pool.GenerateResponse(context.Background(), "sys", nil, streamChan)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	close(streamChan)
	time.Sleep(50 * time.Millisecond)

	if output != "secondary works" {
		t.Errorf("expected 'secondary works', got %q", output)
	}
}

func TestPoolFailover_PrimaryFailsAfterOutput(t *testing.T) {
	primary := &mockProvider{
		err:         errors.New("primary failed"),
		chunksToEmit: []string{"primary", " ", "started"},
	}
	secondary := &mockProvider{
		err:         nil,
		chunksToEmit: []string{"secondary", " ", "works"},
	}

	pool := NewPool(primary, secondary)
	streamChan := make(chan string, 10)

	var output string
	go func() {
		for chunk := range streamChan {
			output += chunk
		}
	}()

	err := pool.GenerateResponse(context.Background(), "sys", nil, streamChan)
	if err == nil {
		t.Fatalf("expected error because primary failed mid-stream, got nil")
	}

	close(streamChan)
	time.Sleep(50 * time.Millisecond)

	if output != "primary started" {
		t.Errorf("expected only 'primary started', got %q", output)
	}
}
