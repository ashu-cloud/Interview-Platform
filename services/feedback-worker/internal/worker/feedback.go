package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/client"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/repository"
)

type LLMClient interface {
	GenerateFeedback(ctx context.Context, prompt string) (*model.FeedbackResponse, error)
}

type FeedbackWorker struct {
	reader        *kafka.Reader
	producer      *kafka.Writer // notification.events producer
	transcriptRdr *kafka.Reader // interview.turns reader
	jcs           *client.JCSClient
	llm           LLMClient
	repo          *repository.FeedbackRepository
	logger        *slog.Logger
	brokers       []string
	dialer        *kafka.Dialer
}

func NewFeedbackWorker(
	brokers []string,
	jcs *client.JCSClient,
	llmClient LLMClient,
	repo *repository.FeedbackRepository,
	logger *slog.Logger,
	dialer *kafka.Dialer,
) *FeedbackWorker {
	if dialer == nil {
		dialer = kafka.DefaultDialer
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "feedback-workers"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    "interview.completed",
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  time.Second,
		Dialer:   dialer,
	})

	transcriptRdr := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "interview.turns",
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  time.Second,
		Dialer:   dialer,
	})

	producer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    "notification.events",
		Balancer: &kafka.Hash{},
		Transport: &kafka.Transport{
			Dial: dialer.DialFunc,
		},
	}

	return &FeedbackWorker{
		reader:        reader,
		producer:      producer,
		transcriptRdr: transcriptRdr,
		jcs:           jcs,
		llm:           llmClient,
		repo:          repo,
		logger:        logger,
		brokers:       brokers,
		dialer:        dialer,
	}
}

func (w *FeedbackWorker) Run(ctx context.Context) error {
	w.logger.Info("Feedback worker started")
	for {
		m, err := w.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			w.logger.Error("fetch message failed", "error", err)
			time.Sleep(time.Second) // basic backoff
			continue
		}

		if err := w.processMessage(ctx, m); err != nil {
			w.logger.Error("process message failed", "error", err, "partition", m.Partition, "offset", m.Offset)
			// Do not commit
			continue
		}

		if err := w.reader.CommitMessages(ctx, m); err != nil {
			w.logger.Error("commit message failed", "error", err)
		}
	}
}

func (w *FeedbackWorker) processMessage(ctx context.Context, m kafka.Message) error {
	var evt model.InterviewCompletedEvent
	if err := json.Unmarshal(m.Value, &evt); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	w.logger.Info("processing interview completed", "interview_id", evt.InterviewID)

	// Idempotency check
	processed, err := w.repo.IsProcessed(ctx, evt.InterviewID)
	if err != nil {
		return fmt.Errorf("check processed: %w", err)
	}
	if processed {
		w.logger.Info("interview already processed", "interview_id", evt.InterviewID)
		return nil
	}

	// 1. Fetch Docket & Candidate
	docket, err := w.jcs.GetDocket(ctx, evt.DocketID)
	if err != nil {
		return fmt.Errorf("get docket: %w", err)
	}

	cand, err := w.jcs.GetCandidate(ctx, evt.CandidateID)
	if err != nil {
		return fmt.Errorf("get candidate: %w", err)
	}
	candName := cand.FullName

	w.logger.Info("fetching transcript", "interview_id", evt.InterviewID, "turn_count", evt.TurnCount)
	// 2. Reconstruct Transcript
	turns, transcriptText, err := w.reconstructTranscript(ctx, evt.InterviewID, evt.TurnCount)
	if err != nil {
		return fmt.Errorf("reconstruct transcript: %w", err)
	}

	// 3. Build Prompt
	prompt := w.buildPrompt(docket, candName, transcriptText)

	w.logger.Info("generating feedback", "interview_id", evt.InterviewID)
	// 4. LLM Generation
	feedback, err := w.generateFeedbackWithRetry(ctx, prompt)
	if err != nil {
		return fmt.Errorf("generate feedback: %w", err)
	}

	// 5. Validation
	if err := w.validateFeedback(feedback, docket); err != nil {
		return fmt.Errorf("invalid feedback: %w", err)
	}

	w.logger.Info("saving feedback", "interview_id", evt.InterviewID)

	transcriptJSON, err := json.Marshal(turns)
	if err != nil {
		return fmt.Errorf("marshal transcript: %w", err)
	}

	// 6. Persist
	feedbackID, err := w.repo.SaveFeedback(ctx, evt.InterviewID, evt.OrgID, feedback, transcriptJSON)
	if err != nil {
		return fmt.Errorf("save feedback: %w", err)
	}

	w.logger.Info("publishing notification", "interview_id", evt.InterviewID)
	// 7. Publish Notification Event
	notif := model.NotificationEvent{
		EventType:   "feedback_ready",
		OrgID:       evt.OrgID,
		InterviewID: evt.InterviewID,
		CandidateID: evt.CandidateID,
		FeedbackID:  feedbackID,
		Timestamp:   time.Now().UTC(),
	}
	b, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	if err := w.producer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(evt.OrgID),
		Value: b,
	}); err != nil {
		return fmt.Errorf("publish notification: %w", err)
	}

	w.logger.Info("successfully processed feedback", "interview_id", evt.InterviewID)
	return nil
}

