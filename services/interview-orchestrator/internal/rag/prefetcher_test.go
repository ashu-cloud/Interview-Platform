package rag

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestPrefetchKey(t *testing.T) {
	key := prefetchKey("session-123", 2)
	expected := "rag:prefetch:session-123:2"
	if key != expected {
		t.Errorf("expected key %s, got %s", expected, key)
	}
}

func TestFormatChunks(t *testing.T) {
	chunks := []string{"First chunk text", "Second chunk text"}
	formatted := formatChunks(chunks)

	if !strings.Contains(formatted, "[Context from assessment knowledge base]") {
		t.Errorf("expected header in formatted output: %s", formatted)
	}
	if !strings.Contains(formatted, "1. First chunk text") || !strings.Contains(formatted, "2. Second chunk text") {
		t.Errorf("expected numbered chunks in formatted output: %s", formatted)
	}

	empty := formatChunks(nil)
	if empty != "" {
		t.Errorf("expected empty string for nil chunks, got %s", empty)
	}
}

func TestPrefetcher_GracefulDegradationWhenNilOrUnconnected(t *testing.T) {
	// A prefetcher with nil pool or nil redisClient should never crash or block
	var p *Prefetcher
	p.TriggerPrefetch(context.Background(), "s1", "b1", 1, "some interim text")
	ctxVal, err := p.GetContext(context.Background(), "s1", "b1", 1, "final text")
	if err != nil {
		t.Errorf("expected nil error on nil prefetcher, got %v", err)
	}
	if ctxVal != "" {
		t.Errorf("expected empty context on nil prefetcher, got %s", ctxVal)
	}

	unconnected := NewPrefetcher(nil, nil, nil, 100*time.Millisecond)
	unconnected.TriggerPrefetch(context.Background(), "s1", "b1", 1, "some interim text")
	ctxVal2, err2 := unconnected.GetContext(context.Background(), "s1", "b1", 1, "final text")
	if err2 != nil {
		t.Errorf("expected nil error on unconnected prefetcher, got %v", err2)
	}
	if ctxVal2 != "" {
		t.Errorf("expected empty context on unconnected prefetcher, got %s", ctxVal2)
	}
}
