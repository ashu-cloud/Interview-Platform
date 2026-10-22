package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
)

type CriterionLLMResponse struct {
	CriterionID    string `json:"criterion_id"`
	Score          int    `json:"score"`
	EvidenceTurnID string `json:"evidence_turn_id"`
	EvidenceQuote  string `json:"evidence_quote"`
}

type EvaluationLLM interface {
	EvaluateCriterion(ctx context.Context, prompt string) (*CriterionLLMResponse, error)
	GenerateSummary(ctx context.Context, prompt string) (string, error)
}

type EvaluationLLMClient struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

type FeedbackLLMClient = EvaluationLLMClient

type MockEvaluationLLMClient struct{}

type MockFeedbackLLMClient = MockEvaluationLLMClient

func NewMockEvaluationLLMClient() *MockEvaluationLLMClient {
	return &MockEvaluationLLMClient{}
}

func NewMockFeedbackLLMClient() *MockFeedbackLLMClient {
	return NewMockEvaluationLLMClient()
}

var (
	rubricLinePattern       = regexp.MustCompile(`(?m)^- \[([0-9a-fA-F-]{36})\].*?\(max (\d+) points\):`)
	criterionIDPattern      = regexp.MustCompile(`"criterion_id":\s*"([^"]+)"`)
	maxScorePattern         = regexp.MustCompile(`MAX SCORE:\s*(\d+)`)
	turnParticipantPattern  = regexp.MustCompile(`(?m)Turn \d+ \[participant\] \(id:([a-zA-Z0-9-]+)\):\s*([^\r\n]+)`)
	turnAnyPattern          = regexp.MustCompile(`(?m)Turn \d+ \[[^\]]+\] \(id:([a-zA-Z0-9-]+)\):\s*([^\r\n]+)`)
)

func (c *MockEvaluationLLMClient) EvaluateCriterion(_ context.Context, prompt string) (*CriterionLLMResponse, error) {
	criterionID := ""
	if m := criterionIDPattern.FindStringSubmatch(prompt); len(m) > 1 {
		criterionID = m[1]
	}

	maxScore := 10
	if m := maxScorePattern.FindStringSubmatch(prompt); len(m) > 1 {
		if s, err := strconv.Atoi(m[1]); err == nil && s > 0 {
			maxScore = s
		}
	}

	turnID := ""
	quote := ""

	matches := turnParticipantPattern.FindAllStringSubmatch(prompt, -1)
	if len(matches) == 0 {
		matches = turnAnyPattern.FindAllStringSubmatch(prompt, -1)
	}

	if len(matches) > 0 {
		turnID = matches[0][1]
		content := strings.TrimSpace(matches[0][2])
		words := strings.Fields(content)
		if len(words) > 15 {
			quote = strings.Join(words[:15], " ")
		} else {
			quote = content
		}
	}

	score := int(math.Round(float64(maxScore) * 0.75))
	if score > maxScore {
		score = maxScore
	}

	return &CriterionLLMResponse{
		CriterionID:    criterionID,
		Score:          score,
		EvidenceTurnID: turnID,
		EvidenceQuote:  quote,
	}, nil
}

func (c *MockEvaluationLLMClient) GenerateSummary(_ context.Context, prompt string) (string, error) {
	return "The participant demonstrated strong competence across evaluated criteria. Responses were structured, technical decisions were well-justified, and communication remained clear and evidence-grounded throughout the session.", nil
}

