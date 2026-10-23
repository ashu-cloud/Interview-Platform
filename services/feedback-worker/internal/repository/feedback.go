package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/model"
)

type FeedbackRepository struct {
	pool *pgxpool.Pool
}

type EvalRepository = FeedbackRepository

func NewFeedbackRepository(pool *pgxpool.Pool) *FeedbackRepository {
	return &FeedbackRepository{pool: pool}
}

func (r *FeedbackRepository) Pool() *pgxpool.Pool {
	return r.pool
}

func NewEvalRepository(pool *pgxpool.Pool) *EvalRepository {
	return NewFeedbackRepository(pool)
}

// IsProcessed checks if the feedback report already exists for the session
func (r *FeedbackRepository) IsProcessed(ctx context.Context, sessionID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT count(1) FROM evaluation_reports WHERE interview_id = $1", sessionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query feedback report count: %w", err)
	}
	return count > 0, nil
}

// DequeueOutboxCompleted fetches pending completion events using FOR UPDATE SKIP LOCKED
func (r *FeedbackRepository) DequeueOutboxCompleted(ctx context.Context, limit int) ([]model.OutboxCompletedRow, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id::text, session_id::text, org_id::text, participant_id::text, blueprint_id::text, turn_count, retry_count
		FROM outbox_completed
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("dequeue outbox completed: %w", err)
	}
	defer rows.Close()

	var result []model.OutboxCompletedRow
	for rows.Next() {
		var row model.OutboxCompletedRow
		if err := rows.Scan(&row.ID, &row.SessionID, &row.OrgID, &row.ParticipantID, &row.BlueprintID, &row.TurnCount, &row.RetryCount); err != nil {
			return nil, fmt.Errorf("scan outbox completed: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox completed: %w", err)
	}
	return result, nil
}

// MarkOutboxCompleted marks an outbox row completed
func (r *FeedbackRepository) MarkOutboxCompleted(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox_completed
		SET status = 'completed', processed_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("mark outbox completed: %w", err)
	}
	return nil
}

// IncrementOutboxRetry increments retry count and sets to failed after 5 retries
func (r *FeedbackRepository) IncrementOutboxRetry(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox_completed
		SET retry_count = retry_count + 1,
		    status = CASE WHEN retry_count + 1 >= 5 THEN 'failed' ELSE 'pending' END,
		    failed_at = CASE WHEN retry_count + 1 >= 5 THEN NOW() ELSE NULL END
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("increment outbox retry: %w", err)
	}
	return nil
}

// SaveEvaluationReport persists evaluation report, per-criterion scores with grounding, and updates session status atomically
func (r *FeedbackRepository) SaveEvaluationReport(ctx context.Context, data model.EvaluationReportData) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	reportID := uuid.New().String()
	_, err = tx.Exec(ctx, `
		INSERT INTO evaluation_reports (id, interview_id, org_id, summary, overall_score, recommendation, transcript_json)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, reportID, data.SessionID, data.OrgID, data.Summary, data.OverallScore, data.Recommendation, string(data.TranscriptJSON))
	if err != nil {
		return "", fmt.Errorf("insert evaluation report: %w", err)
	}

	for _, score := range data.Scores {
		scoreID := uuid.New().String()

		var turnIDVal any = nil
		if score.EvidenceTurnID != "" {
			if _, parseErr := uuid.Parse(score.EvidenceTurnID); parseErr == nil {
				turnIDVal = score.EvidenceTurnID
			}
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO evaluation_scores (id, feedback_id, criterion_id, score, evidence, evidence_turn_id, is_grounded)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, scoreID, reportID, score.CriterionID, score.Score, score.EvidenceQuote, turnIDVal, score.IsGrounded)
		if err != nil {
			return "", fmt.Errorf("insert evaluation score: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE assessment_sessions
		SET status = 'evaluation_ready', updated_at = NOW()
		WHERE id = $1
	`, data.SessionID)
	if err != nil {
		return "", fmt.Errorf("update session status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit evaluation tx: %w", err)
	}

	return reportID, nil
}

// SaveFeedback transactionally inserts the report, scores, and updates interview status (legacy compatibility)
func (r *FeedbackRepository) SaveFeedback(ctx context.Context, interviewID string, orgID string, feedback *model.FeedbackResponse, transcriptJSON []byte) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert report
	feedbackID := uuid.New().String()
	_, err = tx.Exec(ctx,
		`INSERT INTO evaluation_reports (id, interview_id, org_id, summary, overall_score, recommendation, transcript_json) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		feedbackID, interviewID, orgID, feedback.Summary, feedback.OverallScore, feedback.Recommendation, string(transcriptJSON),
	)
	if err != nil {
		return "", fmt.Errorf("insert feedback report: %w", err)
	}

	// Insert scores
	for _, score := range feedback.Scores {
		_, err = tx.Exec(ctx,
			`INSERT INTO evaluation_scores (id, feedback_id, criterion_id, score, evidence)
			 VALUES ($1, $2, $3, $4, $5)`,
			uuid.New().String(), feedbackID, score.CriterionID, score.Score, score.Evidence,
		)
		if err != nil {
			return "", fmt.Errorf("insert feedback score: %w", err)
		}
	}

	// Update interview status
	_, err = tx.Exec(ctx,
		`UPDATE assessment_sessions SET status = 'feedback_ready', updated_at = NOW() WHERE id = $1`,
		interviewID,
	)
	if err != nil {
		return "", fmt.Errorf("update interview status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}

	return feedbackID, nil
}

