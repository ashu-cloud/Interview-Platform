package model

import (
	"time"

	"github.com/google/uuid"
)

type FeedbackScore struct {
	CriterionID uuid.UUID `json:"criterion_id"`
	Score       int       `json:"score"`
	Evidence    string    `json:"evidence"`
}

type FeedbackResponse struct {
	Summary        string          `json:"summary"`
	OverallScore   float64         `json:"overall_score"`
	Recommendation string          `json:"recommendation"`
	Scores         []FeedbackScore `json:"scores"`
}

type InterviewCompletedEvent struct {
	EventType   string    `json:"event_type"`
	InterviewID string    `json:"interview_id"`
	OrgID       string    `json:"org_id"`
	CandidateID string    `json:"candidate_id"`
	DocketID    string    `json:"docket_id"`
	CompletedAt time.Time `json:"completed_at"`
	TurnCount   int       `json:"turn_count"`
}

type TurnEvent struct {
	EventType   string    `json:"event_type"`
	InterviewID string    `json:"interview_id"`
	OrgID       string    `json:"org_id"`
	TurnNumber  int       `json:"turn_number"`
	Role        string    `json:"role"`
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
}

type NotificationEvent struct {
	EventType   string    `json:"event_type"`
	OrgID       string    `json:"org_id"`
	InterviewID string    `json:"interview_id"`
	CandidateID string    `json:"candidate_id"`
	FeedbackID  string    `json:"feedback_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// Minimal structures for Docker fetching
type Docket struct {
	ID     string       `json:"id"`
	Title  string       `json:"title"`
	Rubric []RubricItem `json:"rubric"`
}

type RubricItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxScore    int    `json:"max_score"`
}

type Candidate struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}