func (c *MockEvaluationLLMClient) GenerateFeedback(_ context.Context, prompt string) (*model.FeedbackResponse, error) {
	matches := rubricLinePattern.FindAllStringSubmatch(prompt, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("mock feedback could not find rubric criteria in prompt")
	}

	evidence := "The candidate's responses in the interview transcript were reviewed."
	if transcriptIndex := strings.Index(prompt, "FULL INTERVIEW TRANSCRIPT:\n"); transcriptIndex >= 0 {
		transcript := strings.TrimSpace(prompt[transcriptIndex+len("FULL INTERVIEW TRANSCRIPT:\n"):])
		if transcript != "" {
			if len(transcript) > 240 {
				transcript = transcript[:240]
			}
			evidence = transcript
		}
	}

	scores := make([]model.FeedbackScore, 0, len(matches))
	total, possible := 0, 0
	for _, match := range matches {
		criterionID, err := uuid.Parse(match[1])
		if err != nil {
			return nil, fmt.Errorf("parse mock criterion id: %w", err)
		}
		maxScore, err := strconv.Atoi(match[2])
		if err != nil || maxScore < 0 {
			return nil, fmt.Errorf("parse mock maximum score")
		}
		score := int(math.Round(float64(maxScore) * 0.7))
		scores = append(scores, model.FeedbackScore{CriterionID: criterionID, Score: score, Evidence: evidence})
		total += score
		possible += maxScore
	}

	overall := 0.0
	if possible > 0 {
		overall = math.Round((float64(total)/float64(possible)*10)*10) / 10
	}
	return &model.FeedbackResponse{
		Summary:        "Mock local evaluation generated from the completed interview transcript. Configure EVALUATION_LLM_MODE=groq with GROQ_API_KEY to use a live model.",
		OverallScore:   overall,
		Recommendation: "hire",
		Scores:         scores,
	}, nil
}

