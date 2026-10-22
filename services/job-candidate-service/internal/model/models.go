package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ─── Domain Models ───────────────────────────────────────────────────────────

// Organization represents a tenant.
type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents a hiring manager.
type User struct {
	ID           uuid.UUID `json:"id"`
	OrgID        uuid.UUID `json:"org_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// JobRole represents an open position.
type JobRole struct {
	ID             uuid.UUID `json:"id"`
	OrgID          uuid.UUID `json:"org_id"`
	CreatedBy      uuid.UUID `json:"created_by"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	RequiredSkills []string  `json:"required_skills"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Docket represents an interview question bank and rubric set.
type Docket struct {
	ID           uuid.UUID        `json:"id"`
	JobRoleID    uuid.UUID        `json:"job_role_id"`
	OrgID        uuid.UUID        `json:"org_id"`
	Title        string           `json:"title"`
	Instructions string           `json:"instructions"`
	DurationMins int              `json:"duration_mins"`
	Questions    []DocketQuestion `json:"questions,omitempty"`
	Rubric       []RubricCriteria `json:"rubric,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

// DocketQuestion is one question within a docket.
type DocketQuestion struct {
	ID           uuid.UUID `json:"id"`
	DocketID     uuid.UUID `json:"docket_id"`
	QuestionText string    `json:"question_text"`
	QuestionType string    `json:"question_type"`
	OrderIndex   int       `json:"order_index"`
	FollowUpHint *string   `json:"follow_up_hint,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// RubricCriteria defines a scoring dimension.
type RubricCriteria struct {
	ID          uuid.UUID `json:"id"`
	DocketID    uuid.UUID `json:"docket_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MaxScore    int       `json:"max_score"`
	OrderIndex  int       `json:"order_index"`
	CreatedAt   time.Time `json:"created_at"`
}

// DocumentChunk is one embedded piece of a blueprint's RAG knowledge base.
type DocumentChunk struct {
	BlueprintID uuid.UUID
	OrgID       uuid.UUID
	ChunkText   string
	Embedding   []float32
	ChunkIndex  int
	Source      string
}

// Candidate represents an applicant.
type Candidate struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	JobRoleID uuid.UUID `json:"job_role_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	CreatedAt time.Time `json:"created_at"`
}

// CandidateWithStatus includes the candidate's latest interview status.
type CandidateWithStatus struct {
	ID              uuid.UUID `json:"id"`
	OrgID           uuid.UUID `json:"org_id"`
	JobRoleID       uuid.UUID `json:"job_role_id"`
	Email           string    `json:"email"`
	FullName        string    `json:"full_name"`
	CreatedAt       time.Time `json:"created_at"`
	InterviewStatus string    `json:"interview_status,omitempty"`
}

// InterviewInvite is a unique time-bound invite link.
type InterviewInvite struct {
	ID          uuid.UUID  `json:"id"`
	CandidateID uuid.UUID  `json:"candidate_id"`
	DocketID    uuid.UUID  `json:"docket_id"`
	OrgID       uuid.UUID  `json:"org_id"`
	Token       string     `json:"token"`
	ExpiresAt   time.Time  `json:"expires_at"`
	UsedAt      *time.Time `json:"used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Interview tracks an interview session.
type Interview struct {
	ID          uuid.UUID  `json:"id"`
	InviteID    uuid.UUID  `json:"invite_id"`
	CandidateID uuid.UUID  `json:"candidate_id"`
	DocketID    uuid.UUID  `json:"docket_id"`
	OrgID       uuid.UUID  `json:"org_id"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// FeedbackReport is the structured evaluation report.
type FeedbackReport struct {
	ID             uuid.UUID       `json:"id"`
	InterviewID    uuid.UUID       `json:"interview_id"`
	OrgID          uuid.UUID       `json:"org_id"`
	Summary        string          `json:"summary"`
	OverallScore   float64         `json:"overall_score"`
	Recommendation string          `json:"recommendation"`
	TranscriptJSON []byte          `json:"-"` // raw JSONB — decoded by caller
	Scores         []FeedbackScore `json:"scores,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// FeedbackScore is per-rubric-criterion score.
type FeedbackScore struct {
	ID          uuid.UUID `json:"id"`
	FeedbackID  uuid.UUID `json:"feedback_id"`
	CriterionID uuid.UUID `json:"criterion_id"`
	Score       int       `json:"score"`
	Evidence    string    `json:"evidence"`
	CreatedAt   time.Time `json:"created_at"`
}

// ─── Auth DTOs ────────────────────────────────────────────────────────────────

// RegisterRequest is the request body for POST /api/v1/auth/register.
type RegisterRequest struct {
	OrgName  string `json:"org_name"  validate:"required"`
	OrgSlug  string `json:"org_slug"  validate:"required,alphanum,max=100"`
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=8"`
}

// RegisterResponse is the response body for POST /api/v1/auth/register.
type RegisterResponse struct {
	UserID string `json:"user_id"`
	OrgID  string `json:"org_id"`
	Token  string `json:"token"`
}

// LoginRequest is the request body for POST /api/v1/auth/login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is the response body for POST /api/v1/auth/login.
type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	OrgID  string `json:"org_id"`
}

// ─── Job Role DTOs ────────────────────────────────────────────────────────────

// CreateJobRequest is the request body for POST /api/v1/jobs.
type CreateJobRequest struct {
	Title          string   `json:"title"           validate:"required"`
	Description    string   `json:"description"`
	RequiredSkills []string `json:"required_skills"`
}

// PatchJobRequest is the request body for PATCH /api/v1/jobs/:job_id.
// All fields are optional.
type PatchJobRequest struct {
	Title          *string  `json:"title"`
	Description    *string  `json:"description"`
	RequiredSkills []string `json:"required_skills"`
	Status         *string  `json:"status" validate:"omitempty,oneof=active archived"`
}

// JobListResponse is the paginated response for GET /api/v1/jobs.
type JobListResponse struct {
	Jobs  []JobRole `json:"jobs"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

// ─── Docket DTOs ──────────────────────────────────────────────────────────────

// QuestionInput is a single question within CreateDocketRequest.
type QuestionInput struct {
	QuestionText string  `json:"question_text" validate:"required"`
	QuestionType string  `json:"question_type" validate:"required,oneof=technical behavioral situational"`
	OrderIndex   int     `json:"order_index"   validate:"required,min=1"`
	FollowUpHint *string `json:"follow_up_hint"`
}

// RubricInput is a single rubric criterion within CreateDocketRequest.
type RubricInput struct {
	Name        string `json:"name"        validate:"required"`
	Description string `json:"description" validate:"required"`
	MaxScore    int    `json:"max_score"   validate:"required,min=1"`
	OrderIndex  int    `json:"order_index" validate:"required,min=1"`
}

// CreateDocketRequest is the request body for POST /api/v1/jobs/:job_id/dockets.
type CreateDocketRequest struct {
	Title        string          `json:"title"         validate:"required"`
	Instructions string          `json:"instructions"`
	DurationMins int             `json:"duration_mins" validate:"required,min=1"`
	Questions    []QuestionInput `json:"questions"     validate:"required,min=1,dive"`
	Rubric       []RubricInput   `json:"rubric"        validate:"required,min=1,dive"`
}

// UpdateDocketRequest is the request body for a partial blueprint update.
type UpdateDocketRequest struct {
	Title        *string          `json:"title" validate:"omitempty,min=1"`
	Instructions *string          `json:"instructions"`
	DurationMins *int             `json:"duration_mins" validate:"omitempty,min=1"`
	Questions    *[]QuestionInput `json:"questions" validate:"omitempty,min=1,dive"`
	Rubric       *[]RubricInput   `json:"rubric" validate:"omitempty,min=1,dive"`
}

// ─── Candidate DTOs ───────────────────────────────────────────────────────────

// CreateCandidateRequest is the request body for POST /api/v1/jobs/:job_id/candidates.
type CreateCandidateRequest struct {
	Email    string `json:"email"     validate:"required,email"`
	FullName string `json:"full_name" validate:"required"`
}

// CreateSessionRequest creates or reuses the session for an existing invite.
type CreateSessionRequest struct {
	InviteID string `json:"invite_id" validate:"required,uuid"`
}

// CreateSessionResponse is the canonical assessment envelope for a session.
type CreateSessionResponse struct {
	Session Interview `json:"session"`
}

// ─── Invite DTOs ──────────────────────────────────────────────────────────────

// CreateInviteRequest is the request body for POST /api/v1/candidates/:candidate_id/invite.
type CreateInviteRequest struct {
	DocketID      string `json:"docket_id"       validate:"required,uuid"`
	ExpiresInDays int    `json:"expires_in_days" validate:"required,min=1,max=30"`
}

// InviteResponse is the response body for POST /api/v1/candidates/:candidate_id/invite.
type InviteResponse struct {
	InviteID   string    `json:"invite_id"`
	InviteLink string    `json:"invite_link"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// PublicApplyRequest is the request body for POST /api/v1/public/tracks/:job_id/apply.
type PublicApplyRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Email    string `json:"email"     validate:"required,email"`
}

// PublicTrackResponse is the response payload for GET /api/v1/public/tracks/:job_id.
type PublicTrackResponse struct {
	Track          JobRole `json:"track"`
	BlueprintID    string  `json:"blueprint_id,omitempty"`
	BlueprintTitle string  `json:"blueprint_title,omitempty"`
	DurationMins   int     `json:"duration_mins,omitempty"`
}

// InviteValidateCandidate is the nested candidate info in InviteValidateResponse.
type InviteValidateCandidate struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}

