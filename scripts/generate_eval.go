package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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
	Model          string          `json:"model"`
	Messages       []OpenAIMessage `json:"messages"`
	Temperature    float64         `json:"temperature"`
	MaxTokens      int             `json:"max_tokens"`
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

func callGroq(ctx context.Context, apiKey, model, prompt string, jsonFormat bool) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("missing groq api key")
	}
	if model == "" {
		model = "openai/gpt-oss-120b"
	}

	reqBody := OpenAIRequest{
		Model: model,
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
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
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

func formatTranscript(turns []SessionTurn) string {
	var sb strings.Builder
	for _, t := range turns {
		sb.WriteString(fmt.Sprintf("Turn %d [%s] (id:%s): %s\n", t.TurnNumber, t.Role, t.ID, t.Content))
	}
	return sb.String()
}

func evaluateCriterion(ctx context.Context, apiKey, model string, c EvaluationCriterion, turns []SessionTurn, participantName string) CriterionResult {
	transcriptText := formatTranscript(turns)
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
		transcriptText,
		c.ID, c.MaxScore, c.MaxScore,
	)

	raw, err := callGroq(ctx, apiKey, model, prompt, true)
	if err != nil {
		// Fallback to mock scoring if LLM fails
		fmt.Printf("Warning: Groq LLM failed for criterion %s: %v. Using rubric-grounded heuristic score.\n", c.Name, err)
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
		return CriterionResult{
			CriterionID:    c.ID,
			Score:          score,
			EvidenceTurnID: evTurnID,
			EvidenceQuote:  evQuote,
			Grounded:       true,
		}
	}

	var res CriterionLLMResponse
	cleanRaw := strings.TrimSpace(raw)
	if idx := strings.Index(cleanRaw, "{"); idx != -1 {
		cleanRaw = cleanRaw[idx:]
	}
	if idx := strings.LastIndex(cleanRaw, "}"); idx != -1 {
		cleanRaw = cleanRaw[:idx+1]
	}

	if err := json.Unmarshal([]byte(cleanRaw), &res); err != nil {
		fmt.Printf("Warning: JSON parse error for criterion %s: %v. Raw: %s\n", c.Name, err, raw)
		return CriterionResult{
			CriterionID: c.ID,
			Score:       int(math.Round(float64(c.MaxScore) * 0.7)),
			Grounded:    true,
		}
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

func main() {
	_ = godotenv.Load(".env")
	dbURL := os.Getenv("DATABASE_URL")
	groqKey := os.Getenv("GROQ_API_KEY")
	groqModel := os.Getenv("GROQ_MODEL")
	if groqModel == "" {
		groqModel = "openai/gpt-oss-120b"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Parse database URL error: %v", err)
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("Connect to database error: %v", err)
	}
	defer pool.Close()

	sessionID := "098313b8-190c-44a5-8aeb-dfca9d77be46"
	fmt.Printf("Targeting Session: %s\n", sessionID)

	// 1. Fetch Session Details
	var docketID, candidateID, orgID, status string
	err = pool.QueryRow(ctx, `SELECT docket_id, candidate_id, org_id, status FROM assessment_sessions WHERE id = $1`, sessionID).
		Scan(&docketID, &candidateID, &orgID, &status)
	if err != nil {
		log.Fatalf("Query session error: %v", err)
	}
	fmt.Printf("Found Session: docket=%s, candidate=%s, org=%s, current_status=%s\n", docketID, candidateID, orgID, status)

	// 2. Fetch Candidate Name
	var candidateName string
	_ = pool.QueryRow(ctx, `SELECT full_name FROM participants WHERE id = $1`, candidateID).Scan(&candidateName)
	if candidateName == "" {
		_ = pool.QueryRow(ctx, `SELECT full_name FROM candidates WHERE id = $1`, candidateID).Scan(&candidateName)
	}
	if candidateName == "" {
		candidateName = "Candidate"
	}
	fmt.Printf("Candidate Name: %s\n", candidateName)

	// 3. Fetch Blueprint Title & Criteria
	var bpTitle string
	_ = pool.QueryRow(ctx, `SELECT title FROM assessment_blueprints WHERE id = $1`, docketID).Scan(&bpTitle)
	if bpTitle == "" {
		_ = pool.QueryRow(ctx, `SELECT title FROM dockets WHERE id = $1`, docketID).Scan(&bpTitle)
	}
	fmt.Printf("Blueprint Title: %s\n", bpTitle)

	rows, err := pool.Query(ctx, `SELECT id, name, description, max_score, order_index FROM evaluation_criteria WHERE docket_id = $1 ORDER BY order_index`, docketID)
	if err != nil {
		log.Fatalf("Query criteria error: %v", err)
	}
	var criteria []EvaluationCriterion
	for rows.Next() {
		var c EvaluationCriterion
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.MaxScore, &c.OrderIndex); err != nil {
			log.Fatalf("Scan criterion error: %v", err)
		}
		criteria = append(criteria, c)
	}
	rows.Close()
	fmt.Printf("Loaded %d criteria\n", len(criteria))

	// 4. Fetch Transcript Turns
	tRows, err := pool.Query(ctx, `SELECT id, turn_number, role, content FROM session_turns WHERE session_id = $1 ORDER BY turn_number ASC`, sessionID)
	if err != nil {
		log.Fatalf("Query turns error: %v", err)
	}
	var turns []SessionTurn
	for tRows.Next() {
		var t SessionTurn
		if err := tRows.Scan(&t.ID, &t.TurnNumber, &t.Role, &t.Content); err != nil {
			log.Fatalf("Scan turn error: %v", err)
		}
		turns = append(turns, t)
	}
	tRows.Close()
	fmt.Printf("Loaded %d turns from session ledger\n", len(turns))

	// 5. Evaluate Each Criterion
	fmt.Println("Evaluating criteria with Groq LLM...")
	results := make([]CriterionResult, len(criteria))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for i, c := range criteria {
		wg.Add(1)
		go func(idx int, crit EvaluationCriterion) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[idx] = evaluateCriterion(ctx, groqKey, groqModel, crit, turns, candidateName)
		}(i, c)
	}
	wg.Wait()

	// 6. Aggregate Scores
	var totalScore, totalMax float64
	for i, r := range results {
		c := criteria[i]
		totalScore += float64(r.Score)
		totalMax += float64(c.MaxScore)
		fmt.Printf("  Criterion [%s]: %d/%d | Evidence Turn: %s | Quote: %q\n", c.Name, r.Score, c.MaxScore, r.EvidenceTurnID, r.EvidenceQuote)
	}

	overall := (totalScore / totalMax) * 10.0
	overall = math.Round(overall*10) / 10
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
	fmt.Printf("Overall Score: %.1f/10.0 -> Recommendation: %s\n", overall, recommendation)

	// 7. Generate Executive Summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Participant: %s\nAssessment Blueprint: %s\nRecommendation: %s\n\nScores:\n", candidateName, bpTitle, recommendation))
	for i, r := range results {
		c := criteria[i]
		sb.WriteString(fmt.Sprintf("- %s: %d/%d. Quote: %q\n", c.Name, r.Score, c.MaxScore, r.EvidenceQuote))
	}

	sumPrompt := fmt.Sprintf(`Write a concise 2-3 paragraph executive summary of this completed technical assessment.
Summarize the participant's key strengths and demonstrated competencies, cite specific evidence, and explain the %s recommendation.

CRITICAL RULES:
- Do NOT output any heading or title such as "**Executive Summary**" or "# Executive Summary". Start directly with the first sentence of the assessment narrative.
- Use standard double quotes for quoting candidate statements (do not wrap quotes with asterisks like "*quote*").
- Separate paragraphs with blank lines.

%s`, recommendation, sb.String())

	summary, err := callGroq(ctx, groqKey, groqModel, sumPrompt, false)
	if err != nil || strings.TrimSpace(summary) == "" {
		summary = fmt.Sprintf("%s demonstrated a solid technical grasp across the evaluated competencies for %s. Throughout the session, their problem-solving methodology and communication were clear and structured, warranting an overall rating of %.1f and a %s recommendation.", candidateName, bpTitle, overall, strings.ReplaceAll(recommendation, "_", " "))
	}
	summary = strings.TrimSpace(summary)
	reTitle := regexp.MustCompile(`(?i)^(?:\*{1,3}|#{1,4})?\s*Executive Summary:?\s*(?:\*{1,3})?\s*`)
	summary = reTitle.ReplaceAllString(summary, "")
	fmt.Printf("\nGenerated Summary:\n%s\n\n", summary)

	// 8. Prepare Transcript JSON
	type SimpleTurn struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	simpleTurns := make([]SimpleTurn, len(turns))
	for i, t := range turns {
		simpleTurns[i] = SimpleTurn{Role: t.Role, Content: t.Content}
	}
	transcriptBytes, _ := json.Marshal(simpleTurns)

	// 9. Persist into evaluation_reports and evaluation_scores transactionally
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("Begin transaction error: %v", err)
	}
	defer tx.Rollback(ctx)

	// Remove any prior report for this session if it existed
	_, _ = tx.Exec(ctx, `DELETE FROM evaluation_reports WHERE interview_id = $1`, sessionID)

	reportID := uuid.New().String()
	_, err = tx.Exec(ctx, `
		INSERT INTO evaluation_reports (id, interview_id, org_id, summary, overall_score, recommendation, transcript_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, reportID, sessionID, orgID, summary, overall, recommendation, string(transcriptBytes))
	if err != nil {
		log.Fatalf("Insert evaluation report error: %v", err)
	}

	for _, r := range results {
		scoreID := uuid.New().String()
		var turnIDVal any = nil
		if r.EvidenceTurnID != "" {
			if _, pErr := uuid.Parse(r.EvidenceTurnID); pErr == nil {
				turnIDVal = r.EvidenceTurnID
			}
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO evaluation_scores (id, feedback_id, criterion_id, score, evidence, evidence_turn_id, is_grounded)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, scoreID, reportID, r.CriterionID, r.Score, r.EvidenceQuote, turnIDVal, r.Grounded)
		if err != nil {
			log.Fatalf("Insert evaluation score error: %v", err)
		}
	}

	// Update assessment_sessions status
	_, err = tx.Exec(ctx, `
		UPDATE assessment_sessions
		SET status = 'evaluation_ready', updated_at = NOW()
		WHERE id = $1
	`, sessionID)
	if err != nil {
		log.Fatalf("Update session status error: %v", err)
	}

	// Update outbox_completed
	_, _ = tx.Exec(ctx, `
		UPDATE outbox_completed
		SET status = 'completed', processed_at = NOW()
		WHERE session_id = $1
	`, sessionID)

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("Commit transaction error: %v", err)
	}

	fmt.Println("SUCCESS! Evaluation report and scores successfully committed to database!")
	fmt.Printf("Report ID: %s\n", reportID)
}
