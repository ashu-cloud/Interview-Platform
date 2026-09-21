package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

type anthropicProvider struct {
	client *anthropic.Client
}

func NewAnthropicProvider(apiKey string) Provider {
	return &anthropicProvider{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
	}
}

func (p *anthropicProvider) GenerateResponse(ctx context.Context, systemPrompt string, messages []model.Message, streamChan chan<- string) error {
	anthropicMessages := []anthropic.MessageParam{}
	for _, m := range messages {
		if m.Role == "user" {
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		} else {
			anthropicMessages = append(anthropicMessages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}

	stream := p.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude3_5SonnetLatest),
		System:    anthropic.F([]anthropic.TextBlockParam{anthropic.NewTextBlock(systemPrompt)}),
		MaxTokens: anthropic.F(int64(1000)),
		Messages:  anthropic.F(anthropicMessages),
	})

	for stream.Next() {
		evt := stream.Current()
		switch e := evt.AsUnion().(type) {
		case anthropic.ContentBlockDeltaEvent:
			if textDelta, ok := e.Delta.AsUnion().(anthropic.TextDelta); ok {
				streamChan <- textDelta.Text
			}
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("anthropic stream error: %w", err)
	}

	return nil
}