// InviteValidateJobRole is the nested job role info in InviteValidateResponse.
type InviteValidateJobRole struct {
	Title string `json:"title"`
}

// InviteValidateResponse is the response body for GET /api/v1/invites/validate/:token.
type InviteValidateResponse struct {
	Valid           bool                    `json:"valid"`
	InterviewID     string                  `json:"interview_id"`
	InterviewStatus string                  `json:"interview_status"`
	DocketID        string                  `json:"docket_id"`
	OrgID           string                  `json:"org_id"`
	Candidate       InviteValidateCandidate `json:"candidate"`
	JobRole         InviteValidateJobRole   `json:"job_role"`
	DocketTitle     string                  `json:"docket_title"`
	ExpiresAt       time.Time               `json:"expires_at"`
}

// InviteCacheValue is the JSON shape stored in Redis under invite:token:{token}.
// This matches the spec Redis key schema and includes interview info for reconnection.
type InviteCacheValue struct {
	InviteID    string `json:"invite_id"`
	InterviewID string `json:"interview_id,omitempty"`
	CandidateID string `json:"candidate_id"`
	DocketID    string `json:"docket_id"`
	OrgID       string `json:"org_id"`
	ExpiresAt   string `json:"expires_at"`
}

// ─── Interview DTOs ───────────────────────────────────────────────────────────

