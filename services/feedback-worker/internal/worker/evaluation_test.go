package worker_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/yourusername/ai-interview/services/feedback-worker/internal/llm"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/worker"
)

func TestGroundingValidation_GroundedEvidence(t *testing.T) {
	turns := []worker.SessionTurn{
		{ID: "t1", TurnNumber: 1, Role: "participant",
			Content: "We implemented a sliding-window rate limiter using Redis sorted sets."},
	}
	results := []worker.CriterionResult{
		{CriterionID: "c1", Score: 8, EvidenceTurnID: "t1",
			EvidenceQuote: "sliding-window rate limiter using Redis sorted sets"},
	}
	validated := worker.ValidateGrounding(results, turns)
	if !validated[0].Grounded {
		t.Error("expected grounded=true for verbatim quote from cited turn")
	}
}

func TestGroundingValidation_HallucinatedTurnID(t *testing.T) {
	turns := []worker.SessionTurn{
		{ID: "t1", TurnNumber: 1, Role: "participant", Content: "Some answer."},
	}
	results := []worker.CriterionResult{
		{CriterionID: "c1", Score: 7, EvidenceTurnID: "t999",
			EvidenceQuote: "some quote"},
	}
	validated := worker.ValidateGrounding(results, turns)
	if validated[0].Grounded {
		t.Error("expected grounded=false for non-existent turn_id")
	}
}

func TestGroundingValidation_QuoteNotInCitedTurn(t *testing.T) {
	turns := []worker.SessionTurn{
		{ID: "t1", TurnNumber: 1, Role: "participant",
			Content: "I built a cache layer."},
	}
	results := []worker.CriterionResult{
		{CriterionID: "c1", Score: 6, EvidenceTurnID: "t1",
			EvidenceQuote: "this was never actually said"},
	}
	validated := worker.ValidateGrounding(results, turns)
	if validated[0].Grounded {
		t.Error("expected grounded=false for invented quote")
	}
}

func TestAggregateScores_HireRecommendation(t *testing.T) {
	criteria := []model.EvaluationCriterion{
		{ID: "c1", MaxScore: 10},
		{ID: "c2", MaxScore: 10},
	}
	results := []worker.CriterionResult{
		{CriterionID: "c1", Score: 8},
		{CriterionID: "c2", Score: 6},
	}
	score, rec, err := worker.AggregateScores(results, criteria)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score < 6.9 || score > 7.1 {
		t.Errorf("expected overall score ~7.0, got %v", score)
	}
	if rec != "hire" {
		t.Errorf("expected 'hire', got %s", rec)
	}
}

func TestAggregateScores_ProcessingFailureSeparatedFromScore(t *testing.T) {
	criteria := []model.EvaluationCriterion{
		{ID: "c1", MaxScore: 10},
		{ID: "c2", MaxScore: 10},
	}

	// Case 1: One criterion failed due to model/network error
	resultsPartialFail := []worker.CriterionResult{
		{CriterionID: "c1", Score: 8, Error: nil},
		{CriterionID: "c2", Score: 0, Error: errors.New("provider timeout 504")},
	}
	score, rec, err := worker.AggregateScores(resultsPartialFail, criteria)
	if err == nil {
		t.Fatal("expected non-nil error when criterion evaluation fails")
	}
	if rec != "evaluation_failed" {
		t.Errorf("expected recommendation 'evaluation_failed', got %s", rec)
	}
	if score != 0 {
		t.Errorf("expected score 0 on failure, got %v", score)
	}

	// Case 2: All criteria failed
	resultsAllFail := []worker.CriterionResult{
		{CriterionID: "c1", Score: 0, Error: errors.New("rate limited")},
		{CriterionID: "c2", Score: 0, Error: errors.New("malformed JSON")},
	}
	score, rec, err = worker.AggregateScores(resultsAllFail, criteria)
	if err == nil {
		t.Fatal("expected error when all criteria fail")
	}
	if rec != "evaluation_failed" {
		t.Errorf("expected 'evaluation_failed', got %s (must NOT silently become 'no_hire')", rec)
	}
	if score != 0 {
		t.Errorf("expected 0, got %v", score)
	}
}

