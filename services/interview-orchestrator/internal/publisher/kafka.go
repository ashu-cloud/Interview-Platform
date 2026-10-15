package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
)

type kafkaWriter interface {
	WriteMessages(context.Context, ...kafka.Message) error
	Close() error
}

type kafkaPublisher struct {
	writer         kafkaWriter
	turnsTopic     string
	completedTopic string
}

func newKafkaPublisher(cfg Config) Publisher {
	return &kafkaPublisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(cfg.KafkaBrokers...),
			AllowAutoTopicCreation: true,
			Balancer:               &kafka.Hash{},
		},
		turnsTopic:     cfg.TurnsTopic,
		completedTopic: cfg.CompletedTopic,
	}
}

// NewKafkaPublisher creates a KafkaPublisher using broker string (comma-separated or single).
func NewKafkaPublisher(kafkaBrokers string) Publisher {
	brokers := cleanBrokers(strings.Split(kafkaBrokers, ","))
	return newKafkaPublisher(Config{
		KafkaBrokers:   brokers,
		TurnsTopic:     DefaultTurnsTopic,
		CompletedTopic: DefaultCompletedTopic,
	})
}

func (p *kafkaPublisher) PublishTurn(ctx context.Context, sessionID, orgID string, turn int, role, content string) error {
	event := turnEvent{
		EventType:  "turn",
		SessionID:  sessionID,
		OrgID:      orgID,
		TurnNumber: turn,
		Role:       role,
		Content:    content,
		Timestamp:  nowUTC(),
	}
	if err := validateTurn(event); err != nil {
		return err
	}
	return p.write(ctx, p.turnsTopic, event.SessionID, event)
}

func (p *kafkaPublisher) PublishCompleted(ctx context.Context, sessionID, orgID, participantID, blueprintID string, turnCount int) error {
	event := completedEvent{
		EventType:     "completed",
		SessionID:     sessionID,
		OrgID:         orgID,
		ParticipantID: participantID,
		BlueprintID:   blueprintID,
		TurnCount:     turnCount,
		CompletedAt:   nowUTC(),
	}
	if err := validateCompleted(event); err != nil {
		return err
	}
	return p.write(ctx, p.completedTopic, event.SessionID, event)
}

func (p *kafkaPublisher) write(ctx context.Context, topic, key string, event any) error {
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if err := p.writer.WriteMessages(ctx, kafka.Message{Topic: topic, Key: []byte(key), Value: value}); err != nil {
		return fmt.Errorf("publish to %s: %w", topic, err)
	}
	return nil
}

func (p *kafkaPublisher) Close() error {
	return p.writer.Close()
}
