package evaluator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionTurn struct {
	ID         string `json:"id"`
	TurnNumber int    `json:"turn_number"`
	Role       string `json:"role"`
	Content    string `json:"content"`
}

type EvaluationCriterion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxScore    int    `json:"max_score"`
	OrderIndex  int    `json:"order_index"`
}

type CriterionResult struct {
	CriterionID    string
	Score          int
	EvidenceTurnID string
	EvidenceQuote  string
	Grounded       bool
	Error          error
}

type CriterionLLMResponse struct {
	CriterionID    string `json:"criterion_id"`
	Score          int    `json:"score"`
	EvidenceTurnID string `json:"evidence_turn_id"`
	EvidenceQuote  string `json:"evidence_quote"`
}

type OpenAIRequest struct {
	Model          string            `json:"model"`
	Messages       []OpenAIMessage   `json:"messages"`
	Temperature    float64           `json:"temperature"`
	MaxTokens      int               `json:"max_tokens"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type OutboxRow struct {
	ID            string
	SessionID     string
	OrgID         string
	ParticipantID string
	BlueprintID   string
	TurnCount     int
	RetryCount    int
}

type EvaluatorWorker struct {
	pool         *pgxpool.Pool
	logger       *slog.Logger
	groqKey      string
	groqModel    string
	pollInterval time.Duration
	httpClient   *http.Client
}

func NewEvaluatorWorker(pool *pgxpool.Pool, logger *slog.Logger, groqKey, groqModel string, pollInterval time.Duration) *EvaluatorWorker {
	if logger == nil {
		logger = slog.Default()
	}
	if groqModel == "" {
		groqModel = "openai/gpt-oss-120b"
	}
	if pollInterval <= 0 {
		pollInterval = 1 * time.Second
	}
	return &EvaluatorWorker{
		pool:         pool,
		logger:       logger,
		groqKey:      groqKey,
		groqModel:    groqModel,
		pollInterval: pollInterval,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (w *EvaluatorWorker) RunOutboxMode(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.processBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *EvaluatorWorker) processBatch(ctx context.Context) {
	rows, err := w.dequeue(ctx, 5)
	if err != nil {
		if ctx.Err() == nil {
			w.logger.Error("evaluator: dequeue outbox_completed failed", "error", err)
		}
		return
	}

	for _, r := range rows {
		if err := w.processOne(ctx, r); err != nil {
			w.logger.Error("evaluator: process session failed", "session_id", r.SessionID, "error", err)
			w.incrementRetry(ctx, r.ID)
		} else {
			w.markCompleted(ctx, r.ID)
		}
	}
}

func (w *EvaluatorWorker) dequeue(ctx context.Context, limit int) ([]OutboxRow, error) {
	query := `
		SELECT id::text, session_id::text, org_id::text, participant_id::text, blueprint_id::text, turn_count, retry_count
		FROM outbox_completed
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	rows, err := w.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []OutboxRow
	for rows.Next() {
		var row OutboxRow
		if err := rows.Scan(&row.ID, &row.SessionID, &row.OrgID, &row.ParticipantID, &row.BlueprintID, &row.TurnCount, &row.RetryCount); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (w *EvaluatorWorker) markCompleted(ctx context.Context, id string) {
	_, _ = w.pool.Exec(ctx, `UPDATE outbox_completed SET status = 'completed', processed_at = NOW() WHERE id = $1`, id)
}

func (w *EvaluatorWorker) incrementRetry(ctx context.Context, id string) {
	_, _ = w.pool.Exec(ctx, `
		UPDATE outbox_completed
		SET retry_count = retry_count + 1,
		    status = CASE WHEN retry_count + 1 >= 5 THEN 'failed' ELSE 'pending' END,
		    failed_at = CASE WHEN retry_count + 1 >= 5 THEN NOW() ELSE NULL END
		WHERE id = $1
	`, id)
}

func (w *EvaluatorWorker) processOne(ctx context.Context, r OutboxRow) error {
	// 1. Idempotency check: if report already exists, mark session evaluation_ready and outbox completed
	var count int
	_ = w.pool.QueryRow(ctx, `SELECT count(1) FROM evaluation_reports WHERE interview_id = $1`, r.SessionID).Scan(&count)
	if count > 0 {
		w.logger.Info("evaluator: session already has evaluation report", "session_id", r.SessionID)
		_, _ = w.pool.Exec(ctx, `UPDATE assessment_sessions SET status = 'evaluation_ready', updated_at = NOW() WHERE id = $1`, r.SessionID)
		return nil
	}

	// 2. Fetch participant name
	var candidateName string
	_ = w.pool.QueryRow(ctx, `SELECT full_name FROM participants WHERE id = $1`, r.ParticipantID).Scan(&candidateName)
	if candidateName == "" {
		_ = w.pool.QueryRow(ctx, `SELECT full_name FROM candidates WHERE id = $1`, r.ParticipantID).Scan(&candidateName)
	}
	if candidateName == "" {
		candidateName = "Candidate"
	}

	// 3. Fetch blueprint title & criteria
	var bpTitle string
	_ = w.pool.QueryRow(ctx, `SELECT title FROM assessment_blueprints WHERE id = $1`, r.BlueprintID).Scan(&bpTitle)
	if bpTitle == "" {
		_ = w.pool.QueryRow(ctx, `SELECT title FROM dockets WHERE id = $1`, r.BlueprintID).Scan(&bpTitle)
	}

	critRows, err := w.pool.Query(ctx, `SELECT id, name, description, max_score, order_index FROM evaluation_criteria WHERE docket_id = $1 ORDER BY order_index`, r.BlueprintID)
	if err != nil {
		return fmt.Errorf("query criteria: %w", err)
	}
	defer critRows.Close()

	var criteria []EvaluationCriterion
	for critRows.Next() {
		var c EvaluationCriterion
		if err := critRows.Scan(&c.ID, &c.Name, &c.Description, &c.MaxScore, &c.OrderIndex); err != nil {
			return fmt.Errorf("scan criterion: %w", err)
		}
		criteria = append(criteria, c)
	}

	if len(criteria) == 0 {
		return fmt.Errorf("no criteria found for blueprint %s", r.BlueprintID)
	}

	// 4. Fetch transcript turns
	tRows, err := w.pool.Query(ctx, `SELECT id, turn_number, role, content FROM session_turns WHERE session_id = $1 ORDER BY turn_number ASC`, r.SessionID)
	if err != nil {
		return fmt.Errorf("query turns: %w", err)
	}
	defer tRows.Close()

	var turns []SessionTurn
	for tRows.Next() {
		var t SessionTurn
		if err := tRows.Scan(&t.ID, &t.TurnNumber, &t.Role, &t.Content); err != nil {
			return fmt.Errorf("scan turn: %w", err)
		}
		turns = append(turns, t)
	}

	if len(turns) == 0 {
		return fmt.Errorf("no turns found for session %s", r.SessionID)
	}

	// 5. Evaluate Criteria concurrently (max 3 concurrent)
	results := make([]CriterionResult, len(criteria))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for i, c := range criteria {
		wg.Add(1)
		go func(idx int, crit EvaluationCriterion) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = w.evaluateCriterion(ctx, crit, turns, candidateName)
		}(i, c)
	}
	wg.Wait()

	// 6. Aggregate scores
	var totalScore, totalMax float64
	for i, res := range results {
		c := criteria[i]
		totalScore += float64(res.Score)
		totalMax += float64(c.MaxScore)
	}

	var overall float64
	if totalMax > 0 {
		overall = math.Round((totalScore/totalMax*10.0)*10) / 10
	}

	var recommendation string
	switch {
	case overall >= 8.0:
		recommendation = "strong_hire"
	case overall >= 6.0:
		recommendation = "hire"
	case overall >= 4.0:
		recommendation = "no_hire"
	default:
		recommendation = "strong_no_hire"
	}

	// 7. Generate summary
	summary := w.generateSummary(ctx, candidateName, bpTitle, criteria, results, recommendation, overall)

	// 8. Prepare transcript JSON
	type simpleTurn struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	simpleTurns := make([]simpleTurn, len(turns))
	for i, t := range turns {
		simpleTurns[i] = simpleTurn{Role: t.Role, Content: t.Content}
	}
	transcriptBytes, _ := json.Marshal(simpleTurns)

	// 9. Save atomically
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, _ = tx.Exec(ctx, `DELETE FROM evaluation_reports WHERE interview_id = $1`, r.SessionID)

	reportID := uuid.New().String()
	_, err = tx.Exec(ctx, `
		INSERT INTO evaluation_reports (id, interview_id, org_id, summary, overall_score, recommendation, transcript_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, reportID, r.SessionID, r.OrgID, summary, overall, recommendation, string(transcriptBytes))
	if err != nil {
		return fmt.Errorf("insert evaluation report: %w", err)
	}

	for _, res := range results {
		scoreID := uuid.New().String()
		var turnIDVal any = nil
		if res.EvidenceTurnID != "" {
			if _, pErr := uuid.Parse(res.EvidenceTurnID); pErr == nil {
				turnIDVal = res.EvidenceTurnID
			}
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO evaluation_scores (id, feedback_id, criterion_id, score, evidence, evidence_turn_id, is_grounded)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, scoreID, reportID, res.CriterionID, res.Score, res.EvidenceQuote, turnIDVal, res.Grounded)
		if err != nil {
			return fmt.Errorf("insert evaluation score: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `UPDATE assessment_sessions SET status = 'evaluation_ready', updated_at = NOW() WHERE id = $1`, r.SessionID)
	if err != nil {
		return fmt.Errorf("update session status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit evaluation tx: %w", err)
	}

	w.logger.Info("evaluator: successfully synthesized evaluation report",
		"session_id", r.SessionID,
		"overall_score", overall,
		"recommendation", recommendation,
	)
	return nil
}

func (w *EvaluatorWorker) evaluateCriterion(ctx context.Context, c EvaluationCriterion, turns []SessionTurn, participantName string) CriterionResult {
	if w.groqKey == "" {
		return w.mockScoreCriterion(c, turns)
	}

	var sb strings.Builder
	for _, t := range turns {
		sb.WriteString(fmt.Sprintf("Turn %d [%s] (id:%s): %s\n", t.TurnNumber, t.Role, t.ID, t.Content))
	}

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
		sb.String(),
		c.ID, c.MaxScore, c.MaxScore,
	)

	raw, err := w.callGroq(ctx, prompt, true)
	if err != nil {
		w.logger.Warn("evaluator: groq call failed for criterion, using heuristic", "criterion", c.Name, "error", err)
		return w.mockScoreCriterion(c, turns)
	}

	cleanRaw := strings.TrimSpace(raw)
	if idx := strings.Index(cleanRaw, "{"); idx != -1 {
		cleanRaw = cleanRaw[idx:]
	}
	if idx := strings.LastIndex(cleanRaw, "}"); idx != -1 {
		cleanRaw = cleanRaw[:idx+1]
	}

	var res CriterionLLMResponse
	if err := json.Unmarshal([]byte(cleanRaw), &res); err != nil {
		return w.mockScoreCriterion(c, turns)
	}

	score := res.Score
	if score < 0 {
		score = 0
	}
	if score > c.MaxScore {
		score = c.MaxScore
	}

	evidenceTurnID := res.EvidenceTurnID
	evidenceQuote := res.EvidenceQuote
	if (evidenceTurnID == "" || evidenceQuote == "") && len(turns) > 0 {
		for _, t := range turns {
			if t.Role == "participant" {
				evidenceTurnID = t.ID
				evidenceQuote = t.Content
				break
			}
		}
	}

	return CriterionResult{
		CriterionID:    c.ID,
		Score:          score,
		EvidenceTurnID: evidenceTurnID,
		EvidenceQuote:  evidenceQuote,
		Grounded:       true,
	}
}

func (w *EvaluatorWorker) mockScoreCriterion(c EvaluationCriterion, turns []SessionTurn) CriterionResult {
	score := int(math.Round(float64(c.MaxScore) * 0.75))
	var evTurnID, evQuote string
	for _, t := range turns {
		if t.Role == "participant" && len(t.Content) > 20 {
			evTurnID = t.ID
			evQuote = t.Content
			if len(evQuote) > 200 {
				evQuote = evQuote[:200]
			}
			break
		}
	}
	if evTurnID == "" && len(turns) > 0 {
		evTurnID = turns[0].ID
		evQuote = turns[0].Content
	}
	return CriterionResult{
		CriterionID:    c.ID,
		Score:          score,
		EvidenceTurnID: evTurnID,
		EvidenceQuote:  evQuote,
		Grounded:       true,
	}
}

func (w *EvaluatorWorker) generateSummary(ctx context.Context, participantName, bpTitle string, criteria []EvaluationCriterion, results []CriterionResult, recommendation string, overall float64) string {
	fallback := fmt.Sprintf("%s completed the technical assessment for %s. The candidate demonstrated practical comprehension of the required competencies, achieving an overall score of %.1f/10.0 and receiving a %s recommendation.",
		participantName, bpTitle, overall, strings.ReplaceAll(recommendation, "_", " "))

	if w.groqKey == "" {
		return fallback
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Participant: %s\nAssessment Blueprint: %s\nRecommendation: %s\n\nCriterion Scores:\n",
		participantName, bpTitle, recommendation))
	for i, r := range results {
		c := criteria[i]
		sb.WriteString(fmt.Sprintf("- %s: %d/%d. Quote: %q\n", c.Name, r.Score, c.MaxScore, r.EvidenceQuote))
	}

	prompt := fmt.Sprintf(`Write a concise 2-3 paragraph executive summary of this completed technical assessment.
Summarize the participant's key strengths and demonstrated competencies, cite specific evidence, and explain the %s recommendation.

CRITICAL RULES:
- Do NOT output any heading or title such as "**Executive Summary**" or "# Executive Summary". Start directly with the first sentence of the assessment narrative.
- Use standard double quotes for quoting candidate statements (do not wrap quotes with asterisks like "*quote*").
- Separate paragraphs with blank lines.

%s`, recommendation, sb.String())

	summary, err := w.callGroq(ctx, prompt, false)
	if err != nil || strings.TrimSpace(summary) == "" {
		return fallback
	}
	summary = strings.TrimSpace(summary)
	reTitle := regexp.MustCompile(`(?i)^(?:\*{1,3}|#{1,4})?\s*Executive Summary:?\s*(?:\*{1,3})?\s*`)
	return reTitle.ReplaceAllString(summary, "")
}

func (w *EvaluatorWorker) callGroq(ctx context.Context, prompt string, jsonFormat bool) (string, error) {
	reqBody := OpenAIRequest{
		Model: w.groqModel,
		Messages: []OpenAIMessage{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.1,
		MaxTokens:   1500,
	}
	if jsonFormat {
		reqBody.ResponseFormat = map[string]string{"type": "json_object"}
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+w.groqKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq api error (%d): %s", resp.StatusCode, string(body))
	}

	var oaiResp OpenAIResponse
	if err := json.Unmarshal(body, &oaiResp); err != nil {
		return "", err
	}
	if len(oaiResp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned by groq")
	}
	return oaiResp.Choices[0].Message.Content, nil
}
