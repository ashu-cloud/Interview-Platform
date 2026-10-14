package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
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
		`INSERT INTO assessment_blueprints (job_role_id, org_id, title, instructions, duration_mins)
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
			`INSERT INTO blueprint_prompts (docket_id, question_text, question_type, order_index, follow_up_hint)
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
			`INSERT INTO evaluation_criteria (docket_id, name, description, max_score, order_index)
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

// ListByJob returns blueprints for a track scoped to an org.
func (r *DocketRepository) ListByJob(ctx context.Context, orgID, jobID uuid.UUID) ([]model.Docket, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, job_role_id, org_id, title, instructions, duration_mins, created_at, updated_at
		 FROM assessment_blueprints
		 WHERE org_id = $1 AND job_role_id = $2
		 ORDER BY created_at DESC`,
		orgID, jobID,
	)
	if err != nil {
		return nil, fmt.Errorf("DocketRepository.ListByJob: query: %w", err)
	}
	defer rows.Close()

	dockets := []model.Docket{}
	for rows.Next() {
		var docket model.Docket
		if err := rows.Scan(
			&docket.ID, &docket.JobRoleID, &docket.OrgID, &docket.Title,
			&docket.Instructions, &docket.DurationMins, &docket.CreatedAt, &docket.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("DocketRepository.ListByJob: scan: %w", err)
		}
		dockets = append(dockets, docket)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DocketRepository.ListByJob: rows: %w", err)
	}
	return dockets, nil
}

// Update applies a partial update and replaces nested collections when supplied.
func (r *DocketRepository) Update(
	ctx context.Context,
	orgID, docketID uuid.UUID,
	req *model.UpdateDocketRequest,
) (*model.Docket, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("DocketRepository.Update: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`UPDATE assessment_blueprints
		 SET title = COALESCE($1, title),
		     instructions = COALESCE($2, instructions),
		     duration_mins = COALESCE($3, duration_mins),
		     updated_at = NOW()
		 WHERE id = $4 AND org_id = $5`,
		req.Title, req.Instructions, req.DurationMins, docketID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("DocketRepository.Update: update blueprint: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("DocketRepository.Update: %w", apperror.ErrNotFound)
	}

	if req.Questions != nil || req.Rubric != nil {
		var hasSessions bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM assessment_sessions WHERE docket_id = $1)`,
			docketID,
		).Scan(&hasSessions); err != nil {
			return nil, fmt.Errorf("DocketRepository.Update: check sessions: %w", err)
		}
		if hasSessions {
			return nil, fmt.Errorf("DocketRepository.Update: cannot replace prompts or criteria after sessions exist: %w", apperror.ErrConflict)
		}
	}

	if req.Questions != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM blueprint_prompts WHERE docket_id = $1`, docketID); err != nil {
			return nil, fmt.Errorf("DocketRepository.Update: delete prompts: %w", err)
		}
		for _, q := range *req.Questions {
			if _, err := tx.Exec(ctx,
				`INSERT INTO blueprint_prompts (docket_id, question_text, question_type, order_index, follow_up_hint)
				 VALUES ($1, $2, $3, $4, $5)`,
				docketID, q.QuestionText, q.QuestionType, q.OrderIndex, q.FollowUpHint,
			); err != nil {
				return nil, fmt.Errorf("DocketRepository.Update: insert prompt: %w", err)
			}
		}
	}

	if req.Rubric != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM evaluation_criteria WHERE docket_id = $1`, docketID); err != nil {
			return nil, fmt.Errorf("DocketRepository.Update: delete criteria: %w", err)
		}
		for _, criterion := range *req.Rubric {
			if _, err := tx.Exec(ctx,
				`INSERT INTO evaluation_criteria (docket_id, name, description, max_score, order_index)
				 VALUES ($1, $2, $3, $4, $5)`,
				docketID, criterion.Name, criterion.Description, criterion.MaxScore, criterion.OrderIndex,
			); err != nil {
				return nil, fmt.Errorf("DocketRepository.Update: insert criterion: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("DocketRepository.Update: commit: %w", err)
	}
	return r.GetByID(ctx, orgID, docketID)
}

// Delete removes a blueprint scoped to an org.
func (r *DocketRepository) Delete(ctx context.Context, orgID, docketID uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM assessment_blueprints d
		 WHERE d.id = $1 AND d.org_id = $2
		   AND NOT EXISTS (SELECT 1 FROM assessment_invites i WHERE i.docket_id = d.id)
		   AND NOT EXISTS (SELECT 1 FROM assessment_sessions s WHERE s.docket_id = d.id)`,
		docketID, orgID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return fmt.Errorf("DocketRepository.Delete: blueprint has dependent records: %w", apperror.ErrConflict)
		}
		return fmt.Errorf("DocketRepository.Delete: %w", err)
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	var exists bool
	if err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM assessment_blueprints WHERE id = $1 AND org_id = $2)`,
		docketID, orgID,
	).Scan(&exists); err != nil {
		return fmt.Errorf("DocketRepository.Delete: verify blueprint: %w", err)
	}
	if exists {
		return fmt.Errorf("DocketRepository.Delete: blueprint has dependent records: %w", apperror.ErrConflict)
	}
	return fmt.Errorf("DocketRepository.Delete: %w", apperror.ErrNotFound)
}

