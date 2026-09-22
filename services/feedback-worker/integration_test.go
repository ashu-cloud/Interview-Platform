package main_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/client"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/llm"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/repository"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/worker"
)

type CountingMockLLM struct {
	client *llm.MockFeedbackLLMClient
	Calls  int
}

func (m *CountingMockLLM) GenerateFeedback(ctx context.Context, prompt string) (*model.FeedbackResponse, error) {
	m.Calls++
	return m.client.GenerateFeedback(ctx, prompt)
}

func TestFeedbackWorker_Integration(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/interview_platform?sslmode=disable"
	}
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:9092"
	}

	// 0. Clean up Kafka topics for the in-process worker test. Do not delete
	// topics while verifying a long-running Docker worker, because its consumer
	// group is intentionally already subscribed to them.
	if os.Getenv("EXTERNAL_FEEDBACK_WORKER") != "true" {
		conn, err := kafka.Dial("tcp", kafkaBrokers)
		if err == nil {
			conn.DeleteTopics("interview.completed", "interview.turns")
			conn.Close()
			time.Sleep(2 * time.Second) // allow broker to clean up
		}
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	// 1. Create DB fixtures for testing
	orgID := uuid.New().String()
	jobRoleID := uuid.New().String()
	docketID := uuid.New().String()
	candidateID := uuid.New().String()
	interviewID := uuid.New().String()

	_, err = pool.Exec(ctx, `INSERT INTO organizations (id, name, slug) VALUES ($1, 'Test Org', $2) ON CONFLICT DO NOTHING`, orgID, orgID)
	if err != nil {
		t.Fatalf("insert org: %v", err)
	}

	userID := uuid.New().String()
	_, err = pool.Exec(ctx, `INSERT INTO users (id, org_id, full_name, email, password_hash) VALUES ($1, $2, 'Test User', $3, 'hash')`, userID, orgID, userID+"@test.com")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	_, err = pool.Exec(ctx, `INSERT INTO job_roles (id, org_id, title, description, created_by) VALUES ($1, $2, 'Backend Eng', 'Desc', $3)`, jobRoleID, orgID, userID)
	if err != nil {
		t.Fatalf("insert job_role: %v", err)
	}

	_, err = pool.Exec(ctx, `INSERT INTO dockets (id, job_role_id, org_id, title, instructions) VALUES ($1, $2, $3, 'Test Docket', 'Instructions')`, docketID, jobRoleID, orgID)
	if err != nil {
		t.Fatalf("insert docket: %v", err)
	}

	critID := uuid.New().String()
	_, err = pool.Exec(ctx, `INSERT INTO rubric_criteria (id, docket_id, name, description, max_score, order_index) VALUES ($1, $2, 'Code Quality', 'desc', 10, 1)`, critID, docketID)
	if err != nil {
		t.Fatalf("insert rubric: %v", err)
	}

	_, err = pool.Exec(ctx, `INSERT INTO candidates (id, org_id, job_role_id, full_name, email) VALUES ($1, $2, $3, 'Test Candidate', $4)`, candidateID, orgID, jobRoleID, candidateID+"@test.com")
	if err != nil {
		t.Fatalf("insert candidate: %v", err)
	}

	inviteID := uuid.New().String()
	_, err = pool.Exec(ctx, `INSERT INTO interview_invites (id, candidate_id, docket_id, org_id, token, expires_at) VALUES ($1, $2, $3, $4, $5, NOW() + INTERVAL '1 day')`, inviteID, candidateID, docketID, orgID, uuid.New().String())
	if err != nil {
		t.Fatalf("insert invite: %v", err)
	}

	_, err = pool.Exec(ctx, `INSERT INTO interviews (id, invite_id, candidate_id, docket_id, org_id, status) VALUES ($1, $2, $3, $4, $5, 'completed')`, interviewID, inviteID, candidateID, docketID, orgID)
	if err != nil {
		t.Fatalf("insert interview: %v", err)
	}

	// 2. Setup Kafka topics and produce events
	brokers := []string{kafkaBrokers}

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	dialer.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if addr == "kafka:9092" {
			addr = "localhost:9092"
		}
		var netDialer net.Dialer
		return netDialer.DialContext(ctx, network, addr)
	}

	transport := &kafka.Transport{
		Dial: dialer.DialFunc,
	}

	os.Setenv("KAFKA_GROUP_ID", "test-feedback-"+uuid.New().String())

	turnWriter := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  "interview.turns",
		Balancer:               &kafka.Hash{},
		Transport:              transport,
		AllowAutoTopicCreation: true,
		MaxAttempts:            10,
	}
	defer turnWriter.Close()

	// Write 2 conversational turns. Each turn has a user event and an
	// assistant event, matching the orchestrator's Kafka contract.
	turns := []model.TurnEvent{
		{EventType: "turn", InterviewID: interviewID, OrgID: orgID, TurnNumber: 1, Role: "user", Content: "Hi", Timestamp: time.Now()},
		{EventType: "turn", InterviewID: interviewID, OrgID: orgID, TurnNumber: 1, Role: "assistant", Content: "Hello", Timestamp: time.Now()},
		{EventType: "turn", InterviewID: interviewID, OrgID: orgID, TurnNumber: 2, Role: "user", Content: "I built a service.", Timestamp: time.Now()},
		{EventType: "turn", InterviewID: interviewID, OrgID: orgID, TurnNumber: 2, Role: "assistant", Content: "How did you test it?", Timestamp: time.Now()},
	}
	for _, turn := range turns {
		payload, _ := json.Marshal(turn)
		err = writeWithRetry(ctx, turnWriter, kafka.Message{Key: []byte(interviewID), Value: payload})
		if err != nil {
			t.Fatalf("write turns failed: %v", err)
		}
	}

	compWriter := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  "interview.completed",
		Balancer:               &kafka.Hash{},
		Transport:              transport,
		AllowAutoTopicCreation: true,
		MaxAttempts:            10,
	}
	defer compWriter.Close()

	compEvt := model.InterviewCompletedEvent{
		EventType: "completed", InterviewID: interviewID, OrgID: orgID, CandidateID: candidateID, DocketID: docketID, CompletedAt: time.Now(), TurnCount: 2,
	}
	cb, _ := json.Marshal(compEvt)

	// 3. Start an in-process worker, unless a running Docker worker is being verified.
	jcsURL := os.Getenv("JOB_CANDIDATE_SERVICE_URL")
	if jcsURL == "" {
		jcsURL = "http://localhost:8001"
	}
	externalWorker := os.Getenv("EXTERNAL_FEEDBACK_WORKER") == "true"
	jcsClient := client.NewJCSClient(jcsURL, "default-internal-secret-key-change-in-prod")
	mockLLM := &CountingMockLLM{client: llm.NewMockFeedbackLLMClient()}
	repo := repository.NewFeedbackRepository(pool)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var w *worker.FeedbackWorker
	var cancel context.CancelFunc
	if !externalWorker {
		w = worker.NewFeedbackWorker(brokers, jcsClient, mockLLM, repo, logger, dialer)
		ctxWorker, workerCancel := context.WithCancel(ctx)
		cancel = workerCancel
		go func() {
			w.Run(ctxWorker)
		}()
	}

	time.Sleep(2 * time.Second) // wait for consumer group to stabilize

	err = writeWithRetry(ctx, compWriter, kafka.Message{Key: []byte(interviewID), Value: cb})
	if err != nil {
		t.Fatalf("write completed event failed: %v", err)
	}

	// 4. Verify Database state. A freshly started external consumer group can
	// take longer to receive its first assignment than the in-process worker.
	maxPolls := 30
	if externalWorker {
		maxPolls = 120
	}
	var status string
	for i := 0; i < maxPolls; i++ {
		err = pool.QueryRow(ctx, "SELECT status FROM interviews WHERE id = $1", interviewID).Scan(&status)
		if err != nil {
			t.Fatalf("query status failed: %v", err)
		}
		if status == "feedback_ready" {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if status != "feedback_ready" {
		t.Fatalf("expected status feedback_ready, got %s", status)
	}

	if !externalWorker {
		cancel()
		w.Close()
	}

	var count int
	err = pool.QueryRow(ctx, "SELECT count(1) FROM feedback_reports WHERE interview_id = $1", interviewID).Scan(&count)
	if err != nil {
		t.Fatalf("query reports: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 report, got %d", count)
	}

	// 5. Test idempotency (process same event again)
	err = writeWithRetry(ctx, compWriter, kafka.Message{Key: []byte(interviewID), Value: cb})
	if err != nil {
		t.Fatalf("write completed event again failed: %v", err)
	}

	if !externalWorker {
		ctxWorker2, cancel2 := context.WithCancel(ctx)
		os.Setenv("KAFKA_GROUP_ID", "test-feedback-"+uuid.New().String())
		w2 := worker.NewFeedbackWorker(brokers, jcsClient, mockLLM, repo, logger, dialer)
		go func() {
			w2.Run(ctxWorker2)
		}()
		time.Sleep(2 * time.Second)
		cancel2()
		w2.Close()
	} else {
		time.Sleep(2 * time.Second)
	}

	err = pool.QueryRow(ctx, "SELECT count(1) FROM feedback_reports WHERE interview_id = $1", interviewID).Scan(&count)
	if err != nil {
		t.Fatalf("query reports after replay: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected still 1 report, got %d", count)
	}
	if !externalWorker && mockLLM.Calls != 1 {
		t.Fatalf("expected LLM called 1 time, got %d", mockLLM.Calls)
	}
}

func writeWithRetry(ctx context.Context, w *kafka.Writer, m kafka.Message) error {
	var err error
	for i := 0; i < 10; i++ {
		err = w.WriteMessages(ctx, m)
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}
