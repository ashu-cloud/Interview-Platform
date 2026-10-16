package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/ai-interview/services/feedback-worker/internal/llm"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/repository"
)

func TestEvaluationReliabilityBenchmark(t *testing.T) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		// Fallback to reading .env directly if not exported in env
		envBytes, err := os.ReadFile("../../../../.env")
		if err == nil {
			for _, line := range strings.Split(string(envBytes), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "GROQ_API_KEY=") {
					apiKey = strings.TrimPrefix(line, "GROQ_API_KEY=")
					break
				}
			}
		}
	}
	if apiKey == "" {
		t.Skip("GROQ_API_KEY not set; skipping live evaluation reliability benchmark")
	}

	modelName := "openai/gpt-oss-120b"
	benchLLM := &benchGroqLLM{
		apiKey:     apiKey,
		model:      modelName,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}

	criteria := []model.EvaluationCriterion{
		{
			ID:          "crit-1",
			Name:        "Distributed Concurrency & Locking",
			Description: "Evaluates ability to design race-free distributed locking mechanisms, TTL safety, and lease renewal.",
			MaxScore:    10,
		},
		{
			ID:          "crit-2",
			Name:        "Transactional Data Consistency",
			Description: "Evaluates understanding of database transaction isolation, transactional outbox pattern, and idempotent consumer semantics.",
			MaxScore:    10,
		},
		{
			ID:          "crit-3",
			Name:        "Real-Time Pipeline Latency Optimization",
			Description: "Evaluates techniques to reduce turn latency, streaming token chunking, and handling audio interruption barge-in.",
			MaxScore:    10,
		},
	}

	turns := []repository.SessionTurn{
		{
			ID:         "turn-1",
			TurnNumber: 1,
			Role:       "participant",
			Content:    "In our architecture, we eliminated double-booking races by implementing Redis SETNX locks with a 2-second lease TTL and background ticker renewal. If the lock cannot be acquired, the request fails fast with an ALREADY_IN_PROGRESS error.",
		},
		{
			ID:         "turn-2",
			TurnNumber: 2,
			Role:       "participant",
			Content:    "For database persistence, we avoid bare pool queries and instead execute PostgreSQL outbox processing in a single atomic transaction. We use SELECT FOR UPDATE SKIP LOCKED to allow concurrent workers to dequeue pending events without lock contention or duplicated rows.",
		},
		{
			ID:         "turn-3",
			TurnNumber: 3,
			Role:       "participant",
			Content:    "To minimize turn-taking latency, we stream tokens directly from Groq and chunk text into complete grammatical sentences before sending to ElevenLabs TTS. When the candidate interrupts, we cancel the context to immediately abort downstream audio generation and drop stale chunks on the client.",
		},
	}

	const k = 5
	type RunRecord struct {
		Score        int
		Grounded     bool
		Quote        string
		TurnID       string
		Duration     time.Duration
	}

	records := make([]RunRecord, k)

	targetCriterion := criteria[0] // Distributed Concurrency & Locking
	evaluator := NewEvaluator(benchLLM, nil)

	fmt.Println("\n=========================================================================================")
	fmt.Printf("   EVALUATION RELIABILITY BENCHMARK (k=%d, Criterion=%s, Model=%s)\n", k, targetCriterion.Name, modelName)
	fmt.Println("=========================================================================================")

	for i := 0; i < k; i++ {
		if i > 0 {
			time.Sleep(7 * time.Second) // Respect Groq 8000 TPM limit
		}
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		results := evaluator.EvaluateCriteria(ctx, []model.EvaluationCriterion{targetCriterion}, turns, "Candidate Alex", 1)
		cancel()

		validated := ValidateGrounding(results, turns)
		duration := time.Since(start)

		r := validated[0]
		records[i] = RunRecord{
			Score:    r.Score,
			Grounded: r.Grounded,
			Quote:    r.EvidenceQuote,
			TurnID:   r.EvidenceTurnID,
			Duration: duration,
		}

		fmt.Printf("Run %d: Score=%d/10 Grounded=%v TurnID=%s Time=%v Quote=%q\n",
			i+1, r.Score, r.Grounded, r.EvidenceTurnID, duration.Round(time.Millisecond), r.EvidenceQuote)
	}

	// Calculate statistics
	var sum, sumSq float64
	minScore, maxScore := 999, -1
	groundedCount := 0

	for _, rec := range records {
		s := rec.Score
		sum += float64(s)
		sumSq += float64(s * s)
		if s < minScore {
			minScore = s
		}
		if s > maxScore {
			maxScore = s
		}
		if rec.Grounded {
			groundedCount++
		}
	}

	mean := sum / float64(k)
	variance := (sumSq / float64(k)) - (mean * mean)
	if variance < 0 {
		variance = 0
	}
	stdDev := math.Sqrt(variance)
	groundedRate := float64(groundedCount) / float64(k) * 100.0

	fmt.Println(strings.Repeat("-", 89))
	fmt.Printf("RELIABILITY METRICS FOR %q:\n", targetCriterion.Name)
	fmt.Printf("Mean Score:       %.2f / 10\n", mean)
	fmt.Printf("Std Dev (σ):      %.2f\n", stdDev)
	fmt.Printf("Score Range:      %d - %d\n", minScore, maxScore)
	fmt.Printf("Grounded Rate:    %.1f%% (%d/%d citations passed verbatim validation)\n", groundedRate, groundedCount, k)
	fmt.Println("=========================================================================================")
}

type benchGroqLLM struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func (b *benchGroqLLM) EvaluateCriterion(ctx context.Context, prompt string) (*llm.CriterionLLMResponse, error) {
	reqBody := map[string]any{
		"model":       b.model,
		"temperature": 0.2,
		"max_tokens":  600,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("groq status %d: %s", resp.StatusCode, buf.String())
	}

	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	if len(res.Choices) == 0 {
		return nil, fmt.Errorf("empty choices from groq")
	}

	raw := res.Choices[0].Message.Content
	// Extract JSON block if wrapped in markdown
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}

	var parsed llm.CriterionLLMResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w (raw: %s)", err, raw)
	}
	return &parsed, nil
}

func (b *benchGroqLLM) GenerateSummary(ctx context.Context, prompt string) (string, error) {
	return "summary", nil
}

