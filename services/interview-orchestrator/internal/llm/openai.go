package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

type openAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(apiKey string) Provider {
	return &openAIProvider{
		client: openai.NewClient(option.WithAPIKey(apiKey)),
		model:  string(openai.ChatModelGPT4o),
	}
}

func NewGroqProvider(apiKey, modelName string) Provider {
	if modelName == "" {
		modelName = "openai/gpt-oss-120b"
	}
	return &openAIProvider{
		client: openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL("https://api.groq.com/openai/v1/"),
		),
		model: modelName,
	}
}

func (p *openAIProvider) GenerateResponse(ctx context.Context, systemPrompt string, messages []model.Message, streamChan chan<- string) error {
	oaiMessages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}

	for _, m := range messages {
		if m.Role == "user" {
			oaiMessages = append(oaiMessages, openai.UserMessage(m.Content))
		} else {
			oaiMessages = append(oaiMessages, openai.AssistantMessage(m.Content))
		}
	}

	stream := p.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:    openai.F(openai.ChatModel(p.model)),
		Messages: openai.F(oaiMessages),
	})
	
	for stream.Next() {
		evt := stream.Current()
		if len(evt.Choices) > 0 {
			chunk := evt.Choices[0].Delta.Content
			if chunk != "" {
				streamChan <- chunk
			}
		}
	}

	if err := stream.Err(); err != nil {
		return fmt.Errorf("openai stream error: %w", err)
	}

	return nil
}
