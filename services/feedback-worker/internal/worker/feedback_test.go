package worker

import (
	"strings"
	"testing"

	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
)

func TestTranscriptReconstructionIncludesBothSidesOfEveryTurn(t *testing.T) {
	events := []model.TurnEvent{
		{InterviewID: "target", TurnNumber: 1, Role: "user", Content: "I built the service."},
		{InterviewID: "target", TurnNumber: 1, Role: "assistant", Content: "How did you test it?"},
		{InterviewID: "target", TurnNumber: 2, Role: "user", Content: "With integration tests."},
		{InterviewID: "target", TurnNumber: 2, Role: "assistant", Content: "What did they cover?"},
	}

	expected := expectedTranscriptEventCount(2)
	if expected != len(events) {
		t.Fatalf("expected %d transcript events, got %d", expected, len(events))
	}

	transcript := formatTranscript(events)
	for _, content := range []string{
		"I built the service.",
		"How did you test it?",
		"With integration tests.",
		"What did they cover?",
	} {
		if !strings.Contains(transcript, content) {
			t.Errorf("transcript is missing %q: %q", content, transcript)
		}
	}
}
