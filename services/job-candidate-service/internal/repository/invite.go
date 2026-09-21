package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// inviteWithDetails is an internal projection used by the validate endpoint.
// Not exported — the service maps it to InviteValidateResponse.
type inviteWithDetails struct {
	InviteID          uuid.UUID
	InterviewID       uuid.UUID
	InterviewStatus   string
	CandidateID       uuid.UUID
	DocketID          uuid.UUID
	OrgID             uuid.UUID
	Token             string
	ExpiresAt         time.Time
	UsedAt            *time.Time
	CandidateFullName string
	CandidateEmail    string
	JobTitle          string
	DocketTitle       string
}

// InviteRepository handles persistence for invites, interviews, and feedback.
type InviteRepository struct {
	db *pgxpool.Pool
}

// NewInviteRepository constructs an InviteRepository.
func NewInviteRepository(db *pgxpool.Pool) *InviteRepository {
	return &InviteRepository{db: db}
}

// CreateInviteAndInterview creates an invite and a corresponding interview record
// in a single transaction (interview starts with status = 'invited').
func (r *InviteRepository) CreateInviteAndInterview(
	ctx context.Context,
	candidateID, docketID, orgID uuid.UUID,
	token string,
	expiresAt time.Time,
) (*model.InterviewInvite, *model.Interview, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("InviteRepository.CreateInviteAndInterview: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	invite := &model.InterviewInvite{}
	err = tx.QueryRow(ctx,
		`INSERT INTO interview_invites (candidate_id, docket_id, org_id, token, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, candidate_id, docket_id, org_id, token, expires_at, used_at, created_at`,
		candidateID, docketID, orgID, token, expiresAt,
	).Scan(
		&invite.ID, &invite.CandidateID, &invite.DocketID, &invite.OrgID,
		&invite.Token, &invite.ExpiresAt, &invite.UsedAt, &invite.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("InviteRepository.CreateInviteAndInterview: insert invite: %w", err)
	}

	interview := &model.Interview{}
	err = tx.QueryRow(ctx,
		`INSERT INTO interviews (invite_id, candidate_id, docket_id, org_id, status)
		 VALUES ($1, $2, $3, $4, 'invited')
		 RETURNING id, invite_id, candidate_id, docket_id, org_id, status,
		           started_at, completed_at, created_at, updated_at`,
		invite.ID, candidateID, docketID, orgID,
	).Scan(
		&interview.ID, &interview.InviteID, &interview.CandidateID, &interview.DocketID,
		&interview.OrgID, &interview.Status, &interview.StartedAt, &interview.CompletedAt,
		&interview.CreatedAt, &interview.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("InviteRepository.CreateInviteAndInterview: insert interview: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("InviteRepository.CreateInviteAndInterview: commit: %w", err)
	}
	return invite, interview, nil
}

// GetByToken fetches invite details (with candidate/job/docket info) for token validation.
// Returns ErrNotFound if the token doesn't exist.
func (r *InviteRepository) GetByToken(ctx context.Context, token string) (*inviteWithDetails, error) {
	d := &inviteWithDetails{}
	err := r.db.QueryRow(ctx,
		`SELECT ii.id, i.id, i.status, ii.candidate_id, ii.docket_id, ii.org_id, ii.token,
		        ii.expires_at, ii.used_at,
		        c.full_name, c.email,
		        jr.title,
		        d.title
		 FROM interview_invites ii
		 JOIN interviews i ON i.invite_id = ii.id
		 JOIN candidates c ON c.id = ii.candidate_id
		 JOIN dockets d ON d.id = ii.docket_id
		 JOIN job_roles jr ON jr.id = d.job_role_id
		 WHERE ii.token = $1`,
		token,
	).Scan(
		&d.InviteID, &d.InterviewID, &d.InterviewStatus, &d.CandidateID, &d.DocketID, &d.OrgID, &d.Token,
		&d.ExpiresAt, &d.UsedAt,
		&d.CandidateFullName, &d.CandidateEmail,
		&d.JobTitle,
		&d.DocketTitle,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("InviteRepository.GetByToken: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("InviteRepository.GetByToken: %w", err)
	}
	return d, nil
}

// GetByInviteID fetches minimal invite data (id + token) for cache invalidation.
// Returns ErrNotFound if the invite doesn't exist.
func (r *InviteRepository) GetByInviteID(ctx context.Context, inviteID uuid.UUID) (*model.InterviewInvite, error) {
	invite := &model.InterviewInvite{}
	err := r.db.QueryRow(ctx,
		`SELECT id, candidate_id, docket_id, org_id, token, expires_at, used_at, created_at
		 FROM interview_invites WHERE id = $1`,
		inviteID,
	).Scan(
		&invite.ID, &invite.CandidateID, &invite.DocketID, &invite.OrgID,
		&invite.Token, &invite.ExpiresAt, &invite.UsedAt, &invite.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("InviteRepository.GetByInviteID: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("InviteRepository.GetByInviteID: %w", err)
	}
	return invite, nil
}

// GetInviteByInterviewID fetches the invite associated with an interview ID.
func (r *InviteRepository) GetInviteByInterviewID(ctx context.Context, interviewID uuid.UUID) (*model.InterviewInvite, error) {
	invite := &model.InterviewInvite{}
	err := r.db.QueryRow(ctx,
		`SELECT ii.id, ii.candidate_id, ii.docket_id, ii.org_id, ii.token, ii.expires_at, ii.used_at, ii.created_at
		 FROM interview_invites ii
		 JOIN interviews i ON i.invite_id = ii.id
		 WHERE i.id = $1`,
		interviewID,
	).Scan(
		&invite.ID, &invite.CandidateID, &invite.DocketID, &invite.OrgID,
		&invite.Token, &invite.ExpiresAt, &invite.UsedAt, &invite.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("InviteRepository.GetInviteByInterviewID: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("InviteRepository.GetInviteByInterviewID: %w", err)
	}
	return invite, nil
}

// MarkUsed sets used_at on an invite. Cache invalidation is handled by the service layer.
func (r *InviteRepository) MarkUsed(ctx context.Context, inviteID uuid.UUID, at time.Time) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE interview_invites SET used_at = $1 WHERE id = $2`,
		at, inviteID,
	)
	if err != nil {
		return fmt.Errorf("InviteRepository.MarkUsed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("InviteRepository.MarkUsed: %w", apperror.ErrNotFound)
	}
	return nil
}

// UpdateInterviewStatus updates interview status and associated timestamps.
// NOTE: This is a spec gap resolution — required by the orchestrator but not defined in spec Section 7.
func (r *InviteRepository) UpdateInterviewStatus(ctx context.Context, interviewID uuid.UUID, status string) (bool, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE interviews
		 SET status       = $1::varchar,
		     updated_at   = NOW(),
		     started_at   = CASE WHEN $1::varchar = 'in_progress' AND started_at IS NULL THEN NOW() ELSE started_at END,
		     completed_at = CASE WHEN $1::varchar IN ('completed', 'feedback_ready') AND completed_at IS NULL THEN NOW() ELSE completed_at END
		 WHERE id = $2 AND status != $1::varchar`,
		status, interviewID,
	)
	if err != nil {
		return false, fmt.Errorf("InviteRepository.UpdateInterviewStatus: %w", err)
	}
	
	// If RowsAffected is 0, the status was already set (or interview doesn't exist).
	// We return false to indicate no transition occurred.
	return tag.RowsAffected() > 0, nil
}

// GetInterviews returns paginated interviews for an org, with optional status and job filters.
func (r *InviteRepository) GetInterviews(
	ctx context.Context,
	orgID uuid.UUID,
	statusFilter, jobIDFilter string,
	page, limit int,
) ([]model.InterviewDetail, int64, error) {
	offset := (page - 1) * limit

	// Parse optional job_id filter — NULL if empty
	var jobIDArg interface{}
	if jobIDFilter != "" {
		parsed, err := uuid.Parse(jobIDFilter)
		if err != nil {
			return nil, 0, fmt.Errorf("InviteRepository.GetInterviews: invalid job_id: %w", apperror.ErrBadRequest)
		}
		jobIDArg = parsed
	}

	rows, err := r.db.Query(ctx,
		`SELECT i.id, i.status, i.started_at, i.completed_at, i.created_at,
		        c.id, c.full_name, c.email,
		        d.id, d.title,
		        COUNT(*) OVER() AS total
		 FROM interviews i
		 JOIN candidates c ON c.id = i.candidate_id
		 JOIN dockets d ON d.id = i.docket_id
		 WHERE i.org_id = $1
		   AND ($2 = '' OR i.status = $2)
		   AND ($3::uuid IS NULL OR d.job_role_id = $3::uuid)
		 ORDER BY i.created_at DESC
		 LIMIT $4 OFFSET $5`,
		orgID, statusFilter, jobIDArg, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("InviteRepository.GetInterviews: query: %w", err)
	}
	defer rows.Close()

	var interviews []model.InterviewDetail
	var total int64
	for rows.Next() {
		var d model.InterviewDetail
		if err := rows.Scan(
			&d.ID, &d.Status, &d.StartedAt, &d.CompletedAt, nil,
			&d.Candidate.ID, &d.Candidate.FullName, &d.Candidate.Email,
			&d.Docket.ID, &d.Docket.Title,
			&total,
		); err != nil {
			return nil, 0, fmt.Errorf("InviteRepository.GetInterviews: scan: %w", err)
		}
		interviews = append(interviews, d)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("InviteRepository.GetInterviews: rows: %w", err)
	}
	if interviews == nil {
		interviews = []model.InterviewDetail{}
	}
	return interviews, total, nil
}

// GetInterviewByID returns a single interview detail scoped to an org.
// Returns ErrNotFound if not found or belongs to a different org.
func (r *InviteRepository) GetInterviewByID(
	ctx context.Context,
	orgID, interviewID uuid.UUID,
) (*model.InterviewDetail, error) {
	d := &model.InterviewDetail{}
	err := r.db.QueryRow(ctx,
		`SELECT i.id, i.status, i.started_at, i.completed_at,
		        c.id, c.full_name, c.email,
		        d.id, d.title
		 FROM interviews i
		 JOIN candidates c ON c.id = i.candidate_id
		 JOIN dockets d ON d.id = i.docket_id
		 WHERE i.id = $1 AND i.org_id = $2`,
		interviewID, orgID,
	).Scan(
		&d.ID, &d.Status, &d.StartedAt, &d.CompletedAt,
		&d.Candidate.ID, &d.Candidate.FullName, &d.Candidate.Email,
		&d.Docket.ID, &d.Docket.Title,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("InviteRepository.GetInterviewByID: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("InviteRepository.GetInterviewByID: %w", err)
	}
	return d, nil
}

// GetFeedback returns the structured feedback report for an interview.
// Enforces org ownership via JOIN on interviews.
// Returns ErrNotFound if the interview/feedback doesn't exist or belongs to another org.
func (r *InviteRepository) GetFeedback(
	ctx context.Context,
	orgID, interviewID uuid.UUID,
) (*model.FeedbackDetailResponse, error) {
	// Fetch feedback report (org is checked via the interviews join)
	resp := &model.FeedbackDetailResponse{}
	var transcriptRaw []byte

	err := r.db.QueryRow(ctx,
		`SELECT fr.id, fr.summary, fr.overall_score, fr.recommendation, fr.transcript_json
		 FROM feedback_reports fr
		 JOIN interviews i ON i.id = fr.interview_id
		 WHERE fr.interview_id = $1 AND i.org_id = $2`,
		interviewID, orgID,
	).Scan(&resp.ID, &resp.Summary, &resp.OverallScore, &resp.Recommendation, &transcriptRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("InviteRepository.GetFeedback: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("InviteRepository.GetFeedback: query report: %w", err)
	}

	// Decode transcript JSONB
	if transcriptRaw != nil {
		if err := json.Unmarshal(transcriptRaw, &resp.Transcript); err != nil {
			return nil, fmt.Errorf("InviteRepository.GetFeedback: decode transcript: %w", err)
		}
	}
	if resp.Transcript == nil {
		resp.Transcript = []model.TranscriptTurn{}
	}

	// Fetch per-criterion scores with rubric names
	sRows, err := r.db.Query(ctx,
		`SELECT fs.score, fs.evidence, rc.name, rc.max_score
		 FROM feedback_scores fs
		 JOIN feedback_reports fr ON fr.id = fs.feedback_id
		 JOIN rubric_criteria rc ON rc.id = fs.criterion_id
		 WHERE fr.interview_id = $1
		 ORDER BY rc.order_index`,
		interviewID,
	)
	if err != nil {
		return nil, fmt.Errorf("InviteRepository.GetFeedback: query scores: %w", err)
	}
	defer sRows.Close()

	resp.Scores = []model.FeedbackScoreDetail{}
	for sRows.Next() {
		var s model.FeedbackScoreDetail
		if err := sRows.Scan(&s.Score, &s.Evidence, &s.CriterionName, &s.MaxScore); err != nil {
			return nil, fmt.Errorf("InviteRepository.GetFeedback: scan score: %w", err)
		}
		resp.Scores = append(resp.Scores, s)
	}
	if err := sRows.Err(); err != nil {
		return nil, fmt.Errorf("InviteRepository.GetFeedback: scores rows: %w", err)
	}

	return resp, nil
}