func TestAggregateScores_ValidCandidateZeroScore(t *testing.T) {
	criteria := []model.EvaluationCriterion{
		{ID: "c1", MaxScore: 10},
		{ID: "c2", MaxScore: 10},
	}
	// Candidate was asked questions but gave completely wrong/zero answers, cleanly evaluated
	resultsZero := []worker.CriterionResult{
		{CriterionID: "c1", Score: 0, Error: nil},
		{CriterionID: "c2", Score: 0, Error: nil},
	}
	score, rec, err := worker.AggregateScores(resultsZero, criteria)
	if err != nil {
		t.Fatalf("genuine candidate zero should NOT produce an error, got: %v", err)
	}
	if score != 0.0 {
		t.Errorf("expected candidate score 0.0, got %v", score)
	}
	if rec != "strong_no_hire" {
		t.Errorf("expected candidate recommendation 'strong_no_hire', got %s", rec)
	}
}

func TestParseCriterionResponse_Variations(t *testing.T) {
	// 1. Valid JSON with markdown wrapping
	wrapped := "```json\n{\"criterion_id\": \"c1\", \"score\": 8, \"evidence_turn_id\": \"t1\", \"evidence_quote\": \"Redis SETNX lock\"}\n```"
	resp, err := llm.ParseCriterionResponse(wrapped)
	if err != nil {
		t.Fatalf("failed to parse markdown-wrapped json: %v", err)
	}
	if resp.CriterionID != "c1" || resp.Score != 8 {
		t.Errorf("unexpected parsed response: %+v", resp)
	}

	// 2. Recoverable truncation (missing closing quote and closing brace)
	truncatedRecoverable := `{"criterion_id": "c1", "score": 9, "evidence_turn_id": "t1", "evidence_quote": "We used PostgreSQL outbox`
	resp, err = llm.ParseCriterionResponse(truncatedRecoverable)
	if err != nil {
		t.Fatalf("expected bounded recovery for truncated quote, got: %v", err)
	}
	if resp.Score != 9 || resp.CriterionID != "c1" {
		t.Errorf("unexpected recovered response: %+v", resp)
	}

	// 3. Unrecoverable truncation (cut off before score)
	truncatedUnrecoverable := `{"criterion_id": "c1", "sco`
	_, err = llm.ParseCriterionResponse(truncatedUnrecoverable)
	if err == nil || !errors.Is(err, llm.ErrTruncatedOutput) && !errors.Is(err, llm.ErrMalformedOutput) {
		t.Errorf("expected truncated error, got: %v", err)
	}

	// 4. Missing required field (missing score)
	missingScore := `{"criterion_id": "c1", "evidence_turn_id": "t1", "evidence_quote": "some text"}`
	_, err = llm.ParseCriterionResponse(missingScore)
	if err == nil || !errors.Is(err, llm.ErrMissingRequiredFields) {
		t.Errorf("expected ErrMissingRequiredFields for missing score, got: %v", err)
	}

	// 5. Completely invalid JSON
	garbage := `I cannot evaluate this candidate because the text is unclear.`
	_, err = llm.ParseCriterionResponse(garbage)
	if err == nil || !errors.Is(err, llm.ErrMalformedOutput) {
		t.Errorf("expected ErrMalformedOutput for non-JSON text, got: %v", err)
	}

	// 6. Valid zero score output
	zeroValid := `{"criterion_id": "c1", "score": 0, "evidence_turn_id": "t1", "evidence_quote": "no relevant experience"}`
	resp, err = llm.ParseCriterionResponse(zeroValid)
	if err != nil {
		t.Fatalf("valid zero score failed to parse: %v", err)
	}
	if resp.Score != 0 || resp.CriterionID != "c1" {
		t.Errorf("expected score 0 and c1, got %+v", resp)
	}
}

func TestValidateTranscriptCompleteness(t *testing.T) {
	// Empty transcript
	err := worker.ValidateTranscriptCompleteness(nil, 3)
	if err == nil || !errors.Is(err, worker.ErrEmptyTranscript) {
		t.Errorf("expected ErrEmptyTranscript, got %v", err)
	}

	// Missing turns: expected 4 turns, but turn 2 is missing
	incompleteTurns := []worker.SessionTurn{
		{ID: "t1", TurnNumber: 1, Role: "assessor", Content: "Hello"},
		{ID: "t3", TurnNumber: 3, Role: "assessor", Content: "Next question"},
		{ID: "t4", TurnNumber: 4, Role: "participant", Content: "Answer"},
	}
	err = worker.ValidateTranscriptCompleteness(incompleteTurns, 4)
	if err == nil || !errors.Is(err, worker.ErrIncompleteTranscript) {
		t.Errorf("expected ErrIncompleteTranscript for missing turn 2, got %v", err)
	}

	// Complete transcript: 1..4 all present
	completeTurns := []worker.SessionTurn{
		{ID: "t1", TurnNumber: 1, Role: "assessor", Content: "Hello"},
		{ID: "t2", TurnNumber: 2, Role: "participant", Content: "Ready"},
		{ID: "t3", TurnNumber: 3, Role: "assessor", Content: "Next question"},
		{ID: "t4", TurnNumber: 4, Role: "participant", Content: "Answer"},
	}
	if err := worker.ValidateTranscriptCompleteness(completeTurns, 4); err != nil {
		t.Errorf("expected complete transcript to succeed, got: %v", err)
	}
}

