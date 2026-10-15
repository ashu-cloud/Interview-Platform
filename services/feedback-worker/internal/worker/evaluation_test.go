package worker_test

import (
	"strings"
	"testing"

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
	score, rec := worker.AggregateScores(results, criteria)
	if score < 6.9 || score > 7.1 {
		t.Errorf("expected overall score ~7.0, got %v", score)
	}
	if rec != "hire" {
		t.Errorf("expected 'hire', got %s", rec)
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
