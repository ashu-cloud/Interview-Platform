package model

import (
	"encoding/json"
	"fmt"
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

const (
	RoleParticipant = "participant"
	RoleAssessor    = "assessor"
)

type SessionCompletedEvent struct {
	EventType     string    `json:"event_type,omitempty"`
	SessionID     string    `json:"session_id"`
	OrgID         string    `json:"org_id"`
	ParticipantID string    `json:"participant_id"`
	BlueprintID   string    `json:"blueprint_id"`
	CompletedAt   time.Time `json:"completed_at,omitempty"`
	TurnCount     int       `json:"turn_count"`

	// Legacy field aliases
	InterviewID string `json:"-"`
	CandidateID string `json:"-"`
	DocketID    string `json:"-"`
}

func (e SessionCompletedEvent) MarshalJSON() ([]byte, error) {
	sessionID := firstNonEmpty(e.SessionID, e.InterviewID)
	participantID := firstNonEmpty(e.ParticipantID, e.CandidateID)
	blueprintID := firstNonEmpty(e.BlueprintID, e.DocketID)

	type Alias struct {
		EventType     string    `json:"event_type,omitempty"`
		SessionID     string    `json:"session_id"`
		OrgID         string    `json:"org_id"`
		ParticipantID string    `json:"participant_id"`
		BlueprintID   string    `json:"blueprint_id"`
		CompletedAt   time.Time `json:"completed_at,omitempty"`
		TurnCount     int       `json:"turn_count"`
	}

	return json.Marshal(&Alias{
		EventType:     e.EventType,
		SessionID:     sessionID,
		OrgID:         e.OrgID,
		ParticipantID: participantID,
		BlueprintID:   blueprintID,
		CompletedAt:   e.CompletedAt,
		TurnCount:     e.TurnCount,
	})
}

type InterviewCompletedEvent = SessionCompletedEvent

type TurnEvent struct {
	EventType   string    `json:"event_type,omitempty"`
	InterviewID string    `json:"session_id"`
	OrgID       string    `json:"org_id"`
	TurnNumber  int       `json:"turn_number"`
	Role        string    `json:"role"`
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
}

type completedEventPayload struct {
	EventType     string    `json:"event_type"`
	SessionID     string    `json:"session_id"`
	InterviewID   string    `json:"interview_id"`
	OrgID         string    `json:"org_id"`
	ParticipantID string    `json:"participant_id"`
	CandidateID   string    `json:"candidate_id"`
	BlueprintID   string    `json:"blueprint_id"`
	DocketID      string    `json:"docket_id"`
	CompletedAt   time.Time `json:"completed_at"`
	TurnCount     int       `json:"turn_count"`
}

type turnEventPayload struct {
	EventType   string    `json:"event_type"`
	SessionID   string    `json:"session_id"`
	InterviewID string    `json:"interview_id"`
	OrgID       string    `json:"org_id"`
	TurnNumber  int       `json:"turn_number"`
	Role        string    `json:"role"`
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
}

func ParseSessionCompletedEvent(data []byte) (SessionCompletedEvent, error) {
	var payload completedEventPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return SessionCompletedEvent{}, fmt.Errorf("unmarshal completed event: %w", err)
	}

	sessionID := firstNonEmpty(payload.SessionID, payload.InterviewID)
	participantID := firstNonEmpty(payload.ParticipantID, payload.CandidateID)
	blueprintID := firstNonEmpty(payload.BlueprintID, payload.DocketID)

	return SessionCompletedEvent{
		EventType:     payload.EventType,
		SessionID:     sessionID,
		InterviewID:   sessionID,
		OrgID:         payload.OrgID,
		ParticipantID: participantID,
		CandidateID:   participantID,
		BlueprintID:   blueprintID,
		DocketID:      blueprintID,
		CompletedAt:   payload.CompletedAt,
		TurnCount:     payload.TurnCount,
	}, nil
}

func ParseInterviewCompletedEvent(data []byte) (InterviewCompletedEvent, error) {
	return ParseSessionCompletedEvent(data)
}

func ParseTurnEvent(data []byte) (TurnEvent, error) {
	var payload turnEventPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return TurnEvent{}, fmt.Errorf("unmarshal turn event: %w", err)
	}

	role, err := normalizeRole(payload.Role)
	if err != nil {
		return TurnEvent{}, err
	}

	return TurnEvent{
		EventType:   payload.EventType,
		InterviewID: firstNonEmpty(payload.SessionID, payload.InterviewID),
		OrgID:       payload.OrgID,
		TurnNumber:  payload.TurnNumber,
		Role:        role,
		Content:     payload.Content,
		Timestamp:   payload.Timestamp,
	}, nil
}

func normalizeRole(role string) (string, error) {
	switch role {
	case RoleParticipant, "user":
		return RoleParticipant, nil
	case RoleAssessor, "assistant":
		return RoleAssessor, nil
	default:
		return "", fmt.Errorf("unsupported turn role %q", role)
	}
}

func firstNonEmpty(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

type NotificationEvent struct {
	EventType   string    `json:"event_type"`
	OrgID       string    `json:"org_id"`
	InterviewID string    `json:"interview_id"`
	CandidateID string    `json:"candidate_id"`
	FeedbackID  string    `json:"feedback_id"`
	Timestamp   time.Time `json:"timestamp"`
}

type EvaluationCriterion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MaxScore    int    `json:"max_score"`
}

type RubricItem = EvaluationCriterion

type AssessmentBlueprint struct {
	ID       string                `json:"id"`
	Title    string                `json:"title"`
	Criteria []EvaluationCriterion `json:"criteria"`
	Rubric   []EvaluationCriterion `json:"rubric,omitempty"`
}

type Docket = AssessmentBlueprint

type Participant struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}

type Candidate = Participant

type OutboxCompletedRow struct {
	ID            string
	SessionID     string
	OrgID         string
	ParticipantID string
	BlueprintID   string
	TurnCount     int
	RetryCount    int
}

type EvaluationScoreData struct {
	CriterionID    string
	Score          int
	EvidenceQuote  string
	EvidenceTurnID string
	IsGrounded     bool
}

type EvaluationReportData struct {
	SessionID      string
	OrgID          string
	Summary        string
	OverallScore   float64
	Recommendation string
	Scores         []EvaluationScoreData
	TranscriptJSON []byte
}