// GetByID fetches a full docket (with questions and rubric) scoped to an org.
// Returns ErrNotFound if not found or belongs to a different org.
func (r *DocketRepository) GetByID(ctx context.Context, orgID, docketID uuid.UUID) (*model.Docket, error) {
	docket, err := r.fetchDocket(ctx,
		`SELECT id, job_role_id, org_id, title, instructions, duration_mins, created_at, updated_at
		 FROM assessment_blueprints WHERE id = $1 AND org_id = $2`,
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
		 FROM assessment_blueprints WHERE id = $1`,
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
		 FROM blueprint_prompts WHERE docket_id = $1 ORDER BY order_index`,
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
		 FROM evaluation_criteria WHERE docket_id = $1 ORDER BY order_index`,
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

// UpdateCriterionEmbedding stores the unit-vector embedding for a criterion.
func (r *DocketRepository) UpdateCriterionEmbedding(ctx context.Context, criterionID uuid.UUID, vec []float32) error {
	_, err := r.db.Exec(ctx,
		`UPDATE evaluation_criteria SET embedding = $1 WHERE id = $2`,
		pgvector.NewVector(vec), criterionID)
	if err != nil {
		return fmt.Errorf("DocketRepository.UpdateCriterionEmbedding: %w", err)
	}
	return nil
}

// UpdatePromptEmbedding stores the unit-vector embedding for an assessment prompt.
func (r *DocketRepository) UpdatePromptEmbedding(ctx context.Context, promptID uuid.UUID, vec []float32) error {
	_, err := r.db.Exec(ctx,
		`UPDATE blueprint_prompts SET embedding = $1 WHERE id = $2`,
		pgvector.NewVector(vec), promptID)
	if err != nil {
		return fmt.Errorf("DocketRepository.UpdatePromptEmbedding: %w", err)
	}
	return nil
}

// DeleteDocumentChunks removes a blueprint's RAG chunks so they can be rebuilt.
func (r *DocketRepository) DeleteDocumentChunks(ctx context.Context, blueprintID uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM document_chunks WHERE blueprint_id = $1`, blueprintID); err != nil {
		return fmt.Errorf("DocketRepository.DeleteDocumentChunks: %w", err)
	}
	return nil
}

// InsertDocumentChunks inserts all chunks atomically.
func (r *DocketRepository) InsertDocumentChunks(ctx context.Context, chunks []model.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("DocketRepository.InsertDocumentChunks: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	for _, c := range chunks {
		if _, err := tx.Exec(ctx, `
			INSERT INTO document_chunks
				(blueprint_id, org_id, chunk_text, embedding, chunk_index, source)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT DO NOTHING`,
			c.BlueprintID, c.OrgID, c.ChunkText,
			pgvector.NewVector(c.Embedding), c.ChunkIndex, c.Source,
		); err != nil {
			return fmt.Errorf("DocketRepository.InsertDocumentChunks: insert: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("DocketRepository.InsertDocumentChunks: commit: %w", err)
	}
	return nil
}

// GetCriteriaEmbeddings returns criterion ID -> embedding for a blueprint.
// Criteria without an embedding are omitted; the map is empty if none exist.
func (r *DocketRepository) GetCriteriaEmbeddings(ctx context.Context, blueprintID uuid.UUID) (map[string][]float32, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, embedding::text FROM evaluation_criteria
		 WHERE docket_id = $1 AND embedding IS NOT NULL`,
		blueprintID)
	if err != nil {
		return nil, fmt.Errorf("DocketRepository.GetCriteriaEmbeddings: query: %w", err)
	}
	defer rows.Close()

	out := map[string][]float32{}
	for rows.Next() {
		var id uuid.UUID
		var text string
		if err := rows.Scan(&id, &text); err != nil {
			return nil, fmt.Errorf("DocketRepository.GetCriteriaEmbeddings: scan: %w", err)
		}
		var vec pgvector.Vector
		if err := vec.Parse(text); err != nil {
			return nil, fmt.Errorf("DocketRepository.GetCriteriaEmbeddings: parse vector: %w", err)
		}
		out[id.String()] = vec.Slice()
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DocketRepository.GetCriteriaEmbeddings: rows: %w", err)
	}
	return out, nil
}
