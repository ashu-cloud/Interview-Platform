package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/client"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/config"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/llm"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/repository"
)

type EvaluationWorker struct {
	cfg            *config.Config
	evalRepo       *repository.FeedbackRepository
	transcriptRepo *repository.TranscriptRepository
	platformClient *client.PlatformClient
	evaluator      *Evaluator
	llmClient      llm.EvaluationLLM
	logger         *slog.Logger
	kafkaReader    *kafka.Reader
	kafkaWriter    *kafka.Writer
}

// Backward compatibility alias
type FeedbackWorker = EvaluationWorker

func NewEvaluationWorker(
	cfg *config.Config,
	evalRepo *repository.FeedbackRepository,
	transcriptRepo *repository.TranscriptRepository,
	platformClient *client.PlatformClient,
	evaluator *Evaluator,
	llmClient llm.EvaluationLLM,
	logger *slog.Logger,
	dialer *kafka.Dialer,
) *EvaluationWorker {
	if logger == nil {
		logger = slog.Default()
	}

	var reader *kafka.Reader
	var writer *kafka.Writer

	if cfg.KafkaBrokers != "" {
		brokers := strings.Split(cfg.KafkaBrokers, ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}

		if dialer == nil {
			dialer = kafka.DefaultDialer
		}

		groupID := os.Getenv("KAFKA_GROUP_ID")
		if groupID == "" {
			groupID = "evaluation-workers"
		}

		reader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			GroupID:  groupID,
			Topic:    "session.completed",
			MinBytes: 1,
			MaxBytes: 10e6,
			MaxWait:  time.Second,
			Dialer:   dialer,
		})

		writer = &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    "notification.events",
			Balancer: &kafka.Hash{},
			Transport: &kafka.Transport{
				Dial: dialer.DialFunc,
			},
		}
	}

	return &EvaluationWorker{
		cfg:            cfg,
		evalRepo:       evalRepo,
		transcriptRepo: transcriptRepo,
		platformClient: platformClient,
		evaluator:      evaluator,
		llmClient:      llmClient,
		logger:         logger,
		kafkaReader:    reader,
		kafkaWriter:    writer,
	}
}

// NewFeedbackWorker constructor for backward compatibility
func NewFeedbackWorker(
	brokers []string,
	jcs *client.JCSClient,
	llmClient LLMClient,
	repo *repository.FeedbackRepository,
	logger *slog.Logger,
	dialer *kafka.Dialer,
) *EvaluationWorker {
	brokerStr := strings.Join(brokers, ",")
	cfg := &config.Config{
		KafkaBrokers: brokerStr,
		PollInterval: 500 * time.Millisecond,
	}
	evalLLM, ok := llmClient.(llm.EvaluationLLM)
	if !ok {
		evalLLM = llm.NewMockEvaluationLLMClient()
	}
	evaluator := NewEvaluator(evalLLM, logger)
	transcriptRepo := repository.NewTranscriptRepository(repo.Pool())

	return NewEvaluationWorker(
		cfg,
		repo,
		transcriptRepo,
		jcs,
		evaluator,
		evalLLM,
		logger,
		dialer,
	)
}

// LLMClient interface for legacy callers
type LLMClient interface {
	GenerateFeedback(ctx context.Context, prompt string) (*model.FeedbackResponse, error)
}

func (w *EvaluationWorker) Run(ctx context.Context) error {
	if w.cfg.KafkaBrokers != "" {
		w.logger.Info("evaluation-worker: Kafka mode", "topic", "session.completed")
		return w.runKafkaMode(ctx)
	}
	w.logger.Info("evaluation-worker: Postgres outbox mode")
	return w.runOutboxMode(ctx)
}

func (w *EvaluationWorker) runKafkaMode(ctx context.Context) error {
	for {
		m, err := w.kafkaReader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			w.logger.Error("fetch message failed", "error", err)
			time.Sleep(time.Second)
			continue
		}

		event, err := model.ParseSessionCompletedEvent(m.Value)
		if err != nil {
			w.logger.Error("parse completed event failed", "error", err)
			_ = w.kafkaReader.CommitMessages(ctx, m)
			continue
		}

		if err := w.processMessage(ctx, event); err != nil {
			w.logger.Error("processMessage failed", "session_id", event.SessionID, "error", err)
			// Do not commit message on failure
			continue
		}

		if err := w.kafkaReader.CommitMessages(ctx, m); err != nil {
			w.logger.Error("commit message failed", "error", err)
		}
	}
}

