package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/adaptive"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/config"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/embedding"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/handler"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/integration"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/llm"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/publisher"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/rag"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/session"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/storage"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/stt"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/tts"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to parse redis url", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewClient(opt)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer func() { _ = redisClient.Close() }()
	slog.Info("connected to redis")

	jcsClient := integration.NewJCSClient(cfg.JobCandidateServiceURL, cfg.InternalAPIKey)
	sessMgr := session.NewManager(redisClient)

	sttAdapter := stt.NewDeepgramAdapter(cfg.DeepgramAPIKey)
	var llmPool *llm.Pool
	ttsAdapter := tts.NewElevenLabsAdapter(cfg.ElevenLabsAPIKey, cfg.ElevenLabsVoiceID)
	if cfg.DemoMode {
		slog.Info("using local interview demo providers")
		sttAdapter = stt.NewDemoAdapter()
		demoProvider := llm.NewDemoProvider()
		llmPool = llm.NewPool(demoProvider, demoProvider)
		ttsAdapter = tts.NewDemoAdapter()
	} else {
		slog.Info("using Groq as LLM provider", "model", cfg.GroqModel)
		primaryLLM := llm.NewGroqProvider(cfg.GroqAPIKey, cfg.GroqModel)
		fallbackModel := os.Getenv("GROQ_FALLBACK_MODEL")
		if fallbackModel == "" {
			fallbackModel = "openai/gpt-oss-20b"
		}
		secondaryLLM := llm.NewGroqProvider(cfg.GroqAPIKey, fallbackModel)
		llmPool = llm.NewPool(primaryLLM, secondaryLLM)
	}

	var kafkaBrokers []string
	if cfg.KafkaBrokers != "" {
		kafkaBrokers = strings.Split(cfg.KafkaBrokers, ",")
	}
	publisherCtx, cancelPublisherInit := context.WithTimeout(context.Background(), 10*time.Second)
	eventPublisher, err := publisher.New(publisherCtx, publisher.Config{
		KafkaBrokers:   kafkaBrokers,
		DatabaseURL:    cfg.DatabaseURL,
		TurnsTopic:     cfg.KafkaTurnsTopic,
		CompletedTopic: cfg.KafkaCompletedTopic,
	})
	cancelPublisherInit()
	if err != nil {
		slog.Error("failed to initialize event publisher", "error", err)
		os.Exit(1)
	}
	defer func() { _ = eventPublisher.Close() }()
	if len(kafkaBrokers) > 0 {
		slog.Info("using Kafka event publisher", "turns_topic", cfg.KafkaTurnsTopic, "completed_topic", cfg.KafkaCompletedTopic)
	} else {
		slog.Info("using Postgres outbox event publisher")
	}

	var dbPool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		poolCtx, poolCancel := context.WithTimeout(context.Background(), 5*time.Second)
		p, poolErr := pgxpool.New(poolCtx, cfg.DatabaseURL)
		poolCancel()
		if poolErr == nil {
			dbPool = p
			defer dbPool.Close()
		} else {
			slog.Warn("could not connect database pool for audio recordings", "error", poolErr)
		}
	}

	audioOffloader := storage.NewAudioOffloader(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucketAudio,
		cfg.MinIOUseSSL,
		dbPool,
		logger,
	)
	audioOffloader.Start()
	defer audioOffloader.Stop()

	var openAIClient embedding.EmbeddingClient
	if cfg.OpenAIEmbeddingAPIKey != "" && strings.HasPrefix(cfg.OpenAIEmbeddingAPIKey, "sk-") && !cfg.DemoMode {
		slog.Info("orchestrator embeddings: configured with OpenAI upstream", "model", cfg.EmbeddingModel, "dims", cfg.EmbeddingDims)
		openAIClient = embedding.NewOpenAIEmbeddingClient(cfg.OpenAIEmbeddingAPIKey, cfg.EmbeddingDims).
			WithModel(cfg.EmbeddingModel).
			WithBaseURL(cfg.EmbeddingBaseURL)
	} else {
		slog.Info("orchestrator embeddings: configured with local deterministic semantic engine", "dims", cfg.EmbeddingDims)
	}
	embClient := embedding.NewFallbackEmbeddingClient(openAIClient, cfg.EmbeddingDims)
	adaptiveEngine := adaptive.NewEngine(embClient)
	ragPrefetcher := rag.NewPrefetcher(dbPool, redisClient, embClient, 150*time.Millisecond)
	if dbPool == nil {
		slog.Warn("database pool unavailable, RAG prefetch disabled")
	}

	wsHandler := handler.NewWebSocketHandler(jcsClient, sessMgr, sttAdapter, llmPool, ttsAdapter, eventPublisher)
	wsHandler.SetAudioOffloader(audioOffloader)
	wsHandler.SetAdaptive(adaptiveEngine, ragPrefetcher)
	wsHandler.SetDBPool(dbPool)
	wsHandler.SetVoiceDebug(cfg.VoiceDebug)

	debugHandler := handler.NewDebugHandler(cfg, redisClient, llmPool, ttsAdapter)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"session-runtime"}`))
	})
	mux.HandleFunc("/debug/voice", debugHandler.HandleDebugVoice)
	mux.HandleFunc("/ws/interview/", wsHandler.HandleWS)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("interview-orchestrator starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gracefully...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