// CandidateBrief is a compact candidate representation in interview responses.
type CandidateBrief struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}

// DocketBrief is a compact docket representation in interview responses.
type DocketBrief struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// FeedbackBrief is a compact feedback representation included when status == feedback_ready.
type FeedbackBrief struct {
	ID             string  `json:"id"`
	OverallScore   float64 `json:"overall_score"`
	Recommendation string  `json:"recommendation"`
}

// InterviewDetail is the response for GET /api/v1/interviews/:interview_id.
type InterviewDetail struct {
	ID          string         `json:"id"`
	Status      string         `json:"status"`
	Candidate   CandidateBrief `json:"candidate"`
	Docket      DocketBrief    `json:"docket"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Feedback    *FeedbackBrief `json:"feedback,omitempty"`
}

// InterviewListResponse is the paginated response for GET /api/v1/interviews.
type InterviewListResponse struct {
	Interviews []InterviewDetail `json:"interviews"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
}

// ─── Feedback DTOs ────────────────────────────────────────────────────────────

// FeedbackScoreDetail is per-criterion score detail returned to hiring manager.
type FeedbackScoreDetail struct {
	CriterionName  string  `json:"criterion_name"`
	Score          int     `json:"score"`
	MaxScore       int     `json:"max_score"`
	Evidence       string  `json:"evidence"`
	EvidenceTurnID *string `json:"evidence_turn_id,omitempty"`
	IsGrounded     bool    `json:"is_grounded"`
}

// TranscriptTurn is a single turn in the interview transcript.
type TranscriptTurn struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionTurn is one materialized assessment session turn.
type SessionTurn struct {
	ID         uuid.UUID `json:"id"`
	SessionID  uuid.UUID `json:"session_id"`
	OrgID      uuid.UUID `json:"org_id"`
	TurnNumber int       `json:"turn_number"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

// EvaluationJob is an evaluation request queued in outbox_completed.
type EvaluationJob struct {
	ID            uuid.UUID `json:"id"`
	SessionID     uuid.UUID `json:"session_id"`
	OrgID         uuid.UUID `json:"org_id"`
	ParticipantID uuid.UUID `json:"participant_id"`
	BlueprintID   uuid.UUID `json:"blueprint_id"`
	TurnCount     int       `json:"turn_count"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// FeedbackDetailResponse is the response for GET /api/v1/interviews/:interview_id/feedback.
type FeedbackDetailResponse struct {
	ID             string                `json:"id"`
	Summary        string                `json:"summary"`
	OverallScore   float64               `json:"overall_score"`
	Recommendation string                `json:"recommendation"`
	Scores         []FeedbackScoreDetail `json:"scores"`
	Transcript     []TranscriptTurn      `json:"transcript"`
}

// ─── Internal DTOs ────────────────────────────────────────────────────────────

// UpdateInterviewStatusRequest is the request body for the internal
// PATCH /api/v1/interviews/:interview_id/status endpoint.
// NOTE: This is a spec gap resolution — the spec does not define this endpoint
// but requires the orchestrator to update interview status via job-candidate-service.
type UpdateInterviewStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=in_progress completed feedback_ready evaluation_ready"`
}

// ─── Error DTOs ───────────────────────────────────────────────────────────────

// ErrorBody is the inner error object per spec Section 16.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the top-level error envelope per spec Section 16.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ─── JWT Claims ───────────────────────────────────────────────────────────────

// Claims is the JWT payload per spec Section 7.
// sub = user_id, org_id, email, exp (24h).
type Claims struct {
	jwt.RegisteredClaims
	OrgID string `json:"org_id"`
	Email string `json:"email"`
}
