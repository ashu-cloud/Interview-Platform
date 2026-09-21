package llm

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/model"
)

// DemoProvider produces a concise interview follow-up locally, allowing the
// full interview event and feedback flow to be demonstrated without API keys.
type DemoProvider struct{}

func NewDemoProvider() Provider { return DemoProvider{} }

func extractQuestions(prompt string) []string {
	var questions []string
	scanner := bufio.NewScanner(strings.NewReader(prompt))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 3 && line[0] >= '1' && line[0] <= '9' && line[1] == '.' && line[2] == ' ' {
			q := strings.TrimSpace(line[3:])
			if q != "" {
				questions = append(questions, q)
			}
		} else if len(line) > 4 && line[0] >= '1' && line[0] <= '9' && line[1] >= '0' && line[1] <= '9' && line[2] == '.' && line[3] == ' ' {
			q := strings.TrimSpace(line[4:])
			if q != "" {
				questions = append(questions, q)
			}
		}
	}
	return questions
}

func (DemoProvider) GenerateResponse(_ context.Context, systemPrompt string, messages []model.Message, streamChan chan<- string) error {
	questions := extractQuestions(systemPrompt)

	// Count how many user responses have occurred to determine which question to ask
	candidateAnswers := 0
	for _, m := range messages {
		if m.Role == "user" && !strings.Contains(m.Content, "You are conducting an interview") {
			candidateAnswers++
		}
	}

	var response string
	if candidateAnswers == 0 {
		// First turn: Greeting and initial question
		if len(questions) > 0 {
			response = fmt.Sprintf("Hello! Welcome to your interview. Let's get started: %s", questions[0])
		} else {
			response = "Hello! Welcome to your interview. Could you walk me through a recent project you worked on and your key responsibilities?"
		}
	} else if candidateAnswers < len(questions) {
		// Subsequent turns: Next question from docket
		nextQ := questions[candidateAnswers]
		response = fmt.Sprintf("Thank you for sharing that. Let's move to the next question: %s", nextQ)
	} else {
		// Exhausted questions or no questions: follow up
		if candidateAnswers == len(questions) && len(questions) > 0 {
			response = "Thank you. To follow up on that, what trade-offs did you consider and how did you measure the outcome?"
		} else {
			response = "Thanks for that answer. Could you walk me through a concrete example from your recent work?"
		}
	}

	streamChan <- response
	return nil
}
