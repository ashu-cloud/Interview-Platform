package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// CandidateRepository handles persistence for candidates.
type CandidateRepository struct {
	db *pgxpool.Pool
}

// NewCandidateRepository constructs a CandidateRepository.
func NewCandidateRepository(db *pgxpool.Pool) *CandidateRepository {
	return &CandidateRepository{db: db}
}

// Create inserts a new candidate scoped to an org and job role.
// Returns ErrConflict if the (org_id, job_role_id, email) combination already exists.
func (r *CandidateRepository) Create(
	ctx context.Context,
	orgID, jobRoleID uuid.UUID,
	req *model.CreateCandidateRequest,
) (*model.Candidate, error) {
	c := &model.Candidate{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO candidates (org_id, job_role_id, email, full_name)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, org_id, job_role_id, email, full_name, created_at`,
		orgID, jobRoleID, req.Email, req.FullName,
	).Scan(&c.ID, &c.OrgID, &c.JobRoleID, &c.Email, &c.FullName, &c.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("CandidateRepository.Create: %w", apperror.ErrConflict)
		}
		return nil, fmt.Errorf("CandidateRepository.Create: %w", err)
	}
	return c, nil
}

// ListByJob returns all candidates for a given job role with their latest interview status.
// Uses LATERAL join for a deterministic result — no reliance on JOIN ordering.
func (r *CandidateRepository) ListByJob(
	ctx context.Context,
	orgID, jobRoleID uuid.UUID,
) ([]model.CandidateWithStatus, error) {
	rows, err := r.db.Query(ctx,
		`SELECT c.id, c.org_id, c.job_role_id, c.email, c.full_name, c.created_at,
		        COALESCE(i.status, '') AS interview_status
		 FROM candidates c
		 LEFT JOIN LATERAL (
		     SELECT status
		     FROM interviews
		     WHERE candidate_id = c.id
		     ORDER BY created_at DESC
		     LIMIT 1
		 ) i ON true
		 WHERE c.org_id = $1 AND c.job_role_id = $2
		 ORDER BY c.created_at DESC`,
		orgID, jobRoleID,
	)
	if err != nil {
		return nil, fmt.Errorf("CandidateRepository.ListByJob: query: %w", err)
	}
	defer rows.Close()

	var candidates []model.CandidateWithStatus
	for rows.Next() {
		var c model.CandidateWithStatus
		if err := rows.Scan(
			&c.ID, &c.OrgID, &c.JobRoleID, &c.Email, &c.FullName,
			&c.CreatedAt, &c.InterviewStatus,
		); err != nil {
			return nil, fmt.Errorf("CandidateRepository.ListByJob: scan: %w", err)
		}
		candidates = append(candidates, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("CandidateRepository.ListByJob: rows: %w", err)
	}
	if candidates == nil {
		candidates = []model.CandidateWithStatus{}
	}
	return candidates, nil
}

// GetByID fetches a single candidate scoped to an org.
// Returns ErrNotFound if not found or belongs to a different org.
// Used for ownership validation before creating an invite.
func (r *CandidateRepository) GetByID(
	ctx context.Context,
	orgID, candidateID uuid.UUID,
) (*model.Candidate, error) {
	c := &model.Candidate{}
	err := r.db.QueryRow(ctx,
		`SELECT id, org_id, job_role_id, email, full_name, created_at
		 FROM candidates WHERE id = $1 AND org_id = $2`,
		candidateID, orgID,
	).Scan(&c.ID, &c.OrgID, &c.JobRoleID, &c.Email, &c.FullName, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("CandidateRepository.GetByID: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("CandidateRepository.GetByID: %w", err)
	}
	return c, nil
}

// GetByIDInternal fetches a single candidate globally (by ID only) for internal service-to-service calls.
func (r *CandidateRepository) GetByIDInternal(ctx context.Context, id uuid.UUID) (*model.Candidate, error) {
	c := &model.Candidate{}
	err := r.db.QueryRow(ctx,
		`SELECT id, org_id, job_role_id, email, full_name, created_at
		 FROM candidates WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.OrgID, &c.JobRoleID, &c.Email, &c.FullName, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("CandidateRepository.GetByIDInternal: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("CandidateRepository.GetByIDInternal: %w", err)
	}
	return c, nil
}
