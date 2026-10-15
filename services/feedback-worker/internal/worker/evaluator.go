package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/yourusername/ai-interview/services/feedback-worker/internal/llm"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
)

type CriterionResult struct {
	CriterionID    string
	Score          int
	EvidenceTurnID string
	EvidenceQuote  string
	Grounded       bool
	Error          error
}

type Evaluator struct {
	llmClient llm.EvaluationLLM
	logger    *slog.Logger
}

func NewEvaluator(llmClient llm.EvaluationLLM, logger *slog.Logger) *Evaluator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Evaluator{
		llmClient: llmClient,
		logger:    logger,
	}
}

// FormatTranscriptForCriterion formats session turns for the per-criterion prompt.
func FormatTranscriptForCriterion(turns []SessionTurn) string {
	var sb strings.Builder
	for _, t := range turns {
		sb.WriteString(fmt.Sprintf("Turn %d [%s] (id:%s): %s\n",
			t.TurnNumber, t.Role, t.ID, t.Content))
	}
	return sb.String()
}

func (e *Evaluator) EvaluateCriteria(ctx context.Context,
	criteria []model.EvaluationCriterion,
	turns []SessionTurn,
	participantName string,
	maxConcurrent int) []CriterionResult {

	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}

	sem := make(chan struct{}, maxConcurrent)
	results := make([]CriterionResult, len(criteria))
	var wg sync.WaitGroup

	for i, criterion := range criteria {
		wg.Add(1)
		go func(idx int, c model.EvaluationCriterion) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[idx] = CriterionResult{
					CriterionID: c.ID,
					Error:       ctx.Err(),
				}
				return
			}
			results[idx] = e.evaluateOneCriterion(ctx, c, turns, participantName)
		}(i, criterion)
	}
	wg.Wait()
	return results
}

func (e *Evaluator) evaluateOneCriterion(ctx context.Context,
	c model.EvaluationCriterion,
	turns []SessionTurn,
	participantName string) CriterionResult {

	transcriptFormatted := FormatTranscriptForCriterion(turns)
	prompt := fmt.Sprintf(`You are evaluating ONE specific competency in a completed technical assessment.

PARTICIPANT: %s

COMPETENCY: %s
DESCRIPTION: %s
MAX SCORE: %d

FULL ASSESSMENT TRANSCRIPT:
%s
Respond with EXACTLY this JSON and nothing else:
{
  "criterion_id": "%s",
  "score": <integer 0 to %d>,
  "evidence_turn_id": "<UUID of the specific turn that best demonstrates your score>",
  "evidence_quote": "<exact verbatim quote from that turn, 10-80 words>"
}

Rules:
- score MUST be an integer between 0 and %d inclusive
- evidence_turn_id MUST be a UUID that appears in the transcript above
- evidence_quote MUST be copied verbatim from the content of that turn
- Do not invent or paraphrase the evidence quote
- If the participant never addressed this competency, score 0 and use the first participant turn`,
		participantName, c.Name, c.Description, c.MaxScore,
		transcriptFormatted,
		c.ID, c.MaxScore, c.MaxScore,
	)

	resp, err := e.llmClient.EvaluateCriterion(ctx, prompt)
	if err != nil {
		e.logger.Error("evaluate criterion failed", "criterion_id", c.ID, "error", err)
		return CriterionResult{
			CriterionID: c.ID,
			Score:       0,
			Error:       err,
		}
	}

	score := resp.Score
	if score < 0 {
		score = 0
	}
	if score > c.MaxScore {
		score = c.MaxScore
	}

	evidenceTurnID := resp.EvidenceTurnID
	evidenceQuote := resp.EvidenceQuote

	// If score is 0 and no turn was specified, fallback to first participant turn
	if (evidenceTurnID == "" || evidenceQuote == "") && len(turns) > 0 {
		for _, t := range turns {
			if t.Role == "participant" {
				evidenceTurnID = t.ID
				evidenceQuote = t.Content
				break
			}
		}
		if evidenceTurnID == "" {
			evidenceTurnID = turns[0].ID
			evidenceQuote = turns[0].Content
		}
	}

	return CriterionResult{
		CriterionID:    c.ID,
		Score:          score,
		EvidenceTurnID: evidenceTurnID,
		EvidenceQuote:  evidenceQuote,
	}
}
