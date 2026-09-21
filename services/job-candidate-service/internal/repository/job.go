package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// JobRepository handles persistence for job roles.
type JobRepository struct {
	db *pgxpool.Pool
}

// NewJobRepository constructs a JobRepository.
func NewJobRepository(db *pgxpool.Pool) *JobRepository {
	return &JobRepository{db: db}
}

// Create inserts a new job role and returns the created record.
func (r *JobRepository) Create(
	ctx context.Context,
	orgID, userID uuid.UUID,
	req *model.CreateJobRequest,
) (*model.JobRole, error) {
	skills := req.RequiredSkills
	if skills == nil {
		skills = []string{}
	}
	job := &model.JobRole{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO job_roles (org_id, created_by, title, description, required_skills)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, org_id, created_by, title, description,
		           COALESCE(required_skills, ARRAY[]::text[]),
		           status, created_at, updated_at`,
		orgID, userID, req.Title, req.Description, skills,
	).Scan(
		&job.ID, &job.OrgID, &job.CreatedBy, &job.Title, &job.Description,
		&job.RequiredSkills, &job.Status, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("JobRepository.Create: %w", err)
	}
	return job, nil
}

// List returns paginated job roles for an org, optionally filtered by status.
// total is the count across all pages.
func (r *JobRepository) List(
	ctx context.Context,
	orgID uuid.UUID,
	status string,
	page, limit int,
) ([]model.JobRole, int64, error) {
	offset := (page - 1) * limit

	rows, err := r.db.Query(ctx,
		`SELECT id, org_id, created_by, title, description,
		        COALESCE(required_skills, ARRAY[]::text[]),
		        status, created_at, updated_at,
		        COUNT(*) OVER() AS total
		 FROM job_roles
		 WHERE org_id = $1
		   AND ($2 = '' OR status = $2)
		 ORDER BY created_at DESC
		 LIMIT $3 OFFSET $4`,
		orgID, status, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("JobRepository.List: query: %w", err)
	}
	defer rows.Close()

	var jobs []model.JobRole
	var total int64
	for rows.Next() {
		var j model.JobRole
		if err := rows.Scan(
			&j.ID, &j.OrgID, &j.CreatedBy, &j.Title, &j.Description,
			&j.RequiredSkills, &j.Status, &j.CreatedAt, &j.UpdatedAt,
			&total,
		); err != nil {
			return nil, 0, fmt.Errorf("JobRepository.List: scan: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("JobRepository.List: rows: %w", err)
	}
	if jobs == nil {
		jobs = []model.JobRole{}
	}
	return jobs, total, nil
}

// GetByID fetches a single job role scoped to an org.
// Returns ErrNotFound if not found or belongs to a different org.
func (r *JobRepository) GetByID(
	ctx context.Context,
	orgID, jobID uuid.UUID,
) (*model.JobRole, error) {
	job := &model.JobRole{}
	err := r.db.QueryRow(ctx,
		`SELECT id, org_id, created_by, title, description,
		        COALESCE(required_skills, ARRAY[]::text[]),
		        status, created_at, updated_at
		 FROM job_roles
		 WHERE id = $1 AND org_id = $2`,
		jobID, orgID,
	).Scan(
		&job.ID, &job.OrgID, &job.CreatedBy, &job.Title, &job.Description,
		&job.RequiredSkills, &job.Status, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("JobRepository.GetByID: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("JobRepository.GetByID: %w", err)
	}
	return job, nil
}

// Patch applies a partial update to a job role.
// Only non-nil fields in PatchJobRequest are updated.
// Returns ErrNotFound if the job doesn't belong to the org.
func (r *JobRepository) Patch(
	ctx context.Context,
	orgID, jobID uuid.UUID,
	req *model.PatchJobRequest,
) (*model.JobRole, error) {
	setClauses := []string{}
	args := []interface{}{}
	argN := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argN))
		args = append(args, *req.Title)
		argN++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argN))
		args = append(args, *req.Description)
		argN++
	}
	if req.RequiredSkills != nil {
		setClauses = append(setClauses, fmt.Sprintf("required_skills = $%d", argN))
		args = append(args, req.RequiredSkills)
		argN++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argN))
		args = append(args, *req.Status)
		argN++
	}

	if len(setClauses) == 0 {
		// Nothing to update — return current record
		return r.GetByID(ctx, orgID, jobID)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, orgID, jobID)

	query := fmt.Sprintf(
		`UPDATE job_roles SET %s
		 WHERE org_id = $%d AND id = $%d
		 RETURNING id, org_id, created_by, title, description,
		           COALESCE(required_skills, ARRAY[]::text[]),
		           status, created_at, updated_at`,
		strings.Join(setClauses, ", "), argN, argN+1,
	)

	job := &model.JobRole{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&job.ID, &job.OrgID, &job.CreatedBy, &job.Title, &job.Description,
		&job.RequiredSkills, &job.Status, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("JobRepository.Patch: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("JobRepository.Patch: %w", err)
	}
	return job, nil
}