func (w *FeedbackWorker) generateFeedbackWithRetry(ctx context.Context, prompt string) (*model.FeedbackResponse, error) {
	resp, err := w.llm.GenerateFeedback(ctx, prompt)
	if err == nil {
		return resp, nil
	}
	w.logger.Warn("llm generation failed, retrying once", "error", err)
	// Retry once
	return w.llm.GenerateFeedback(ctx, prompt)
}

func (w *FeedbackWorker) validateFeedback(feedback *model.FeedbackResponse, docket *model.Docket) error {
	if feedback.OverallScore < 0 || feedback.OverallScore > 10 {
		return fmt.Errorf("overall score out of bounds")
	}

	switch feedback.Recommendation {
	case "strong_hire", "hire", "no_hire", "strong_no_hire":
	default:
		return fmt.Errorf("invalid recommendation: %s", feedback.Recommendation)
	}

	if len(feedback.Scores) != len(docket.Rubric) {
		return fmt.Errorf("expected %d scores, got %d", len(docket.Rubric), len(feedback.Scores))
	}

	rubricMap := make(map[string]model.RubricItem)
	for _, r := range docket.Rubric {
		rubricMap[r.ID] = r
	}

	seenCriteria := make(map[string]struct{}, len(feedback.Scores))
	for _, s := range feedback.Scores {
		if _, seen := seenCriteria[s.CriterionID.String()]; seen {
			return fmt.Errorf("duplicate score for criterion %s", s.CriterionID)
		}
		seenCriteria[s.CriterionID.String()] = struct{}{}
		r, ok := rubricMap[s.CriterionID.String()]
		if !ok {
			return fmt.Errorf("unknown criterion id: %s", s.CriterionID)
		}
		if s.Score < 0 || s.Score > r.MaxScore {
			return fmt.Errorf("score out of bounds for criterion %s: %d", r.ID, s.Score)
		}
		if strings.TrimSpace(s.Evidence) == "" {
			return fmt.Errorf("empty evidence for criterion %s", r.ID)
		}
	}

	return nil
}

func (w *FeedbackWorker) buildPrompt(docket *model.Docket, candName string, transcript string) string {
	var sb strings.Builder
	sb.WriteString("You are evaluating a completed job interview. Analyze the transcript and generate structured feedback.\n\n")
	sb.WriteString(fmt.Sprintf("JOB ROLE: %s\nCANDIDATE: %s\n\n", docket.Title, candName))

	sb.WriteString("RUBRIC CRITERIA:\n")
	for _, crit := range docket.Rubric {
		sb.WriteString(fmt.Sprintf("- [%s] %s (max %d points): %s\n", crit.ID, crit.Name, crit.MaxScore, crit.Description))
	}

	sb.WriteString("\nFULL INTERVIEW TRANSCRIPT:\n")
	sb.WriteString(transcript)

	sb.WriteString(`

Generate a JSON response with EXACTLY this structure (no other text, just JSON):
{
    "summary": "2-3 paragraph overall assessment of the candidate",
    "overall_score": <number between 0 and 10, one decimal place>,
    "recommendation": "<one of: strong_hire, hire, no_hire, strong_no_hire>",
    "scores": [
        {
            "criterion_id": "<uuid from rubric>",
            "score": <integer between 0 and max_score>,
            "evidence": "<direct quote or reference from transcript supporting this score>"
        }
    ]
}

Be specific. Every score must have evidence from the transcript.`)

	return sb.String()
}