func NewEvaluationLLMClient(apiKey string) *EvaluationLLMClient {
	return &EvaluationLLMClient{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1/chat/completions",
		Model:   "gpt-4o",
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func NewFeedbackLLMClient(apiKey string) *EvaluationLLMClient {
	return NewEvaluationLLMClient(apiKey)
}

func NewGroqEvaluationLLMClient(apiKey, model string) *EvaluationLLMClient {
	if model == "" {
		model = "openai/gpt-oss-120b"
	}
	return &EvaluationLLMClient{
		APIKey:  apiKey,
		BaseURL: "https://api.groq.com/openai/v1/chat/completions",
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func NewGroqFeedbackLLMClient(apiKey, model string) *EvaluationLLMClient {
	return NewGroqEvaluationLLMClient(apiKey, model)
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiRequest struct {
	Model          string            `json:"model"`
	Messages       []openaiMessage   `json:"messages"`
	Temperature    float64           `json:"temperature"`
	MaxTokens      int               `json:"max_tokens"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *EvaluationLLMClient) callOpenAI(ctx context.Context, prompt string, temperature float64, jsonObject bool, maxTokens int) (string, error) {
	modelName := c.Model
	if modelName == "" {
		modelName = "openai/gpt-oss-120b"
	}
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1/chat/completions"
	}

	reqBody := openaiRequest{
		Model:       modelName,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Messages: []openaiMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
	if jsonObject {
		reqBody.ResponseFormat = map[string]string{"type": "json_object"}
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal req: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody bytes.Buffer
		_, _ = errBody.ReadFrom(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, errBody.String())
	}

	var res openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decode res: %w", err)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return res.Choices[0].Message.Content, nil
}

var (
	ErrMalformedOutput       = errors.New("malformed LLM output")
	ErrTruncatedOutput       = errors.New("truncated LLM output")
	ErrMissingRequiredFields = errors.New("missing required fields in LLM output")
)

// ParseCriterionResponse parses and validates an LLM criterion evaluation response with bounded recovery.
func ParseCriterionResponse(content string) (*CriterionLLMResponse, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, fmt.Errorf("%w: empty response", ErrMalformedOutput)
	}

	// 1. Strip markdown code fences if present
	if strings.HasPrefix(trimmed, "```json") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	} else if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
		trimmed = strings.TrimSpace(trimmed)
	}

	// 2. Locate JSON object boundary
	firstBrace := strings.Index(trimmed, "{")
	if firstBrace < 0 {
		return nil, fmt.Errorf("%w: missing opening brace in %q", ErrMalformedOutput, trimmed)
	}

	lastBrace := strings.LastIndex(trimmed, "}")
	var rawJSON string
	if lastBrace > firstBrace {
		rawJSON = trimmed[firstBrace : lastBrace+1]
	} else {
		// Truncated: no closing brace found
		rawCandidate := trimmed[firstBrace:]
		recovered := attemptBoundedTruncationRepair(rawCandidate)
		if recovered == "" {
			return nil, fmt.Errorf("%w: unrecoverable truncated JSON %q", ErrTruncatedOutput, rawCandidate)
		}
		rawJSON = recovered
	}

	// 3. Unmarshal into generic map to verify presence of required keys
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(rawJSON), &rawMap); err != nil {
		recovered := attemptBoundedTruncationRepair(rawJSON)
		if recovered != "" && json.Unmarshal([]byte(recovered), &rawMap) == nil {
			rawJSON = recovered
		} else {
			return nil, fmt.Errorf("%w: %v (raw: %s)", ErrMalformedOutput, err, rawJSON)
		}
	}

	// 4. Validate required fields
	criterionRaw, hasCrit := rawMap["criterion_id"]
	scoreRaw, hasScore := rawMap["score"]
	if !hasCrit || !hasScore {
		return nil, fmt.Errorf("%w: criterion_id or score missing", ErrMissingRequiredFields)
	}

	var critID string
	if err := json.Unmarshal(criterionRaw, &critID); err != nil || strings.TrimSpace(critID) == "" {
		return nil, fmt.Errorf("%w: invalid or empty criterion_id", ErrMissingRequiredFields)
	}

	var score int
	if err := json.Unmarshal(scoreRaw, &score); err != nil {
		return nil, fmt.Errorf("%w: invalid non-integer score: %v", ErrMalformedOutput, err)
	}

	var turnID string
	if tidRaw, ok := rawMap["evidence_turn_id"]; ok {
		_ = json.Unmarshal(tidRaw, &turnID)
	}

	var quote string
	if qRaw, ok := rawMap["evidence_quote"]; ok {
		_ = json.Unmarshal(qRaw, &quote)
	}

	return &CriterionLLMResponse{
		CriterionID:    strings.TrimSpace(critID),
		Score:          score,
		EvidenceTurnID: strings.TrimSpace(turnID),
		EvidenceQuote:  strings.TrimSpace(quote),
	}, nil
}

func attemptBoundedTruncationRepair(raw string) string {
	raw = strings.TrimSpace(raw)
	// If unclosed quote
	quoteCount := strings.Count(raw, `"`)
	if quoteCount%2 != 0 {
		raw += `"`
	}
	if !strings.HasSuffix(raw, "}") {
		raw += "}"
	}
	var testMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &testMap); err == nil {
		return raw
	}
	return ""
}

func (c *EvaluationLLMClient) EvaluateCriterion(ctx context.Context, prompt string) (*CriterionLLMResponse, error) {
	content, err := c.callOpenAI(ctx, prompt, 0.1, true, 500)
	if err != nil {
		// One retry
		content, err = c.callOpenAI(ctx, prompt, 0.1, true, 500)
		if err != nil {
			return nil, fmt.Errorf("call LLM for criterion: %w", err)
		}
	}

	return ParseCriterionResponse(content)
}

func (c *EvaluationLLMClient) GenerateSummary(ctx context.Context, prompt string) (string, error) {
	content, err := c.callOpenAI(ctx, prompt, 0.3, false, 1500)
	if err != nil {
		content, err = c.callOpenAI(ctx, prompt, 0.3, false, 1500)
		if err != nil {
			return "", fmt.Errorf("call LLM for summary: %w", err)
		}
	}
	return strings.TrimSpace(content), nil
}

func (c *EvaluationLLMClient) GenerateFeedback(ctx context.Context, prompt string) (*model.FeedbackResponse, error) {
	content, err := c.callOpenAI(ctx, prompt, 0.3, true, 2000)
	if err != nil {
		return nil, err
	}

	var feedback model.FeedbackResponse
	if err := json.Unmarshal([]byte(content), &feedback); err != nil {
		return nil, fmt.Errorf("unmarshal JSON output: %w", err)
	}

	return &feedback, nil
}

