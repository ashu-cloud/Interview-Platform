package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/cache"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/handler"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/repository"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/service"
)

func TestCanonicalAndCompatibilityHTTPIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("INTEGRATION_TEST is not 1 — skipping HTTP integration test")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL must be set when INTEGRATION_TEST=1")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create DB pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("DB ping: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	jobRepo := repository.NewJobRepository(pool)
	docketRepo := repository.NewDocketRepository(pool)
	candidateRepo := repository.NewCandidateRepository(pool)
	inviteRepo := repository.NewInviteRepository(pool)

	var redisClient = func() *redis.Client { return nil }()
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		client, redisErr := cache.NewClient(redisURL)
		if redisErr == nil {
			redisClient = client
			t.Cleanup(func() { _ = client.Close() })
		} else {
			t.Log("Redis is unavailable; invite flow will use the database fallback")
		}
	}

	const jwtSecret = "http-integration-jwt-secret"
	const internalKey = "http-integration-internal-key"
	authSvc := service.NewAuthService(userRepo, jwtSecret)
	jobSvc := service.NewJobService(jobRepo)
	docketSvc := service.NewDocketService(docketRepo, jobRepo)
	candidateSvc := service.NewCandidateService(candidateRepo, jobRepo)
	inviteSvc := service.NewInviteService(inviteRepo, candidateRepo, docketRepo, redisClient, "http://example.test")

	gin.SetMode(gin.TestMode)
	router := newRouter(routeHandlers{
		auth:      handler.NewAuthHandler(authSvc),
		job:       handler.NewJobHandler(jobSvc),
		docket:    handler.NewDocketHandler(docketSvc),
		candidate: handler.NewCandidateHandler(candidateSvc),
		invite:    handler.NewInviteHandler(inviteSvc),
	}, jwtSecret, internalKey)

	suffix := uuid.NewString()
	org, user, err := userRepo.CreateOrgAndUser(ctx,
		"HTTP Integration Org", "http-integration-"+suffix, "http-"+suffix+"@example.test", "hash", "HTTP Tester",
	)
	if err != nil {
		t.Fatalf("create integration organization: %v", err)
	}
	t.Cleanup(func() { cleanupHTTPIntegration(ctx, pool, org.ID) })
	token := signIntegrationJWT(t, jwtSecret, user.ID, org.ID, user.Email)

	var track model.JobRole
	doJSON(t, router, token, http.MethodPost, "/api/v1/tracks", map[string]any{
		"title": "Platform Engineer", "description": "Build platform", "required_skills": []string{"Go"},
	}, http.StatusCreated, &track)

	var canonicalTracks struct {
		Tracks []model.JobRole `json:"tracks"`
		Total  int64           `json:"total"`
	}
	doJSON(t, router, token, http.MethodGet, "/api/v1/tracks", nil, http.StatusOK, &canonicalTracks)
	if len(canonicalTracks.Tracks) == 0 || canonicalTracks.Total == 0 {
		t.Fatalf("canonical tracks envelope is empty: %+v", canonicalTracks)
	}

	var fetchedTrack model.JobRole
	doJSON(t, router, token, http.MethodGet, "/api/v1/tracks/"+track.ID.String(), nil, http.StatusOK, &fetchedTrack)
	doJSON(t, router, token, http.MethodPatch, "/api/v1/tracks/"+track.ID.String(), map[string]any{
		"description": "Updated platform",
	}, http.StatusOK, &fetchedTrack)

	var legacyJobs struct {
		Jobs []model.JobRole `json:"jobs"`
	}
	doJSON(t, router, token, http.MethodGet, "/api/v1/jobs", nil, http.StatusOK, &legacyJobs)
	if len(legacyJobs.Jobs) == 0 {
		t.Fatal("legacy jobs envelope is empty")
	}

	var disposableTrack model.JobRole
	doJSON(t, router, token, http.MethodPost, "/api/v1/tracks", map[string]any{
		"title": "Disposable Track",
	}, http.StatusCreated, &disposableTrack)
	doJSON(t, router, token, http.MethodDelete, "/api/v1/tracks/"+disposableTrack.ID.String(), nil, http.StatusNoContent, nil)

	blueprintBody := map[string]any{
		"title": "Platform Assessment", "instructions": "Be concise", "duration_mins": 30,
		"questions": []map[string]any{{"question_text": "Explain a queue", "question_type": "technical", "order_index": 1}},
		"rubric": []map[string]any{{"name": "Systems", "description": "Systems knowledge", "max_score": 10, "order_index": 1}},
	}
	var blueprint model.Docket
	doJSON(t, router, token, http.MethodPost, "/api/v1/tracks/"+track.ID.String()+"/blueprints", blueprintBody, http.StatusCreated, &blueprint)

	var canonicalBlueprints struct {
		Blueprints []model.Docket `json:"blueprints"`
	}
	doJSON(t, router, token, http.MethodGet, "/api/v1/tracks/"+track.ID.String()+"/blueprints", nil, http.StatusOK, &canonicalBlueprints)
	if len(canonicalBlueprints.Blueprints) != 1 {
		t.Fatalf("canonical blueprints length = %d, want 1", len(canonicalBlueprints.Blueprints))
	}

	var fetchedBlueprint model.Docket
	doJSON(t, router, token, http.MethodGet, "/api/v1/blueprints/"+blueprint.ID.String(), nil, http.StatusOK, &fetchedBlueprint)
	doJSON(t, router, token, http.MethodPatch, "/api/v1/blueprints/"+blueprint.ID.String(), map[string]any{
		"title": "Updated Platform Assessment",
	}, http.StatusOK, &fetchedBlueprint)

	var legacyDockets struct {
		Dockets []model.Docket `json:"dockets"`
	}
	doJSON(t, router, token, http.MethodGet, "/api/v1/jobs/"+track.ID.String()+"/dockets", nil, http.StatusOK, &legacyDockets)
	if len(legacyDockets.Dockets) != 1 {
		t.Fatalf("legacy dockets length = %d, want 1", len(legacyDockets.Dockets))
	}

	var disposableBlueprint model.Docket
	disposableBody := blueprintBody
	disposableBody["title"] = "Disposable Blueprint"
	doJSON(t, router, token, http.MethodPost, "/api/v1/tracks/"+track.ID.String()+"/blueprints", disposableBody, http.StatusCreated, &disposableBlueprint)
	doJSON(t, router, token, http.MethodDelete, "/api/v1/blueprints/"+disposableBlueprint.ID.String(), nil, http.StatusNoContent, nil)

	var participant model.Candidate
	doJSON(t, router, token, http.MethodPost, "/api/v1/tracks/"+track.ID.String()+"/participants", map[string]any{
		"email": "candidate-" + suffix + "@example.test", "full_name": "Candidate One",
	}, http.StatusCreated, &participant)

	var canonicalParticipants struct {
		Participants []model.CandidateWithStatus `json:"participants"`
	}
	doJSON(t, router, token, http.MethodGet, "/api/v1/tracks/"+track.ID.String()+"/participants", nil, http.StatusOK, &canonicalParticipants)
	if len(canonicalParticipants.Participants) != 1 {
		t.Fatalf("canonical participants length = %d, want 1", len(canonicalParticipants.Participants))
	}
	doJSON(t, router, token, http.MethodGet, "/api/v1/participants/"+participant.ID.String(), nil, http.StatusOK, &participant)

	var legacyCandidates struct {
		Candidates []model.CandidateWithStatus `json:"candidates"`
	}
	doJSON(t, router, token, http.MethodGet, "/api/v1/jobs/"+track.ID.String()+"/candidates", nil, http.StatusOK, &legacyCandidates)
	if len(legacyCandidates.Candidates) != 1 {
		t.Fatalf("legacy candidates length = %d, want 1", len(legacyCandidates.Candidates))
	}

	var invite model.InviteResponse
	doJSON(t, router, token, http.MethodPost, "/api/v1/participants/"+participant.ID.String()+"/invite", map[string]any{
		"docket_id": blueprint.ID.String(), "expires_in_days": 1,
	}, http.StatusCreated, &invite)

	var session model.CreateSessionResponse
	doJSON(t, router, token, http.MethodPost, "/api/v1/sessions", map[string]any{
		"invite_id": invite.InviteID,
	}, http.StatusCreated, &session)
	var reusedSession model.CreateSessionResponse
	doJSON(t, router, token, http.MethodPost, "/api/v1/sessions", map[string]any{
		"invite_id": invite.InviteID,
	}, http.StatusCreated, &reusedSession)
	if reusedSession.Session.ID != session.Session.ID {
		t.Fatalf("session POST was not idempotent: %v vs %v", reusedSession.Session.ID, session.Session.ID)
	}

	var sessionDetail model.InterviewDetail
	doJSON(t, router, token, http.MethodGet, "/api/v1/sessions/"+session.Session.ID.String(), nil, http.StatusOK, &sessionDetail)

	if _, err := pool.Exec(ctx,
		`INSERT INTO session_turns (session_id, org_id, turn_number, role, content)
		 VALUES ($1, $2, 1, 'assessor', 'first'), ($1, $2, 3, 'participant', 'third')`,
		session.Session.ID, org.ID,
	); err != nil {
		t.Fatalf("insert sparse session turns: %v", err)
	}
	var turnsEnvelope struct {
		Turns []model.SessionTurn `json:"turns"`
	}
	doJSON(t, router, token, http.MethodGet, "/api/v1/sessions/"+session.Session.ID.String()+"/turns", nil, http.StatusOK, &turnsEnvelope)
	if len(turnsEnvelope.Turns) != 2 {
		t.Fatalf("turns length = %d, want 2", len(turnsEnvelope.Turns))
	}

	doJSONWithHeader(t, router, "X-Internal-Key", internalKey, http.MethodPatch,
		"/api/v1/interviews/"+session.Session.ID.String()+"/status", map[string]any{"status": "completed"}, http.StatusOK, nil)

	var evaluation model.EvaluationJob
	doJSON(t, router, token, http.MethodPost, "/api/v1/sessions/"+session.Session.ID.String()+"/evaluation", nil, http.StatusAccepted, &evaluation)
	if evaluation.TurnCount != 3 {
		t.Fatalf("evaluation turn_count = %d, want MAX(turn_number) = 3", evaluation.TurnCount)
	}
	var reusedEvaluation model.EvaluationJob
	doJSON(t, router, token, http.MethodPost, "/api/v1/sessions/"+session.Session.ID.String()+"/evaluation", nil, http.StatusAccepted, &reusedEvaluation)
	if reusedEvaluation.ID != evaluation.ID {
		t.Fatalf("evaluation POST created duplicate jobs: %v vs %v", reusedEvaluation.ID, evaluation.ID)
	}

	var legacyInterviews struct {
		Interviews []model.InterviewDetail `json:"interviews"`
	}
	doJSON(t, router, token, http.MethodGet, "/api/v1/interviews", nil, http.StatusOK, &legacyInterviews)
	if len(legacyInterviews.Interviews) == 0 {
		t.Fatal("legacy interviews envelope is empty")
	}

	doJSON(t, router, token, http.MethodDelete, "/api/v1/blueprints/"+blueprint.ID.String(), nil, http.StatusConflict, nil)
	doJSON(t, router, token, http.MethodDelete, "/api/v1/tracks/"+track.ID.String(), nil, http.StatusConflict, nil)
}

