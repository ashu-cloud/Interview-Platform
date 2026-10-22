package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	KafkaBrokers           string        // optional — empty -> outbox mode
	PollInterval           time.Duration // default: 500ms
	DatabaseURL            string
	JobCandidateServiceURL string
	InternalAPIKey         string
	GroqAPIKey             string
	GroqModel              string
	LLMMode                string
}

func Load() (*Config, error) {
	pollInterval := 500 * time.Millisecond
	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		if d, err := time.ParseDuration(envPoll); err == nil && d > 0 {
			pollInterval = d
		}
	} else if envPoll := os.Getenv("OUTBOX_POLL_INTERVAL"); envPoll != "" {
		if d, err := time.ParseDuration(envPoll); err == nil && d > 0 {
			pollInterval = d
		}
	}

	serviceURL := os.Getenv("PLATFORM_SERVICE_URL")
	if serviceURL == "" {
		serviceURL = os.Getenv("JOB_CANDIDATE_SERVICE_URL")
	}

	llmMode := os.Getenv("EVALUATION_LLM_MODE")
	if llmMode == "" {
		llmMode = os.Getenv("FEEDBACK_LLM_MODE")
	}
	llmMode = strings.ToLower(strings.TrimSpace(llmMode))

	groqModel := strings.TrimSpace(os.Getenv("GROQ_MODEL"))
	if groqModel == "" {
		groqModel = "openai/gpt-oss-120b"
	}

	c := &Config{
		KafkaBrokers:           strings.TrimSpace(os.Getenv("KAFKA_BROKERS")),
		PollInterval:           pollInterval,
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		JobCandidateServiceURL: serviceURL,
		InternalAPIKey:         os.Getenv("INTERNAL_API_KEY"),
		GroqAPIKey:             os.Getenv("GROQ_API_KEY"),
		GroqModel:              groqModel,
		LLMMode:                llmMode,
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if c.JobCandidateServiceURL == "" {
		return nil, fmt.Errorf("PLATFORM_SERVICE_URL or JOB_CANDIDATE_SERVICE_URL is required")
	}
	if c.InternalAPIKey == "" {
		return nil, fmt.Errorf("INTERNAL_API_KEY is required")
	}
	if c.LLMMode == "" {
		c.LLMMode = "groq"
	}
	if c.LLMMode != "groq" && c.LLMMode != "mock" {
		return nil, fmt.Errorf("EVALUATION_LLM_MODE must be groq or mock")
	}
	if c.LLMMode == "groq" && c.GroqAPIKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY is required when EVALUATION_LLM_MODE is groq")
	}

	return c, nil
}

