package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// DocketRepository handles persistence for dockets, questions, and rubric criteria.
type DocketRepository struct {
	db *pgxpool.Pool
}

// NewDocketRepository constructs a DocketRepository.
func NewDocketRepository(db *pgxpool.Pool) *DocketRepository {
	return &DocketRepository{db: db}
}

// Create inserts a docket with all questions and rubric criteria in a single transaction.
func (r *DocketRepository) Create(
	ctx context.Context,
	orgID, jobRoleID uuid.UUID,
	req *model.CreateDocketRequest,
) (*model.Docket, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("DocketRepository.Create: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Insert docket
	docket := &model.Docket{}
	err = tx.QueryRow(ctx,
		`INSERT INTO dockets (job_role_id, org_id, title, instructions, duration_mins)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, job_role_id, org_id, title, instructions, duration_mins, created_at, updated_at`,
		jobRoleID, orgID, req.Title, req.Instructions, req.DurationMins,
	).Scan(
		&docket.ID, &docket.JobRoleID, &docket.OrgID, &docket.Title,
		&docket.Instructions, &docket.DurationMins, &docket.CreatedAt, &docket.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("DocketRepository.Create: insert docket: %w", err)
	}

	// Insert questions
	docket.Questions = make([]model.DocketQuestion, 0, len(req.Questions))
	for _, q := range req.Questions {
		question := model.DocketQuestion{}
		err = tx.QueryRow(ctx,
			`INSERT INTO docket_questions (docket_id, question_text, question_type, order_index, follow_up_hint)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, docket_id, question_text, question_type, order_index, follow_up_hint, created_at`,
			docket.ID, q.QuestionText, q.QuestionType, q.OrderIndex, q.FollowUpHint,
		).Scan(
			&question.ID, &question.DocketID, &question.QuestionText, &question.QuestionType,
			&question.OrderIndex, &question.FollowUpHint, &question.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("DocketRepository.Create: insert question: %w", err)
		}
		docket.Questions = append(docket.Questions, question)
	}

	// Insert rubric criteria
	docket.Rubric = make([]model.RubricCriteria, 0, len(req.Rubric))
	for _, rb := range req.Rubric {
		criteria := model.RubricCriteria{}
		err = tx.QueryRow(ctx,
			`INSERT INTO rubric_criteria (docket_id, name, description, max_score, order_index)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, docket_id, name, description, max_score, order_index, created_at`,
			docket.ID, rb.Name, rb.Description, rb.MaxScore, rb.OrderIndex,
		).Scan(
			&criteria.ID, &criteria.DocketID, &criteria.Name, &criteria.Description,
			&criteria.MaxScore, &criteria.OrderIndex, &criteria.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("DocketRepository.Create: insert rubric: %w", err)
		}
		docket.Rubric = append(docket.Rubric, criteria)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("DocketRepository.Create: commit: %w", err)
	}

	return docket, nil
}

// GetByID fetches a full docket (with questions and rubric) scoped to an org.
// Returns ErrNotFound if not found or belongs to a different org.
func (r *DocketRepository) GetByID(ctx context.Context, orgID, docketID uuid.UUID) (*model.Docket, error) {
	docket, err := r.fetchDocket(ctx,
		`SELECT id, job_role_id, org_id, title, instructions, duration_mins, created_at, updated_at
		 FROM dockets WHERE id = $1 AND org_id = $2`,
		docketID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("DocketRepository.GetByID: %w", err)
	}
	return docket, nil
}

// GetByIDInternal fetches a full docket without org scoping (internal use by orchestrator).
// Returns ErrNotFound if the docket does not exist.
func (r *DocketRepository) GetByIDInternal(ctx context.Context, docketID uuid.UUID) (*model.Docket, error) {
	docket, err := r.fetchDocket(ctx,
		`SELECT id, job_role_id, org_id, title, instructions, duration_mins, created_at, updated_at
		 FROM dockets WHERE id = $1`,
		docketID,
	)
	if err != nil {
		return nil, fmt.Errorf("DocketRepository.GetByIDInternal: %w", err)
	}
	return docket, nil
}

// fetchDocket is a shared helper that runs the docket query and fetches nested data.
func (r *DocketRepository) fetchDocket(ctx context.Context, query string, args ...interface{}) (*model.Docket, error) {
	docket := &model.Docket{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&docket.ID, &docket.JobRoleID, &docket.OrgID, &docket.Title,
		&docket.Instructions, &docket.DurationMins, &docket.CreatedAt, &docket.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	// Fetch questions ordered by order_index
	qRows, err := r.db.Query(ctx,
		`SELECT id, docket_id, question_text, question_type, order_index, follow_up_hint, created_at
		 FROM docket_questions WHERE docket_id = $1 ORDER BY order_index`,
		docket.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch questions: %w", err)
	}
	defer qRows.Close()

	docket.Questions = []model.DocketQuestion{}
	for qRows.Next() {
		var q model.DocketQuestion
		if err := qRows.Scan(
			&q.ID, &q.DocketID, &q.QuestionText, &q.QuestionType,
			&q.OrderIndex, &q.FollowUpHint, &q.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan question: %w", err)
		}
		docket.Questions = append(docket.Questions, q)
	}
	if err := qRows.Err(); err != nil {
		return nil, fmt.Errorf("questions rows: %w", err)
	}

	// Fetch rubric criteria ordered by order_index
	rRows, err := r.db.Query(ctx,
		`SELECT id, docket_id, name, description, max_score, order_index, created_at
		 FROM rubric_criteria WHERE docket_id = $1 ORDER BY order_index`,
		docket.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch rubric: %w", err)
	}
	defer rRows.Close()

	docket.Rubric = []model.RubricCriteria{}
	for rRows.Next() {
		var rc model.RubricCriteria
		if err := rRows.Scan(
			&rc.ID, &rc.DocketID, &rc.Name, &rc.Description,
			&rc.MaxScore, &rc.OrderIndex, &rc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan rubric: %w", err)
		}
		docket.Rubric = append(docket.Rubric, rc)
	}
	if err := rRows.Err(); err != nil {
		return nil, fmt.Errorf("rubric rows: %w", err)
	}

	return docket, nil
}
