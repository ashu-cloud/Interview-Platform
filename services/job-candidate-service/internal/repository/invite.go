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
		`INSERT INTO assessment_invites (candidate_id, docket_id, org_id, token, expires_at)
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

	interview, err := createSession(ctx, tx, invite.ID, candidateID, docketID, orgID)
	if err != nil {
		return nil, nil, fmt.Errorf("InviteRepository.CreateInviteAndInterview: insert interview: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("InviteRepository.CreateInviteAndInterview: commit: %w", err)
	}
	return invite, interview, nil
}

// CreateSession validates an org-owned invite and idempotently creates or reuses its session.
func (r *InviteRepository) CreateSession(
	ctx context.Context,
	orgID, inviteID uuid.UUID,
) (*model.Interview, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("InviteRepository.CreateSession: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	invite := &model.InterviewInvite{}
	err = tx.QueryRow(ctx,
		`SELECT id, candidate_id, docket_id, org_id, token, expires_at, used_at, created_at
		 FROM assessment_invites
		 WHERE id = $1 AND org_id = $2
		 FOR UPDATE`,
		inviteID, orgID,
	).Scan(
		&invite.ID, &invite.CandidateID, &invite.DocketID, &invite.OrgID,
		&invite.Token, &invite.ExpiresAt, &invite.UsedAt, &invite.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("InviteRepository.CreateSession: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("InviteRepository.CreateSession: get invite: %w", err)
	}

	interview := &model.Interview{}
	err = tx.QueryRow(ctx,
		`SELECT id, invite_id, candidate_id, docket_id, org_id, status,
		        started_at, completed_at, created_at, updated_at
		 FROM assessment_sessions
		 WHERE invite_id = $1 AND org_id = $2
		 ORDER BY created_at
		 LIMIT 1`,
		inviteID, orgID,
	).Scan(
		&interview.ID, &interview.InviteID, &interview.CandidateID, &interview.DocketID,
		&interview.OrgID, &interview.Status, &interview.StartedAt, &interview.CompletedAt,
		&interview.CreatedAt, &interview.UpdatedAt,
	)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("InviteRepository.CreateSession: commit reuse: %w", err)
		}
		return interview, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("InviteRepository.CreateSession: find session: %w", err)
	}
	if invite.UsedAt != nil {
		return nil, fmt.Errorf("InviteRepository.CreateSession: invite already used: %w", apperror.ErrConflict)
	}
	if !invite.ExpiresAt.After(time.Now().UTC()) {
		return nil, fmt.Errorf("InviteRepository.CreateSession: invite expired: %w", apperror.ErrNotFound)
	}

	interview, err = createSession(ctx, tx, invite.ID, invite.CandidateID, invite.DocketID, invite.OrgID)
	if err != nil {
		return nil, fmt.Errorf("InviteRepository.CreateSession: insert session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("InviteRepository.CreateSession: commit: %w", err)
	}
	return interview, nil
}

type sessionRowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func createSession(
	ctx context.Context,
	q sessionRowQuerier,
	inviteID, candidateID, docketID, orgID uuid.UUID,
) (*model.Interview, error) {
	interview := &model.Interview{}
	err := q.QueryRow(ctx,
		`INSERT INTO assessment_sessions (invite_id, candidate_id, docket_id, org_id, status)
		 VALUES ($1, $2, $3, $4, 'invited')
		 RETURNING id, invite_id, candidate_id, docket_id, org_id, status,
		           started_at, completed_at, created_at, updated_at`,
		inviteID, candidateID, docketID, orgID,
	).Scan(
		&interview.ID, &interview.InviteID, &interview.CandidateID, &interview.DocketID,
		&interview.OrgID, &interview.Status, &interview.StartedAt, &interview.CompletedAt,
		&interview.CreatedAt, &interview.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return interview, nil
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
		 FROM assessment_invites ii
		 JOIN assessment_sessions i ON i.invite_id = ii.id
		 JOIN participants c ON c.id = ii.candidate_id
		 JOIN assessment_blueprints d ON d.id = ii.docket_id
		 JOIN assessment_tracks jr ON jr.id = d.job_role_id
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
		 FROM assessment_invites WHERE id = $1`,
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
		 FROM assessment_invites ii
		 JOIN assessment_sessions i ON i.invite_id = ii.id
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
		`UPDATE assessment_invites SET used_at = $1 WHERE id = $2`,
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
		`UPDATE assessment_sessions
		 SET status       = $1::varchar,
		     updated_at   = NOW(),
		     started_at   = CASE WHEN $1::varchar = 'in_progress' AND started_at IS NULL THEN NOW() ELSE started_at END,
		     completed_at = CASE WHEN $1::varchar IN ('completed', 'feedback_ready', 'evaluation_ready') AND completed_at IS NULL THEN NOW() ELSE completed_at END
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
		        er.id::text, er.overall_score, er.recommendation,
		        COUNT(*) OVER() AS total
		 FROM assessment_sessions i
		 JOIN participants c ON c.id = i.candidate_id
		 JOIN assessment_blueprints d ON d.id = i.docket_id
		 LEFT JOIN evaluation_reports er ON er.interview_id = i.id
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
		var fbID *string
		var fbScore *float64
		var fbRec *string
		if err := rows.Scan(
			&d.ID, &d.Status, &d.StartedAt, &d.CompletedAt, nil,
			&d.Candidate.ID, &d.Candidate.FullName, &d.Candidate.Email,
			&d.Docket.ID, &d.Docket.Title,
			&fbID, &fbScore, &fbRec,
			&total,
		); err != nil {
			return nil, 0, fmt.Errorf("InviteRepository.GetInterviews: scan: %w", err)
		}
		if fbID != nil && fbScore != nil && fbRec != nil {
			d.Feedback = &model.FeedbackBrief{
				ID:             *fbID,
				OverallScore:   *fbScore,
				Recommendation: *fbRec,
			}
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
		 FROM assessment_sessions i
		 JOIN participants c ON c.id = i.candidate_id
		 JOIN assessment_blueprints d ON d.id = i.docket_id
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

	if d.Status == "feedback_ready" || d.Status == "evaluation_ready" || d.Status == "completed" {
		var fb model.FeedbackBrief
		fbErr := r.db.QueryRow(ctx,
			`SELECT id::text, overall_score, recommendation FROM evaluation_reports WHERE interview_id = $1`,
			interviewID,
		).Scan(&fb.ID, &fb.OverallScore, &fb.Recommendation)
		if fbErr == nil {
			d.Feedback = &fb
		}
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
		 FROM evaluation_reports fr
		 JOIN assessment_sessions i ON i.id = fr.interview_id
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
		`SELECT fs.score, fs.evidence, rc.name, rc.max_score, fs.evidence_turn_id::text, COALESCE(fs.is_grounded, true)
		 FROM evaluation_scores fs
		 JOIN evaluation_reports fr ON fr.id = fs.feedback_id
		 JOIN evaluation_criteria rc ON rc.id = fs.criterion_id
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
		var turnID *string
		if err := sRows.Scan(&s.Score, &s.Evidence, &s.CriterionName, &s.MaxScore, &turnID, &s.IsGrounded); err != nil {
			return nil, fmt.Errorf("InviteRepository.GetFeedback: scan score: %w", err)
		}
		if turnID != nil && *turnID != "" {
			s.EvidenceTurnID = turnID
		}
		resp.Scores = append(resp.Scores, s)
	}
	if err := sRows.Err(); err != nil {
		return nil, fmt.Errorf("InviteRepository.GetFeedback: scores rows: %w", err)
	}

	return resp, nil
}

