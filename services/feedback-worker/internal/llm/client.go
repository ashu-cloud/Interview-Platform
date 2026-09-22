package llm

import (
	"bytes"
	"context"
	"encoding/json"
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

type FeedbackLLMClient struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// MockFeedbackLLMClient produces deterministic rubric-complete feedback for local demos.
// It intentionally derives criterion IDs and maximum scores from the prompt, so it exercises
// the same validation and persistence path as a real provider response.
type MockFeedbackLLMClient struct{}

func NewMockFeedbackLLMClient() *MockFeedbackLLMClient {
	return &MockFeedbackLLMClient{}
}

var rubricLinePattern = regexp.MustCompile(`(?m)^- \[([0-9a-fA-F-]{36})\].*?\(max (\d+) points\):`)

func (c *MockFeedbackLLMClient) GenerateFeedback(_ context.Context, prompt string) (*model.FeedbackResponse, error) {
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
		Summary:        "Mock local evaluation generated from the completed interview transcript. Configure FEEDBACK_LLM_MODE=groq with GROQ_API_KEY to use a live model.",
		OverallScore:   overall,
		Recommendation: "hire",
		Scores:         scores,
	}, nil
}

func NewFeedbackLLMClient(apiKey string) *FeedbackLLMClient {
	return &FeedbackLLMClient{
		APIKey:  apiKey,
		BaseURL: "https://api.openai.com/v1/chat/completions",
		Model:   "gpt-4o",
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func NewGroqFeedbackLLMClient(apiKey string) *FeedbackLLMClient {
	return &FeedbackLLMClient{
		APIKey:  apiKey,
		BaseURL: "https://api.groq.com/openai/v1/chat/completions",
		Model:   "openai/gpt-oss-120b",
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
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
	ResponseFormat map[string]string `json:"response_format"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *FeedbackLLMClient) GenerateFeedback(ctx context.Context, prompt string) (*model.FeedbackResponse, error) {
	modelName := c.Model
	if modelName == "" {
		modelName = "llama-3.3-70b-versatile"
	}
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1/chat/completions"
	}

	reqBody := openaiRequest{
		Model:          modelName,
		Temperature:    0.3,
		MaxTokens:      2000,
		ResponseFormat: map[string]string{"type": "json_object"},
		Messages: []openaiMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal req: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody bytes.Buffer
		_, _ = errBody.ReadFrom(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, errBody.String())
	}

	var res openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode res: %w", err)
	}

	if len(res.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	content := res.Choices[0].Message.Content

	var feedback model.FeedbackResponse
	if err := json.Unmarshal([]byte(content), &feedback); err != nil {
		return nil, fmt.Errorf("unmarshal JSON output: %w", err)
	}

	return &feedback, nil
}