func (w *EvaluationWorker) runOutboxMode(ctx context.Context) error {
	pollInterval := w.cfg.PollInterval
	if pollInterval <= 0 {
		pollInterval = 500 * time.Millisecond
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Initial immediate check
	w.processOutboxBatch(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			w.processOutboxBatch(ctx)
		}
	}
}

func (w *EvaluationWorker) processOutboxBatch(ctx context.Context) {
	rows, err := w.evalRepo.DequeueOutboxCompleted(ctx, 10)
	if err != nil {
		w.logger.Error("dequeue failed", "error", err)
		return
	}
	for _, row := range rows {
		event := model.SessionCompletedEvent{
			SessionID:     row.SessionID,
			OrgID:         row.OrgID,
			ParticipantID: row.ParticipantID,
			BlueprintID:   row.BlueprintID,
			TurnCount:     row.TurnCount,
		}
		if err := w.processMessage(ctx, event); err != nil {
			w.logger.Error("processMessage failed", "session_id", row.SessionID, "error", err)
			_ = w.evalRepo.IncrementOutboxRetry(ctx, row.ID)
			continue
		}
		_ = w.evalRepo.MarkOutboxCompleted(ctx, row.ID)
	}
}

var (
	ErrIncompleteTranscript = errors.New("incomplete transcript")
	ErrEmptyTranscript      = errors.New("empty transcript")
)

// ValidateTranscriptCompleteness verifies all turns from turn 1 to expectedTurnCount are present.
func ValidateTranscriptCompleteness(turns []SessionTurn, expectedTurnCount int) error {
	if len(turns) == 0 {
		return ErrEmptyTranscript
	}
	if expectedTurnCount <= 0 {
		return nil
	}

	seenTurns := make(map[int]bool, len(turns))
	highestTurn := 0
	for _, t := range turns {
		if t.TurnNumber > 0 {
			seenTurns[t.TurnNumber] = true
			if t.TurnNumber > highestTurn {
				highestTurn = t.TurnNumber
			}
		}
	}

	var missingTurns []int
	for k := 1; k <= expectedTurnCount; k++ {
		if !seenTurns[k] {
			missingTurns = append(missingTurns, k)
		}
	}

	if len(missingTurns) > 0 {
		return fmt.Errorf("%w: missing turns %v (found %d/%d turns, max seen: %d)",
			ErrIncompleteTranscript, missingTurns, len(seenTurns), expectedTurnCount, highestTurn)
	}

	return nil
}

func (w *EvaluationWorker) waitForTranscript(ctx context.Context,
	sessionID string, expectedMinTurns int) ([]SessionTurn, error) {

	const maxAttempts = 5
	backoff := 100 * time.Millisecond

	for attempt := 0; attempt < maxAttempts; attempt++ {
		turns, err := w.transcriptRepo.GetTranscript(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("get transcript: %w", err)
		}

		if err := ValidateTranscriptCompleteness(turns, expectedMinTurns); err == nil {
			return turns, nil
		}

		w.logger.Info("waiting for turn projector completeness",
			"attempt", attempt+1, "session_id", sessionID, "found_turns", len(turns), "expected_turn_count", expectedMinTurns)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}

	turns, err := w.transcriptRepo.GetTranscript(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get transcript final attempt: %w", err)
	}
	if err := ValidateTranscriptCompleteness(turns, expectedMinTurns); err != nil {
		return nil, fmt.Errorf("%w for session %s: %v", ErrIncompleteTranscript, sessionID, err)
	}
	return turns, nil
}

