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

func NewFeedbackRepository(pool *pgxpool.Pool) *FeedbackRepository {
	return &FeedbackRepository{pool: pool}
}

// IsProcessed checks if the feedback report already exists for the interview
func (r *FeedbackRepository) IsProcessed(ctx context.Context, interviewID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT count(1) FROM feedback_reports WHERE interview_id = $1", interviewID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query feedback report count: %w", err)
	}
	return count > 0, nil
}

// SaveFeedback transactionally inserts the report, scores, and updates interview status
func (r *FeedbackRepository) SaveFeedback(ctx context.Context, interviewID string, orgID string, feedback *model.FeedbackResponse, transcriptJSON []byte) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert report
	feedbackID := uuid.New().String()
	_, err = tx.Exec(ctx,
		`INSERT INTO feedback_reports (id, interview_id, org_id, summary, overall_score, recommendation, transcript_json) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		feedbackID, interviewID, orgID, feedback.Summary, feedback.OverallScore, feedback.Recommendation, transcriptJSON,
	)
	if err != nil {
		return "", fmt.Errorf("insert feedback report: %w", err)
	}

	// Insert scores
	for _, score := range feedback.Scores {
		_, err = tx.Exec(ctx,
			`INSERT INTO feedback_scores (id, feedback_id, criterion_id, score, evidence)
			 VALUES ($1, $2, $3, $4, $5)`,
			uuid.New().String(), feedbackID, score.CriterionID, score.Score, score.Evidence,
		)
		if err != nil {
			return "", fmt.Errorf("insert feedback score: %w", err)
		}
	}

	// Update interview status
	_, err = tx.Exec(ctx,
		`UPDATE interviews SET status = 'feedback_ready', updated_at = NOW() WHERE id = $1`,
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
