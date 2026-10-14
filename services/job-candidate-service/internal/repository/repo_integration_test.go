package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// setupTestDB connects to the real Postgres instance only when explicitly enabled.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("INTEGRATION_TEST is not 1 — skipping integration test")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Fatal("DATABASE_URL must be set when INTEGRATION_TEST=1")
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("DB ping: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

// TestUserRepository_CreateAndGet exercises the full create-org-and-user → get-by-email round-trip.
func TestUserRepository_CreateAndGet(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	org, user, err := repo.CreateOrgAndUser(ctx,
		"Test Org", "testorg-"+t.Name(), "test@example.com", "hashed_pw", "Test User",
	)
	if err != nil {
		t.Fatalf("CreateOrgAndUser: %v", err)
	}
	if org.ID.String() == "" {
		t.Fatal("org ID is empty")
	}
	if user.Email != "test@example.com" {
		t.Errorf("user.Email = %q, want %q", user.Email, "test@example.com")
	}

	// Clean up
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", org.ID)
	}()

	// Verify GetByEmail
	fetched, err := repo.GetByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if fetched.ID != user.ID {
		t.Errorf("fetched.ID = %v, want %v", fetched.ID, user.ID)
	}
}

// TestJobRepository_CRUD exercises create → list → get → patch for job roles.
func TestJobRepository_CRUD(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()

	// Need an org + user first
	userRepo := NewUserRepository(pool)
	org, user, err := userRepo.CreateOrgAndUser(ctx,
		"Job Test Org", "jobtestorg-"+t.Name(), "jobtest@example.com", "pw", "Job Tester",
	)
	if err != nil {
		t.Fatalf("setup org/user: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", org.ID)
	}()

	repo := NewJobRepository(pool)

	// Create
	job, err := repo.Create(ctx, org.ID, user.ID, &model.CreateJobRequest{
		Title:          "Software Engineer",
		Description:    "Build things",
		RequiredSkills: []string{"Go", "PostgreSQL"},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if job.Title != "Software Engineer" {
		t.Errorf("Title = %q, want %q", job.Title, "Software Engineer")
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM assessment_tracks WHERE id = $1", job.ID)
	}()

	// GetByID
	fetched, err := repo.GetByID(ctx, org.ID, job.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if fetched.ID != job.ID {
		t.Errorf("fetched.ID = %v, want %v", fetched.ID, job.ID)
	}

	// List
	jobs, total, err := repo.List(ctx, org.ID, "", 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 1 {
		t.Errorf("total = %d, want >= 1", total)
	}
	if len(jobs) < 1 {
		t.Error("expected at least 1 job in list")
	}

	// Patch
	newTitle := "Senior Software Engineer"
	patched, err := repo.Patch(ctx, org.ID, job.ID, &model.PatchJobRequest{
		Title: &newTitle,
	})
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
	if patched.Title != newTitle {
		t.Errorf("patched.Title = %q, want %q", patched.Title, newTitle)
	}

	docketRepo := NewDocketRepository(pool)
	docket, err := docketRepo.Create(ctx, org.ID, job.ID, &model.CreateDocketRequest{
		Title:        "Disposable Blueprint",
		DurationMins: 15,
		Questions: []model.QuestionInput{
			{QuestionText: "Question", QuestionType: "technical", OrderIndex: 1},
		},
		Rubric: []model.RubricInput{
			{Name: "Criterion", Description: "Description", MaxScore: 5, OrderIndex: 1},
		},
	})
	if err != nil {
		t.Fatalf("create blueprint for delete: %v", err)
	}
	if err := docketRepo.Delete(ctx, org.ID, docket.ID); err != nil {
		t.Fatalf("delete blueprint: %v", err)
	}
	if _, err := docketRepo.GetByID(ctx, org.ID, docket.ID); err == nil {
		t.Fatal("GetByID blueprint after Delete returned nil error")
	}

	if err := repo.Delete(ctx, org.ID, job.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, org.ID, job.ID); err == nil {
		t.Fatal("GetByID after Delete returned nil error")
	}
}

// TestInviteRepository_ReconnectionLifecycle tests invite creation, in_progress reconnect, and completion.
func TestInviteRepository_ReconnectionLifecycle(t *testing.T) {
	pool := setupTestDB(t)
	ctx := context.Background()

	userRepo := NewUserRepository(pool)
	org, user, err := userRepo.CreateOrgAndUser(ctx,
		"Invite Test Org", "invitetestorg-"+t.Name(), "invite-tester@example.com", "pw", "Invite Tester",
	)
	if err != nil {
		t.Fatalf("setup org/user: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", org.ID)
	}()

	jobRepo := NewJobRepository(pool)
	job, err := jobRepo.Create(ctx, org.ID, user.ID, &model.CreateJobRequest{
		Title: "Backend Developer",
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	docketRepo := NewDocketRepository(pool)
	docket, err := docketRepo.Create(ctx, org.ID, job.ID, &model.CreateDocketRequest{
		Title:        "Go Coding Interview",
		DurationMins: 30,
		Questions: []model.QuestionInput{
			{QuestionText: "Explain concurrency in Go", QuestionType: "technical", OrderIndex: 1},
		},
		Rubric: []model.RubricInput{
			{Name: "Go Knowledge", Description: "Understanding of channels/goroutines", MaxScore: 10, OrderIndex: 1},
		},
	})
	if err != nil {
		t.Fatalf("create docket: %v", err)
	}

	candidateRepo := NewCandidateRepository(pool)
	candidate, err := candidateRepo.Create(ctx, org.ID, job.ID, &model.CreateCandidateRequest{
		Email:    "applicant@example.com",
		FullName: "Alice Candidate",
	})
	if err != nil {
		t.Fatalf("create candidate: %v", err)
	}

	inviteRepo := NewInviteRepository(pool)
	token := "test-token-" + t.Name()
	expiresAt := time.Now().Add(24 * time.Hour)

	invite, interview, err := inviteRepo.CreateInviteAndInterview(ctx, candidate.ID, docket.ID, org.ID, token, expiresAt)
	if err != nil {
		t.Fatalf("CreateInviteAndInterview: %v", err)
	}

	reused, err := inviteRepo.CreateSession(ctx, org.ID, invite.ID)
	if err != nil {
		t.Fatalf("CreateSession reuse: %v", err)
	}
	if reused.ID != interview.ID {
		t.Fatalf("CreateSession reused ID = %v, want %v", reused.ID, interview.ID)
	}
	reusedAgain, err := inviteRepo.CreateSession(ctx, org.ID, invite.ID)
	if err != nil || reusedAgain.ID != interview.ID {
		t.Fatalf("CreateSession second reuse = %v, %v", reusedAgain, err)
	}
	if _, err := inviteRepo.CreateSession(ctx, uuid.New(), invite.ID); err == nil {
		t.Fatal("CreateSession with another org returned nil error")
	}

	var sessionlessInviteID uuid.UUID
	err = pool.QueryRow(ctx,
		`INSERT INTO assessment_invites (candidate_id, docket_id, org_id, token, expires_at)
		 VALUES ($1, $2, $3, $4, NOW() + INTERVAL '1 day') RETURNING id`,
		candidate.ID, docket.ID, org.ID, "sessionless-"+uuid.NewString(),
	).Scan(&sessionlessInviteID)
	if err != nil {
		t.Fatalf("create sessionless invite: %v", err)
	}
	createdSession, err := inviteRepo.CreateSession(ctx, org.ID, sessionlessInviteID)
	if err != nil {
		t.Fatalf("CreateSession new: %v", err)
	}
	createdSessionAgain, err := inviteRepo.CreateSession(ctx, org.ID, sessionlessInviteID)
	if err != nil || createdSessionAgain.ID != createdSession.ID {
		t.Fatalf("CreateSession new invite idempotency = %v, %v", createdSessionAgain, err)
	}

	// 1. Initial lookup -> invited
	details, err := inviteRepo.GetByToken(ctx, token)
	if err != nil {
		t.Fatalf("GetByToken (initial): %v", err)
	}
	if details.InterviewStatus != "invited" {
		t.Errorf("InterviewStatus = %q, want 'invited'", details.InterviewStatus)
	}
	if details.InterviewID != interview.ID {
		t.Errorf("InterviewID = %v, want %v", details.InterviewID, interview.ID)
	}

	if fetchedCandidate, err := candidateRepo.GetByID(ctx, org.ID, candidate.ID); err != nil || fetchedCandidate.ID != candidate.ID {
		t.Fatalf("GetByID participant = %v, %v", fetchedCandidate, err)
	}
	blueprints, err := docketRepo.ListByJob(ctx, org.ID, job.ID)
	if err != nil || len(blueprints) != 1 {
		t.Fatalf("ListByJob blueprints length = %d, err = %v", len(blueprints), err)
	}
	updatedTitle := "Updated Go Interview"
	if updated, err := docketRepo.Update(ctx, org.ID, docket.ID, &model.UpdateDocketRequest{Title: &updatedTitle}); err != nil || updated.Title != updatedTitle {
		t.Fatalf("Update blueprint = %v, %v", updated, err)
	}
	replacementQuestions := []model.QuestionInput{
		{QuestionText: "Replacement", QuestionType: "technical", OrderIndex: 1},
	}
	if _, err := docketRepo.Update(ctx, org.ID, docket.ID, &model.UpdateDocketRequest{Questions: &replacementQuestions}); err == nil {
		t.Fatal("replacing prompts after session creation returned nil error")
	}
	if err := docketRepo.Delete(ctx, org.ID, docket.ID); err == nil {
		t.Fatal("deleting blueprint with a session returned nil error")
	}
	if err := jobRepo.Delete(ctx, org.ID, job.ID); err == nil {
		t.Fatal("deleting track with dependent records returned nil error")
	}

	// 2. Start interview -> in_progress
	if _, err := inviteRepo.UpdateInterviewStatus(ctx, interview.ID, "in_progress"); err != nil {
		t.Fatalf("UpdateInterviewStatus(in_progress): %v", err)
	}

	// 3. Reconnect lookup -> still valid, in_progress
	details, err = inviteRepo.GetByToken(ctx, token)
	if err != nil {
		t.Fatalf("GetByToken (reconnect): %v", err)
	}
	if details.InterviewStatus != "in_progress" {
		t.Errorf("InterviewStatus = %q, want 'in_progress'", details.InterviewStatus)
	}
	if details.UsedAt != nil {
		t.Error("expected UsedAt to remain nil while interview is in_progress")
	}

	if _, err := pool.Exec(ctx,
		`INSERT INTO session_turns (session_id, org_id, turn_number, role, content) VALUES ($1, $2, 1, 'participant', 'answer')`,
		interview.ID, org.ID,
	); err != nil {
		t.Fatalf("insert session turn: %v", err)
	}
	turns, err := inviteRepo.ListSessionTurns(ctx, org.ID, interview.ID)
	if err != nil || len(turns) != 1 {
		t.Fatalf("ListSessionTurns length = %d, err = %v", len(turns), err)
	}

	// 4. Complete interview -> completed
	if _, err := inviteRepo.UpdateInterviewStatus(ctx, interview.ID, "completed"); err != nil {
		t.Fatalf("UpdateInterviewStatus(completed): %v", err)
	}

	// 5. GetInviteByInterviewID
	byInterview, err := inviteRepo.GetInviteByInterviewID(ctx, interview.ID)
	if err != nil {
		t.Fatalf("GetInviteByInterviewID: %v", err)
	}
	if byInterview.ID != invite.ID {
		t.Errorf("invite ID mismatch: %v vs %v", byInterview.ID, invite.ID)
	}

	evaluation, err := inviteRepo.EnqueueEvaluation(ctx, org.ID, interview.ID)
	if err != nil {
		t.Fatalf("EnqueueEvaluation: %v", err)
	}
	if evaluation.SessionID != interview.ID || evaluation.TurnCount != 1 || evaluation.Status != "pending" {
		t.Errorf("unexpected evaluation job: %+v", evaluation)
	}
	reusedEvaluation, err := inviteRepo.EnqueueEvaluation(ctx, org.ID, interview.ID)
	if err != nil {
		t.Fatalf("EnqueueEvaluation reuse: %v", err)
	}
	if reusedEvaluation.ID != evaluation.ID {
		t.Errorf("EnqueueEvaluation created duplicate job: %v vs %v", reusedEvaluation.ID, evaluation.ID)
	}
}
