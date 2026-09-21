package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	PublishTurn(ctx context.Context, interviewID, orgID string, turnNumber int, role, content string) error
	PublishCompleted(ctx context.Context, interviewID, orgID, candidateID, docketID string, completedAt time.Time, turnCount int) error
	Close() error
}

type kafkaProducer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) Producer {
	// A single writer configured without a specific topic.
	// The topic is specified per-message.
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		AllowAutoTopicCreation: true,
		Balancer:               &kafka.Hash{},
	}

	return &kafkaProducer{
		writer: writer,
	}
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

func (p *kafkaProducer) PublishTurn(ctx context.Context, interviewID, orgID string, turnNumber int, role, content string) error {
	evt := TurnEvent{
		EventType:   "turn",
		InterviewID: interviewID,
		OrgID:       orgID,
		TurnNumber:  turnNumber,
		Role:        role,
		Content:     content,
		Timestamp:   time.Now().UTC(),
	}

	b, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal turn event failed: %w", err)
	}

	msg := kafka.Message{
		Topic: "interview.turns",
		Key:   []byte(interviewID),
		Value: b,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("publish turn failed: %w", err)
	}
	return nil
}

type CompletedEvent struct {
	EventType   string    `json:"event_type"`
	InterviewID string    `json:"interview_id"`
	OrgID       string    `json:"org_id"`
	CandidateID string    `json:"candidate_id"`
	DocketID    string    `json:"docket_id"`
	CompletedAt time.Time `json:"completed_at"`
	TurnCount   int       `json:"turn_count"`
}

func (p *kafkaProducer) PublishCompleted(ctx context.Context, interviewID, orgID, candidateID, docketID string, completedAt time.Time, turnCount int) error {
	evt := CompletedEvent{
		EventType:   "completed",
		InterviewID: interviewID,
		OrgID:       orgID,
		CandidateID: candidateID,
		DocketID:    docketID,
		CompletedAt: completedAt,
		TurnCount:   turnCount,
	}

	b, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal completed event failed: %w", err)
	}

	msg := kafka.Message{
		Topic: "interview.completed",
		Key:   []byte(interviewID),
		Value: b,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("publish completed failed: %w", err)
	}
	return nil
}

func (p *kafkaProducer) Close() error {
	return p.writer.Close()
}
