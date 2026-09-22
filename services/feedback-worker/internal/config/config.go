package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers           string
	DatabaseURL            string
	JobCandidateServiceURL string
	InternalAPIKey         string
	GroqAPIKey             string
	LLMMode                string
}

func Load() (*Config, error) {
	c := &Config{
		KafkaBrokers:           os.Getenv("KAFKA_BROKERS"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		JobCandidateServiceURL: os.Getenv("JOB_CANDIDATE_SERVICE_URL"),
		InternalAPIKey:         os.Getenv("INTERNAL_API_KEY"),
		GroqAPIKey:             os.Getenv("GROQ_API_KEY"),
		LLMMode:                strings.ToLower(strings.TrimSpace(os.Getenv("FEEDBACK_LLM_MODE"))),
	}

	if c.KafkaBrokers == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS is required")
	}
	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if c.JobCandidateServiceURL == "" {
		return nil, fmt.Errorf("JOB_CANDIDATE_SERVICE_URL is required")
	}
	if c.InternalAPIKey == "" {
		return nil, fmt.Errorf("INTERNAL_API_KEY is required")
	}
	if c.LLMMode == "" {
		c.LLMMode = "groq"
	}
	if c.LLMMode != "groq" && c.LLMMode != "mock" {
		return nil, fmt.Errorf("FEEDBACK_LLM_MODE must be groq or mock")
	}
	if c.LLMMode == "groq" && c.GroqAPIKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY is required when FEEDBACK_LLM_MODE is groq")
	}

	return c, nil
}
