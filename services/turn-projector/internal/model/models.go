package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Role values stored in session_turns / outbox_turns.
const (
	RoleParticipant = "participant"
	RoleAssessor    = "assessor"
	RoleSystem      = "system"
)

// SessionTurnEvent is the shared shape for both Kafka messages and outbox rows.
type SessionTurnEvent struct {
	ID         string    `json:"-"` // outbox row ID — empty when from Kafka
	SessionID  string    `json:"session_id"`
	OrgID      string    `json:"org_id"`
	TurnNumber int       `json:"turn_number"`
	Role       string    `json:"role"` // "participant" or "assessor"
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
}

// wireTurnEvent accepts both the current runtime payload (interview_id,
// role user/assistant) and the session_id/participant/assessor shape.
type wireTurnEvent struct {
	SessionID   string    `json:"session_id"`
	InterviewID string    `json:"interview_id"`
	OrgID       string    `json:"org_id"`
	TurnNumber  int       `json:"turn_number"`
	Role        string    `json:"role"`
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
}

// NormalizeRole maps runtime role values onto session_turns roles.
func NormalizeRole(role string) (string, error) {
	switch role {
	case "user", RoleParticipant:
		return RoleParticipant, nil
	case "assistant", RoleAssessor:
		return RoleAssessor, nil
	case RoleSystem:
		return RoleSystem, nil
	default:
		return "", fmt.Errorf("unsupported role %q", role)
	}
}

// ParseTurnEvent decodes and validates a Kafka message value.
func ParseTurnEvent(data []byte) (SessionTurnEvent, error) {
	var w wireTurnEvent
	if err := json.Unmarshal(data, &w); err != nil {
		return SessionTurnEvent{}, fmt.Errorf("invalid json: %w", err)
	}
	sessionID := w.SessionID
	if sessionID == "" {
		sessionID = w.InterviewID
	}
	if _, err := uuid.Parse(sessionID); err != nil {
		return SessionTurnEvent{}, fmt.Errorf("invalid session id %q: %w", sessionID, err)
	}
	if _, err := uuid.Parse(w.OrgID); err != nil {
		return SessionTurnEvent{}, fmt.Errorf("invalid org_id %q: %w", w.OrgID, err)
	}
	if w.TurnNumber < 0 {
		return SessionTurnEvent{}, errors.New("turn_number must not be negative")
	}
	role, err := NormalizeRole(w.Role)
	if err != nil {
		return SessionTurnEvent{}, err
	}
	if w.Content == "" {
		return SessionTurnEvent{}, errors.New("content must not be empty")
	}
	return SessionTurnEvent{
		SessionID:  sessionID,
		OrgID:      w.OrgID,
		TurnNumber: w.TurnNumber,
		Role:       role,
		Content:    w.Content,
		Timestamp:  w.Timestamp,
	}, nil
}
