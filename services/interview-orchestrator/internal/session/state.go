package session

import "time"

// ConversationTurn represents a single conversational turn in an interview session.
type ConversationTurn struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionState represents the full state stored in Redis for an ongoing interview.
type SessionState struct {
	InterviewID         string             `json:"interview_id"`
	CandidateID         string             `json:"candidate_id"`
	OrgID               string             `json:"org_id"`
	DocketID            string             `json:"docket_id"`
	Docket              any                `json:"docket"`
	ConversationHistory []ConversationTurn `json:"conversation_history"`
	TurnNumber          int                `json:"turn_number"`
	Status              string             `json:"status"`
	StartedAt           time.Time          `json:"started_at"`
	LastActive          time.Time          `json:"last_active"`
}