// AggregateScores computes overall score normalized to 0-10 and recommendation.
// Malformed, incomplete, or errored criteria return an explicit error and "evaluation_failed",
// preserving the distinction between candidate performance and system evaluation failure.
func AggregateScores(results []CriterionResult,
	criteria []model.EvaluationCriterion) (float64, string, error) {

	if len(criteria) == 0 {
		return 0, "incomplete", fmt.Errorf("no evaluation criteria provided")
	}

	maxByID := make(map[string]int, len(criteria))
	for _, c := range criteria {
		maxByID[c.ID] = c.MaxScore
	}

	var totalScore, totalMax float64
	var failedCriteria []string
	resultsByCrit := make(map[string]CriterionResult, len(results))

	for _, r := range results {
		resultsByCrit[r.CriterionID] = r
		if r.Error != nil {
			failedCriteria = append(failedCriteria, fmt.Sprintf("%s: %v", r.CriterionID, r.Error))
			continue
		}
		maxScore, exists := maxByID[r.CriterionID]
		if !exists {
			continue
		}
		totalScore += float64(r.Score)
		totalMax += float64(maxScore)
	}

	// Verify all blueprint criteria have results
	for _, c := range criteria {
		if _, ok := resultsByCrit[c.ID]; !ok {
			failedCriteria = append(failedCriteria, fmt.Sprintf("%s: missing result", c.ID))
		}
	}

	if len(failedCriteria) > 0 {
		return 0, "evaluation_failed", fmt.Errorf("processing failure on %d criteria: %s",
			len(failedCriteria), strings.Join(failedCriteria, "; "))
	}

	if totalMax == 0 {
		return 0, "evaluation_failed", fmt.Errorf("total maximum score is zero across criteria")
	}

	overall := (totalScore / totalMax) * 10.0
	overall = math.Round(overall*10) / 10

	switch {
	case overall >= 8.0:
		return overall, "strong_hire", nil
	case overall >= 6.0:
		return overall, "hire", nil
	case overall >= 4.0:
		return overall, "no_hire", nil
	default:
		return overall, "strong_no_hire", nil
	}
}

func (w *EvaluationWorker) processMessage(ctx context.Context,
	event model.SessionCompletedEvent) error {

	// 1. Idempotency check
	if processed, _ := w.evalRepo.IsProcessed(ctx, event.SessionID); processed {
		w.logger.Info("already evaluated, skipping", "session_id", event.SessionID)
		return nil
	}

	// 2. Fetch blueprint + participant from platform-service
	blueprint, err := w.platformClient.GetBlueprint(ctx, event.BlueprintID)
	if err != nil {
		return fmt.Errorf("get blueprint: %w", err)
	}
	participant, err := w.platformClient.GetParticipant(ctx, event.ParticipantID)
	if err != nil {
		return fmt.Errorf("get participant: %w", err)
	}

	// 3. O(1) transcript fetch from session_turns with completeness guarantee
	turns, err := w.waitForTranscript(ctx, event.SessionID, event.TurnCount)
	if err != nil {
		return fmt.Errorf("transcript: %w", err)
	}
	if len(turns) == 0 {
		return fmt.Errorf("no session turns found for %s", event.SessionID)
	}

	// 4. Scatter-gather: one LLM call per criterion, max 5 concurrent
	results := w.evaluator.EvaluateCriteria(ctx, blueprint.Criteria, turns,
		participant.FullName, 5)

	// 5. Grounding validation
	results = ValidateGrounding(results, turns)

	// 6. Aggregate
	overallScore, recommendation, err := AggregateScores(results, blueprint.Criteria)
	if err != nil {
		return fmt.Errorf("aggregate scores: %w", err)
	}

	// 7. Generate written summary (one additional LLM call after all scores known)
	summary, err := w.generateSummary(ctx, participant.FullName,
		blueprint.Title, results, blueprint.Criteria, recommendation)
	if err != nil {
		summary = "Summary generation failed."
	}

	// 8. Save evaluation report atomically
	evalData, err := buildEvaluationData(
		event, participant, summary, overallScore, recommendation, results, turns,
	)
	if err != nil {
		return fmt.Errorf("build evaluation data: %w", err)
	}

	reportID, err := w.evalRepo.SaveEvaluationReport(ctx, evalData)
	if err != nil {
		return fmt.Errorf("save evaluation: %w", err)
	}

	// 9. Publish notification
	return w.publishNotification(ctx, event, reportID)
}