func signIntegrationJWT(t *testing.T, secret string, userID, orgID uuid.UUID, email string) string {
	t.Helper()
	claims := model.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		OrgID: orgID.String(),
		Email: email,
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign JWT: %v", err)
	}
	return token
}

func doJSON(t *testing.T, router http.Handler, token, method, path string, body any, wantStatus int, out any) {
	t.Helper()
	doJSONWithHeader(t, router, "Authorization", "Bearer "+token, method, path, body, wantStatus, out)
}

func doJSONWithHeader(t *testing.T, router http.Handler, header, value, method, path string, body any, wantStatus int, out any) {
	t.Helper()
	var encoded []byte
	var err error
	if body != nil {
		encoded, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal %s %s body: %v", method, path, err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(encoded))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if value != "" {
		req.Header.Set(header, value)
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != wantStatus {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, path, resp.Code, wantStatus, resp.Body.String())
	}
	if out != nil && resp.Body.Len() > 0 {
		if err := json.Unmarshal(resp.Body.Bytes(), out); err != nil {
			t.Fatalf("decode %s %s response: %v; body=%s", method, path, err, resp.Body.String())
		}
	}
}

func cleanupHTTPIntegration(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID) {
	statements := []string{
		"DELETE FROM outbox_completed WHERE org_id = $1",
		"DELETE FROM session_turns WHERE org_id = $1",
		"DELETE FROM assessment_sessions WHERE org_id = $1",
		"DELETE FROM assessment_invites WHERE org_id = $1",
		"DELETE FROM participants WHERE org_id = $1",
		"DELETE FROM assessment_blueprints WHERE org_id = $1",
		"DELETE FROM assessment_tracks WHERE org_id = $1",
		"DELETE FROM users WHERE org_id = $1",
		"DELETE FROM organizations WHERE id = $1",
	}
	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement, orgID); err != nil {
			fmt.Fprintf(os.Stderr, "HTTP integration cleanup failed\n")
			return
		}
	}
}
