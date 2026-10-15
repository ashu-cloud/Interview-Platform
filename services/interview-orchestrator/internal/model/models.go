package model

import (
	"time"
)

type DocketQuestion struct {
	ID           string `json:"id"`
	QuestionText string `json:"question_text"`
	QuestionType string `json:"question_type"`
	OrderIndex   int    `json:"order_index"`
	FollowUpHint *string `json:"follow_up_hint,omitempty"`
}

type RubricCriterion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxScore    int    `json:"max_score"`
}

type Docket struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Instructions string            `json:"instructions"`
	DurationMins int               `json:"duration_mins"`
	Questions    []DocketQuestion  `json:"questions"`
	Rubric       []RubricCriterion `json:"rubric"`
}

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Session struct {
	InterviewID         string    `json:"interview_id"`
	CandidateID         string    `json:"candidate_id"`
	OrgID               string    `json:"org_id"`
	DocketID            string    `json:"docket_id"`
	Docket              Docket    `json:"docket"`
	ConversationHistory []Message `json:"conversation_history"`
	TurnNumber          int       `json:"turn_number"`
	Status              string    `json:"status"` // in_progress | completed
	StartedAt           time.Time `json:"started_at"`
	LastActive          time.Time `json:"last_active"`

	// Adaptive engine state. Both are empty when criteria embeddings are
	// unavailable, which disables the adaptive engine for the session.
	CoverageMatrix     map[string]float64   `json:"coverage_matrix,omitempty"`
	CriteriaEmbeddings map[string][]float32 `json:"criteria_embeddings,omitempty"`
}