type transcriptTurnView struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func buildEvaluationData(
	event model.SessionCompletedEvent,
	participant *model.Participant,
	summary string,
	overallScore float64,
	recommendation string,
	results []CriterionResult,
	turns []SessionTurn,
) (model.EvaluationReportData, error) {

	scores := make([]model.EvaluationScoreData, len(results))
	for i, r := range results {
		scores[i] = model.EvaluationScoreData{
			CriterionID:    r.CriterionID,
			Score:          r.Score,
			EvidenceQuote:  r.EvidenceQuote,
			EvidenceTurnID: r.EvidenceTurnID,
			IsGrounded:     r.Grounded,
		}
	}

	turnViews := make([]transcriptTurnView, len(turns))
	for i, t := range turns {
		role := t.Role
		if role == "participant" {
			role = "user"
		} else if role == "assessor" {
			role = "assistant"
		}
		turnViews[i] = transcriptTurnView{
			Role:    role,
			Content: t.Content,
		}
	}
	transcriptBytes, err := json.Marshal(turnViews)
	if err != nil {
		return model.EvaluationReportData{}, err
	}

	return model.EvaluationReportData{
		SessionID:      event.SessionID,
		OrgID:          event.OrgID,
		Summary:        summary,
		OverallScore:   overallScore,
		Recommendation: recommendation,
		Scores:         scores,
		TranscriptJSON: transcriptBytes,
	}, nil
}

func (w *EvaluationWorker) generateSummary(ctx context.Context,
	participantName string,
	blueprintTitle string,
	results []CriterionResult,
	criteria []model.EvaluationCriterion,
	recommendation string,
) (string, error) {

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Participant: %s\nAssessment Blueprint: %s\nRecommendation: %s\n\nCriterion Scores:\n",
		participantName, blueprintTitle, recommendation))

	criteriaMap := make(map[string]model.EvaluationCriterion)
	for _, c := range criteria {
		criteriaMap[c.ID] = c
	}

	for _, r := range results {
		c := criteriaMap[r.CriterionID]
		groundedStatus := "unverified"
		if r.Grounded {
			groundedStatus = "grounded"
		}
		sb.WriteString(fmt.Sprintf("- %s: %d/%d (%s) - Quote: %q\n",
			c.Name, r.Score, c.MaxScore, groundedStatus, r.EvidenceQuote))
	}

	prompt := fmt.Sprintf(`Write a concise 2-3 paragraph executive summary of this completed technical assessment.
Summarize the participant's key strengths and demonstrated competencies, cite specific evidence, and explain the %s recommendation.

CRITICAL RULES:
- Do NOT output any heading or title such as "**Executive Summary**" or "# Executive Summary". Start directly with the first sentence of the assessment narrative.
- Use standard double quotes for quoting candidate statements (do not wrap quotes with asterisks like "*quote*").
- Separate paragraphs with blank lines.

%s`, recommendation, sb.String())

	summary, err := w.llmClient.GenerateSummary(ctx, prompt)
	if err != nil {
		return "", err
	}
	summary = strings.TrimSpace(summary)
	// Strip redundant heading if LLM outputs it anyway
	reTitle := regexp.MustCompile(`(?i)^(?:\*{1,3}|#{1,4})?\s*Executive Summary:?\s*(?:\*{1,3})?\s*`)
	summary = reTitle.ReplaceAllString(summary, "")
	return strings.TrimSpace(summary), nil
}

func (w *EvaluationWorker) publishNotification(ctx context.Context,
	event model.SessionCompletedEvent,
	reportID string,
) error {
	notif := model.NotificationEvent{
		EventType:   "evaluation_ready",
		OrgID:       event.OrgID,
		InterviewID: event.SessionID,
		CandidateID: event.ParticipantID,
		FeedbackID:  reportID,
		Timestamp:   time.Now().UTC(),
	}

	b, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	if w.kafkaWriter != nil {
		if err := w.kafkaWriter.WriteMessages(ctx, kafka.Message{
			Key:   []byte(event.OrgID),
			Value: b,
		}); err != nil {
			w.logger.Warn("publish notification via kafka failed", "error", err)
		}
	}

	return nil
}

func (w *EvaluationWorker) Close() error {
	if w.kafkaWriter != nil {
		_ = w.kafkaWriter.Close()
	}
	if w.kafkaReader != nil {
		_ = w.kafkaReader.Close()
	}
	return nil
}
