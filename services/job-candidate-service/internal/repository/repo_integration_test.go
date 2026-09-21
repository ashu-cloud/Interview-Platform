package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// setupTestDB connects to the real Postgres instance.
// Skips the test if DATABASE_URL is not set, keeping `go test ./...` green
// without Docker. Run with DATABASE_URL set to execute integration tests.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set — skipping integration test")
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
		_, _ = pool.Exec(ctx, "DELETE FROM job_roles WHERE id = $1", job.ID)
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
}