// ListSessionTurns returns materialized turns for a session owned by an org.
func (r *InviteRepository) ListSessionTurns(
	ctx context.Context,
	orgID, sessionID uuid.UUID,
) ([]model.SessionTurn, error) {
	rows, err := r.db.Query(ctx,
		`SELECT st.id, st.session_id, st.org_id, st.turn_number, st.role, st.content, st.created_at
		 FROM session_turns st
		 JOIN assessment_sessions s ON s.id = st.session_id
		 WHERE st.session_id = $1 AND s.org_id = $2
		 ORDER BY st.turn_number, st.created_at, st.id`,
		sessionID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("InviteRepository.ListSessionTurns: query: %w", err)
	}
	defer rows.Close()

	turns := []model.SessionTurn{}
	for rows.Next() {
		var turn model.SessionTurn
		if err := rows.Scan(
			&turn.ID, &turn.SessionID, &turn.OrgID, &turn.TurnNumber,
			&turn.Role, &turn.Content, &turn.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("InviteRepository.ListSessionTurns: scan: %w", err)
		}
		turns = append(turns, turn)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("InviteRepository.ListSessionTurns: rows: %w", err)
	}
	if len(turns) == 0 {
		var exists bool
		if err := r.db.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM assessment_sessions WHERE id = $1 AND org_id = $2)`,
			sessionID, orgID,
		).Scan(&exists); err != nil {
			return nil, fmt.Errorf("InviteRepository.ListSessionTurns: verify session: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("InviteRepository.ListSessionTurns: %w", apperror.ErrNotFound)
		}
	}
	return turns, nil
}

// EnqueueEvaluation idempotently creates an outbox_completed event for a completed session.
func (r *InviteRepository) EnqueueEvaluation(
	ctx context.Context,
	orgID, sessionID uuid.UUID,
) (*model.EvaluationJob, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("InviteRepository.EnqueueEvaluation: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var participantID, blueprintID uuid.UUID
	var status string
	err = tx.QueryRow(ctx,
		`SELECT candidate_id, docket_id, status
		 FROM assessment_sessions
		 WHERE id = $1 AND org_id = $2
		 FOR UPDATE`,
		sessionID, orgID,
	).Scan(&participantID, &blueprintID, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("InviteRepository.EnqueueEvaluation: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("InviteRepository.EnqueueEvaluation: lock session: %w", err)
	}
	if status != "completed" && status != "feedback_ready" && status != "evaluation_ready" {
		return nil, fmt.Errorf("InviteRepository.EnqueueEvaluation: session is not completed: %w", apperror.ErrConflict)
	}

	job := &model.EvaluationJob{}
	err = tx.QueryRow(ctx,
		`SELECT id, session_id, org_id, participant_id, blueprint_id, turn_count, status, created_at
		 FROM outbox_completed
		 WHERE session_id = $1 AND org_id = $2
		 ORDER BY created_at, id
		 LIMIT 1`,
		sessionID, orgID,
	).Scan(
		&job.ID, &job.SessionID, &job.OrgID, &job.ParticipantID, &job.BlueprintID,
		&job.TurnCount, &job.Status, &job.CreatedAt,
	)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("InviteRepository.EnqueueEvaluation: commit existing: %w", err)
		}
		return job, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("InviteRepository.EnqueueEvaluation: find existing job: %w", err)
	}

	var turnCount int
	if err := tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(turn_number), 0) FROM session_turns WHERE session_id = $1`,
		sessionID,
	).Scan(&turnCount); err != nil {
		return nil, fmt.Errorf("InviteRepository.EnqueueEvaluation: get turn count: %w", err)
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO outbox_completed (session_id, org_id, participant_id, blueprint_id, turn_count)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, session_id, org_id, participant_id, blueprint_id, turn_count, status, created_at`,
		sessionID, orgID, participantID, blueprintID, turnCount,
	).Scan(
		&job.ID, &job.SessionID, &job.OrgID, &job.ParticipantID, &job.BlueprintID,
		&job.TurnCount, &job.Status, &job.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("InviteRepository.EnqueueEvaluation: insert job: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("InviteRepository.EnqueueEvaluation: commit: %w", err)
	}
	return job, nil
}

// GetPublicTrack retrieves track details and the first blueprint for unauthenticated applicants.
func (r *InviteRepository) GetPublicTrack(ctx context.Context, jobID uuid.UUID) (*model.PublicTrackResponse, error) {
	resp := &model.PublicTrackResponse{}
	err := r.db.QueryRow(ctx,
		`SELECT id, org_id, created_by, title, description,
		        COALESCE(required_skills, ARRAY[]::text[]),
		        status, created_at, updated_at
		 FROM assessment_tracks
		 WHERE id = $1`,
		jobID,
	).Scan(
		&resp.Track.ID, &resp.Track.OrgID, &resp.Track.CreatedBy, &resp.Track.Title, &resp.Track.Description,
		&resp.Track.RequiredSkills, &resp.Track.Status, &resp.Track.CreatedAt, &resp.Track.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("InviteRepository.GetPublicTrack: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("InviteRepository.GetPublicTrack: query track: %w", err)
	}

	// Fetch active blueprint if available
	var bpID uuid.UUID
	var bpTitle string
	var bpDuration int
	bpErr := r.db.QueryRow(ctx,
		`SELECT id, title, duration_mins
		 FROM assessment_blueprints
		 WHERE job_role_id = $1
		 ORDER BY created_at ASC
		 LIMIT 1`,
		jobID,
	).Scan(&bpID, &bpTitle, &bpDuration)
	if bpErr == nil {
		resp.BlueprintID = bpID.String()
		resp.BlueprintTitle = bpTitle
		resp.DurationMins = bpDuration
	}

	return resp, nil
}

// PublicApply finds or creates a participant for the track and generates an invite + interview.
func (r *InviteRepository) PublicApply(
	ctx context.Context,
	jobID uuid.UUID,
	req *model.PublicApplyRequest,
	token string,
	expiresAt time.Time,
) (*model.InterviewInvite, *model.Interview, uuid.UUID, uuid.UUID, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Verify track
	var orgID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT org_id FROM assessment_tracks WHERE id = $1`, jobID).Scan(&orgID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: track %w", apperror.ErrNotFound)
		}
		return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: query track: %w", err)
	}

	// 2. Find blueprint
	var blueprintID uuid.UUID
	if err := tx.QueryRow(ctx,
		`SELECT id FROM assessment_blueprints WHERE job_role_id = $1 ORDER BY created_at ASC LIMIT 1`,
		jobID,
	).Scan(&blueprintID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: track has no active blueprint: %w", apperror.ErrConflict)
		}
		return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: query blueprint: %w", err)
	}

	// 3. Find or create participant
	var participantID uuid.UUID
	err = tx.QueryRow(ctx,
		`SELECT id FROM participants WHERE job_role_id = $1 AND email = $2`,
		jobID, req.Email,
	).Scan(&participantID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx,
			`INSERT INTO participants (org_id, job_role_id, full_name, email)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id`,
			orgID, jobID, req.FullName, req.Email,
		).Scan(&participantID)
		if err != nil {
			return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: insert participant: %w", err)
		}
	} else if err != nil {
		return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: find participant: %w", err)
	}

	// 4. Create invite
	invite := &model.InterviewInvite{}
	err = tx.QueryRow(ctx,
		`INSERT INTO participant_invites (candidate_id, docket_id, org_id, token, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, candidate_id, docket_id, org_id, token, expires_at, used_at, created_at`,
		participantID, blueprintID, orgID, token, expiresAt,
	).Scan(
		&invite.ID, &invite.CandidateID, &invite.DocketID, &invite.OrgID,
		&invite.Token, &invite.ExpiresAt, &invite.UsedAt, &invite.CreatedAt,
	)
	if err != nil {
		return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: insert invite: %w", err)
	}

	// 5. Create interview
	interview := &model.Interview{}
	err = tx.QueryRow(ctx,
		`INSERT INTO interview_sessions (invite_id, candidate_id, docket_id, org_id, status)
		 VALUES ($1, $2, $3, $4, 'invited')
		 RETURNING id, invite_id, candidate_id, docket_id, org_id, status, created_at, updated_at`,
		invite.ID, participantID, blueprintID, orgID,
	).Scan(
		&interview.ID, &interview.InviteID, &interview.CandidateID, &interview.DocketID,
		&interview.OrgID, &interview.Status, &interview.CreatedAt, &interview.UpdatedAt,
	)
	if err != nil {
		return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: insert interview: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, uuid.Nil, uuid.Nil, fmt.Errorf("InviteRepository.PublicApply: commit: %w", err)
	}

	return invite, interview, orgID, participantID, nil
}
