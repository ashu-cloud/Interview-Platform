package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/client"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/config"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/llm"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/repository"
	"github.com/yourusername/ai-interview/services/feedback-worker/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to parse database config", "error", err)
		os.Exit(1)
	}
	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	evalRepo := repository.NewFeedbackRepository(pool)
	transcriptRepo := repository.NewTranscriptRepository(pool)
	platformClient := client.NewPlatformClient(cfg.JobCandidateServiceURL, cfg.InternalAPIKey)

	var llmClient llm.EvaluationLLM
	if cfg.LLMMode == "mock" {
		llmClient = llm.NewMockEvaluationLLMClient()
		logger.Info("using mock evaluation LLM")
	} else {
		llmClient = llm.NewGroqEvaluationLLMClient(cfg.GroqAPIKey, cfg.GroqModel)
		logger.Info("using Groq evaluation LLM", "model", cfg.GroqModel)
	}

	evaluator := worker.NewEvaluator(llmClient, logger)
	w := worker.NewEvaluationWorker(cfg, evalRepo, transcriptRepo, platformClient, evaluator, llmClient, logger, nil)

	// Start HTTP health server for cloud platforms (e.g. Render Web Services) requiring an open port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(`{"status":"ok","service":"feedback-worker"}`))
	})
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(`{"status":"running","service":"feedback-worker"}`))
	})
	healthServer := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		logger.Info("health server listening", "port", port)
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Warn("health server error", "error", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("shutting down evaluation worker")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		_ = healthServer.Shutdown(shutdownCtx)
		cancel()
		w.Close()
	}()

	if err := w.Run(ctx); err != nil {
		logger.Error("worker exited with error", "error", err)
		os.Exit(1)
	}
	logger.Info("worker stopped normally")
}
