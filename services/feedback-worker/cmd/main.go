package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
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
		llmClient = llm.NewGroqEvaluationLLMClient(cfg.GroqAPIKey)
		logger.Info("using Groq evaluation LLM", "model", "llama-3.3-70b-versatile")
	}

	evaluator := worker.NewEvaluator(llmClient, logger)
	w := worker.NewEvaluationWorker(cfg, evalRepo, transcriptRepo, platformClient, evaluator, llmClient, logger, nil)

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("shutting down evaluation worker")
		cancel()
		w.Close()
	}()

	if err := w.Run(ctx); err != nil {
		logger.Error("worker exited with error", "error", err)
		os.Exit(1)
	}
	logger.Info("worker stopped normally")
}