func TestFormatTranscriptForCriterion(t *testing.T) {
	turns := []worker.SessionTurn{
		{ID: "t1", TurnNumber: 1, Role: "participant", Content: "Hello."},
		{ID: "t2", TurnNumber: 1, Role: "assessor", Content: "Great."},
	}
	formatted := worker.FormatTranscriptForCriterion(turns)
	if !strings.Contains(formatted, "Turn 1 [participant] (id:t1): Hello.") {
		t.Error("missing participant turn with ID in formatted transcript")
	}
	if !strings.Contains(formatted, "Turn 1 [assessor] (id:t2)") {
		t.Error("missing assessor turn with ID in formatted transcript")
	}
}

func TestGroundingHeldOutCases(t *testing.T) {
	heldOutTurns := []worker.SessionTurn{
		{
			ID:         "turn-101",
			TurnNumber: 1,
			Role:       "participant",
			Content:    "We configured pgxpool with MaxConns=25 and MinConns=5 to avoid connection starvation under traffic spikes.",
		},
		{
			ID:         "turn-102",
			TurnNumber: 2,
			Role:       "participant",
			Content:    "During audio streaming, our Web Audio ScriptProcessor uses 16kHz linear PCM with an adaptive noise floor calculation.",
		},
	}

	testCases := []struct {
		name        string
		turnID      string
		quote       string
		expectMatch bool
		reason      string
	}{
		{
			name:        "verbatim_with_smart_quotes",
			turnID:      "turn-101",
			quote:       "“MaxConns=25 and MinConns=5”",
			expectMatch: true,
			reason:      "smart quotes should be normalized to match verbatim text",
		},
		{
			name:        "verbatim_with_boundary_punctuation",
			turnID:      "turn-101",
			quote:       "avoid connection starvation under traffic spikes.",
			expectMatch: true,
			reason:      "trailing period should not prevent substring match",
		},
		{
			name:        "verbatim_multiline_whitespace_collapse",
			turnID:      "turn-102",
			quote:       "Web Audio   ScriptProcessor \n  uses 16kHz",
			expectMatch: true,
			reason:      "internal whitespace should be collapsed cleanly",
		},
		{
			name:        "short_stop_words_rejected",
			turnID:      "turn-101",
			quote:       "with and to",
			expectMatch: false,
			reason:      "non-substantive stop-tokens must be rejected",
		},
		{
			name:        "very_short_phrase_rejected",
			turnID:      "turn-102",
			quote:       "with an",
			expectMatch: false,
			reason:      "short phrase failing substantive threshold must be rejected",
		},
		{
			name:        "wrong_turn_citation_rejected",
			turnID:      "turn-102",
			quote:       "avoid connection starvation under traffic spikes",
			expectMatch: false,
			reason:      "quote actually occurred in turn-101, not cited turn-102",
		},
		{
			name:        "fabricated_hallucination_rejected",
			turnID:      "turn-101",
			quote:       "We used Apache Cassandra for high write throughput across three datacenters",
			expectMatch: false,
			reason:      "hallucinated claim never spoken by candidate",
		},
		{
			name:        "semantic_paraphrase_rejected_by_verbatim_contract",
			turnID:      "turn-102",
			quote:       "The microphone downsamples audio to 16000Hz mono before sending",
			expectMatch: false,
			reason:      "paraphrase is rejected by contract; verbatim required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := []worker.CriterionResult{
				{
					CriterionID:    "crit-heldout",
					Score:          8,
					EvidenceTurnID: tc.turnID,
					EvidenceQuote:  tc.quote,
				},
			}
			out := worker.ValidateGrounding(res, heldOutTurns)
			if out[0].Grounded != tc.expectMatch {
				t.Errorf("[%s] expected Grounded=%v, got %v (%s)", tc.name, tc.expectMatch, out[0].Grounded, tc.reason)
			}
		})
	}
}
