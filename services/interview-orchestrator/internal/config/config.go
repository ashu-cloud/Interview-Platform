package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port                   string
	RedisURL               string
	KafkaBrokers           string
	JobCandidateServiceURL string
	InternalAPIKey         string
	DeepgramAPIKey         string
	GroqAPIKey             string
	GroqModel              string
	ElevenLabsAPIKey       string
	ElevenLabsVoiceID      string
	DemoMode               bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                   os.Getenv("PORT"),
		RedisURL:               os.Getenv("REDIS_URL"),
		KafkaBrokers:           os.Getenv("KAFKA_BROKERS"),
		JobCandidateServiceURL: os.Getenv("JOB_CANDIDATE_SERVICE_URL"),
		InternalAPIKey:         os.Getenv("INTERNAL_API_KEY"),
		DeepgramAPIKey:         os.Getenv("DEEPGRAM_API_KEY"),
		GroqAPIKey:             os.Getenv("GROQ_API_KEY"),
		GroqModel:              os.Getenv("GROQ_MODEL"),
		ElevenLabsAPIKey:       os.Getenv("ELEVENLABS_API_KEY"),
		ElevenLabsVoiceID:      os.Getenv("ELEVENLABS_VOICE_ID"),
		DemoMode:               strings.EqualFold(os.Getenv("INTERVIEW_DEMO_MODE"), "true"),
	}

	if cfg.Port == "" {
		cfg.Port = "8002"
	}
	if cfg.GroqModel == "" {
		cfg.GroqModel = "openai/gpt-oss-120b"
	}
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}
	if cfg.KafkaBrokers == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS is required")
	}
	if cfg.JobCandidateServiceURL == "" {
		return nil, fmt.Errorf("JOB_CANDIDATE_SERVICE_URL is required")
	}
	if cfg.InternalAPIKey == "" {
		return nil, fmt.Errorf("INTERNAL_API_KEY is required")
	}
	if !cfg.DemoMode && cfg.DeepgramAPIKey == "" {
		return nil, fmt.Errorf("DEEPGRAM_API_KEY is required")
	}
	if !cfg.DemoMode && cfg.GroqAPIKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY is required")
	}
	if !cfg.DemoMode && cfg.ElevenLabsAPIKey == "" {
		return nil, fmt.Errorf("ELEVENLABS_API_KEY is required")
	}
	if cfg.ElevenLabsVoiceID == "" {
		cfg.ElevenLabsVoiceID = "21m00Tcm4TlvDq8ikWAM" // Default Rachel
	}

	return cfg, nil
}
