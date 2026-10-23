package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                   string
	RedisURL               string
	KafkaBrokers           string
	KafkaTurnsTopic        string
	KafkaCompletedTopic    string
	DatabaseURL            string
	JobCandidateServiceURL string
	InternalAPIKey         string
	DeepgramAPIKey         string
	GroqAPIKey             string
	GroqModel              string
	ElevenLabsAPIKey       string
	ElevenLabsVoiceID      string
	DemoMode               bool
	MinIOEndpoint          string
	MinIOAccessKey         string
	MinIOSecretKey         string
	MinIOBucketAudio       string
	MinIOUseSSL            bool
	VoiceDebug             bool

	OpenAIEmbeddingAPIKey string // required unless ASSESSMENT_DEMO_MODE=true
	EmbeddingModel        string
	EmbeddingDims         int
	EmbeddingBaseURL      string
}

func Load() (*Config, error) {
	minioUser := strings.TrimSpace(os.Getenv("MINIO_ROOT_USER"))
	if minioUser == "" {
		minioUser = strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY"))
	}
	minioSecret := strings.TrimSpace(os.Getenv("MINIO_ROOT_PASSWORD"))
	if minioSecret == "" {
		minioSecret = strings.TrimSpace(os.Getenv("MINIO_SECRET_KEY"))
	}

	embKey := strings.TrimSpace(os.Getenv("EMBEDDING_API_KEY"))
	if embKey == "" {
		embKey = strings.TrimSpace(os.Getenv("OPENAI_EMBEDDING_API_KEY"))
	}

	kafkaBrokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	isCloud := os.Getenv("RENDER") != "" || os.Getenv("RAILWAY_ENVIRONMENT") != "" || os.Getenv("FLY_APP_NAME") != ""
	isLocalhost := strings.Contains(kafkaBrokers, "localhost") || strings.Contains(kafkaBrokers, "127.0.0.1")
	if strings.EqualFold(os.Getenv("USE_OUTBOX"), "true") || strings.EqualFold(os.Getenv("DISABLE_KAFKA"), "true") || (isCloud && isLocalhost) {
		kafkaBrokers = ""
	}

	cfg := &Config{
		Port:                   os.Getenv("PORT"),
		RedisURL:               CleanRedisURL(os.Getenv("REDIS_URL")),
		KafkaBrokers:           kafkaBrokers,
		KafkaTurnsTopic:        strings.TrimSpace(os.Getenv("KAFKA_TURNS_TOPIC")),
		KafkaCompletedTopic:    strings.TrimSpace(os.Getenv("KAFKA_COMPLETED_TOPIC")),
		DatabaseURL:            strings.TrimSpace(os.Getenv("DATABASE_URL")),
		JobCandidateServiceURL: os.Getenv("JOB_CANDIDATE_SERVICE_URL"),
		InternalAPIKey:         os.Getenv("INTERNAL_API_KEY"),
		DeepgramAPIKey:         os.Getenv("DEEPGRAM_API_KEY"),
		GroqAPIKey:             os.Getenv("GROQ_API_KEY"),
		GroqModel:              os.Getenv("GROQ_MODEL"),
		ElevenLabsAPIKey:       os.Getenv("ELEVENLABS_API_KEY"),
		ElevenLabsVoiceID:      os.Getenv("ELEVENLABS_VOICE_ID"),
		DemoMode:               strings.EqualFold(os.Getenv("ASSESSMENT_DEMO_MODE"), "true") || strings.EqualFold(os.Getenv("INTERVIEW_DEMO_MODE"), "true"),
		MinIOEndpoint:          strings.TrimSpace(os.Getenv("MINIO_ENDPOINT")),
		MinIOAccessKey:         minioUser,
		MinIOSecretKey:         minioSecret,
		MinIOBucketAudio:       strings.TrimSpace(os.Getenv("MINIO_BUCKET_AUDIO")),
		MinIOUseSSL:            strings.EqualFold(os.Getenv("MINIO_USE_SSL"), "true"),
		VoiceDebug:             strings.EqualFold(os.Getenv("VOICE_DEBUG"), "true"),

		OpenAIEmbeddingAPIKey: embKey,
		EmbeddingModel:        strings.TrimSpace(os.Getenv("EMBEDDING_MODEL")),
		EmbeddingBaseURL:      strings.TrimSpace(os.Getenv("EMBEDDING_BASE_URL")),
	}

	cfg.EmbeddingDims = 768
	if raw := strings.TrimSpace(os.Getenv("EMBEDDING_DIMS")); raw != "" {
		dims, err := strconv.Atoi(raw)
		if err != nil || dims <= 0 {
			return nil, fmt.Errorf("EMBEDDING_DIMS must be a positive integer, got %q", raw)
		}
		cfg.EmbeddingDims = dims
	}
	if cfg.EmbeddingModel == "" {
		cfg.EmbeddingModel = "text-embedding-3-small"
	}

	if cfg.Port == "" {
		cfg.Port = "8002"
	}
	if cfg.GroqModel == "" {
		cfg.GroqModel = "openai/gpt-oss-120b"
	}
	if cfg.MinIOBucketAudio == "" {
		cfg.MinIOBucketAudio = "assessment-audio"
	}
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}
	if cfg.KafkaTurnsTopic == "" {
		cfg.KafkaTurnsTopic = "session.turns"
	}
	if cfg.KafkaCompletedTopic == "" {
		cfg.KafkaCompletedTopic = "session.completed"
	}
	if cfg.KafkaBrokers == "" && cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required when KAFKA_BROKERS is not set")
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

// CleanRedisURL normalizes a Redis connection string.
// Handles cases where users paste raw CLI commands copied from provider dashboards
// (e.g. `redis-cli --tls -u redis://default:xxx@host:6379`), removes quotes,
// and enforces `rediss://` for Upstash or when TLS was requested.
func CleanRedisURL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, `"'`)

	isTLS := strings.Contains(s, "--tls")

	if idx := strings.Index(s, "rediss://"); idx != -1 {
		s = s[idx:]
	} else if idx := strings.Index(s, "redis://"); idx != -1 {
		s = s[idx:]
	}

	if endIdx := strings.IndexAny(s, " \t\r\n\"'"); endIdx != -1 {
		s = s[:endIdx]
	}

	if strings.HasPrefix(s, "redis://") {
		lower := strings.ToLower(s)
		if isTLS || strings.Contains(lower, "upstash.io") {
			s = "rediss://" + s[len("redis://"):]
		}
	}

	return s
}
