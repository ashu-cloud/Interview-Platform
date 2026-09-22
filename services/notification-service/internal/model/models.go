package model

import "time"

// NotificationEvent represents the payload consumed from notification.events.
type NotificationEvent struct {
	EventType   string    `json:"event_type"`
	OrgID       string    `json:"org_id"`
	InterviewID string    `json:"interview_id"`
	CandidateID string    `json:"candidate_id"`
	FeedbackID  string    `json:"feedback_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// ResendEmailPayload represents the payload sent to Resend REST API.
type ResendEmailPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}
