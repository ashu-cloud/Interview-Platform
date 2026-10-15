package worker

import (
	"strings"

	"github.com/yourusername/ai-interview/services/feedback-worker/internal/repository"
)

type SessionTurn = repository.SessionTurn

// ValidateGrounding checks each result's evidence_quote exists verbatim in the cited turn.
// Sets result.Grounded accordingly. Never fails the evaluation job.
func ValidateGrounding(results []CriterionResult, turns []SessionTurn) []CriterionResult {
	turnContent := make(map[string]string, len(turns))
	for _, t := range turns {
		turnContent[t.ID] = t.Content
	}

	for i, r := range results {
		if r.Error != nil {
			results[i].Grounded = false
			continue
		}
		content, ok := turnContent[r.EvidenceTurnID]
		if !ok {
			results[i].Grounded = false
			continue
		}

		quoteNorm := strings.ToLower(strings.Join(strings.Fields(r.EvidenceQuote), " "))
		if quoteNorm == "" {
			results[i].Grounded = false
			continue
		}

		results[i].Grounded = strings.Contains(
			strings.ToLower(strings.Join(strings.Fields(content), " ")),
			quoteNorm,
		)
	}
	return results
}