func (w *FeedbackWorker) reconstructTranscript(ctx context.Context, interviewID string, turnCount int) ([]model.TurnEvent, string, error) {
	// Fetch partitions to determine which one contains our interviewID
	conn, err := w.dialer.DialContext(ctx, "tcp", w.brokers[0])
	if err != nil {
		return nil, "", fmt.Errorf("dial kafka to get partitions: %w", err)
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions("interview.turns")
	if err != nil {
		return nil, "", fmt.Errorf("read partitions: %w", err)
	}

	var parts []int
	for _, p := range partitions {
		parts = append(parts, p.ID)
	}
	partition := (&kafka.Hash{}).Balance(kafka.Message{Key: []byte(interviewID)}, parts...)

	w.logger.Info("reconstructTranscript determined partition", "interview_id", interviewID, "partition", partition)

	// A new reader reading from offset 0 to collect all turns for this interview.
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   w.brokers,
		Topic:     "interview.turns",
		Partition: partition,
		Dialer:    w.dialer,
		MinBytes:  1,
		MaxBytes:  10e6,
		MaxWait:   time.Second,
	})
	defer reader.Close()

	if err := reader.SetOffset(kafka.FirstOffset); err != nil {
		return nil, "", fmt.Errorf("set offset: %w", err)
	}

	// Drain approach: read all events for this interview, stop when no new
	// message arrives for 3 seconds (end-of-partition). This is robust against
	// counting mismatches (e.g. initial greeting publishing only 1 Kafka event).
	minEvents := turnCount // at minimum we expect at least this many events
	w.logger.Info("reconstructTranscript started", "interview_id", interviewID, "min_expected", minEvents)

	var turns []model.TurnEvent
	overallTimeout := time.After(60 * time.Second)
	idleTimeout := 3 * time.Second

	for {
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-overallTimeout:
			if len(turns) == 0 {
				return nil, "", fmt.Errorf("timeout: no transcript events found")
			}
			w.logger.Warn("overall timeout reached, proceeding with partial transcript", "interview_id", interviewID, "count", len(turns))
			goto done
		default:
		}

		ctxFetch, cancel := context.WithTimeout(ctx, idleTimeout)
		m, err := reader.FetchMessage(ctxFetch)
		cancel()
		if err != nil {
			// Idle timeout hit - no more messages in partition
			if len(turns) >= minEvents {
				// We have at least the minimum, we are done
				goto done
			}
			// Not enough yet, keep waiting up to overall timeout
			continue
		}

		if string(m.Key) == interviewID {
			var t model.TurnEvent
			if err := json.Unmarshal(m.Value, &t); err != nil {
				w.logger.Warn("unmarshal turn", "error", err)
				continue
			}
			// Skip pure system/instruction messages (no real content role "user" that is the hidden prompt)
			turns = append(turns, t)
			w.logger.Info("found turn", "interview_id", interviewID, "count", len(turns), "role", t.Role)
		}
	}

done:
	transcript := formatTranscript(turns)
	w.logger.Info("reconstructTranscript finished", "interview_id", interviewID, "total_turns", len(turns))
	return turns, transcript, nil
}


func formatTranscript(turns []model.TurnEvent) string {
	// Sort turns by TurnNumber since that represents the chronological sequence number.
	sort.SliceStable(turns, func(i, j int) bool {
		return turns[i].TurnNumber < turns[j].TurnNumber
	})

	var sb strings.Builder
	for _, t := range turns {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", t.Role, t.Content))
	}
	return sb.String()
}

func (w *FeedbackWorker) Close() error {
	w.producer.Close()
	return w.reader.Close()
}
