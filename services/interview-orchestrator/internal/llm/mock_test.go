package llm

import (
	"context"
	"strings"
	"testing"

	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

func TestDemoProvider_WithDocketQuestions(t *testing.T) {
	provider := NewDemoProvider()
	prompt := `You are conducting a professional job interview.

Instructions:
Evaluate candidate's backend knowledge.

Interview Questions to cover:
1. Explain how goroutines communicate safely in Go.
2. How do you design an idempotent payment API?

CRITICAL: Never use markdown formatting.`

	// Turn 1 (greeting)
	ch1 := make(chan string, 1)
	err := provider.GenerateResponse(context.Background(), prompt, []model.Message{
		{Role: "user", Content: "You are conducting an interview."},
	}, ch1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp1 := <-ch1
	if !strings.Contains(resp1, "goroutines communicate safely") {
		t.Errorf("expected response to contain question 1, got: %s", resp1)
	}

	// Turn 2 (candidate answered question 1)
	ch2 := make(chan string, 1)
	err = provider.GenerateResponse(context.Background(), prompt, []model.Message{
		{Role: "user", Content: "You are conducting an interview."},
		{Role: "assistant", Content: resp1},
		{Role: "user", Content: "We use channels and sync package with mutexes."},
	}, ch2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp2 := <-ch2
	if !strings.Contains(resp2, "idempotent payment API") {
		t.Errorf("expected response to contain question 2, got: %s", resp2)
	}
}
