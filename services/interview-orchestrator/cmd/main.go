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

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/config"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/handler"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/integration"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/kafka"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/llm"
	"github.com/yourusername/ai-interview/services/interview-orchestrator/internal/session"
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

	kafkaBrokers := strings.Split(cfg.KafkaBrokers, ",")
	kafkaProd := kafka.NewProducer(kafkaBrokers)
	defer func() { _ = kafkaProd.Close() }()

	wsHandler := handler.NewWebSocketHandler(jcsClient, sessMgr, sttAdapter, llmPool, ttsAdapter, kafkaProd)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})
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
