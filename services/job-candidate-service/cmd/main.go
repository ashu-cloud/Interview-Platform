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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/cache"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/config"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/db"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/embedding"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/handler"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/projector"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/repository"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/service"
)

func main() {
	// ── slog structured logging ───────────────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ── Load .env file (check local directory first, then root directory) ────
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../../.env")

	// ── Config — fail fast if any required variable is missing ───────────────
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// ── Database connection pool ──────────────────────────────────────────────
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to postgres")

	// ── Redis client ─────────────────────────────────────────────────────────
	redisClient, err := cache.NewClient(cfg.RedisURL)
	if err != nil {
		slog.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer func() { _ = redisClient.Close() }()
	slog.Info("connected to redis")

	// ── Repositories ─────────────────────────────────────────────────────────
	userRepo := repository.NewUserRepository(pool)
	jobRepo := repository.NewJobRepository(pool)
	docketRepo := repository.NewDocketRepository(pool)
	candidateRepo := repository.NewCandidateRepository(pool)
	inviteRepo := repository.NewInviteRepository(pool)

	// ── Services ─────────────────────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	jobSvc := service.NewJobService(jobRepo)
	docketSvc := service.NewDocketService(docketRepo, jobRepo)
	var openAIClient embedding.EmbeddingClient
	if cfg.OpenAIEmbeddingAPIKey != "" && strings.HasPrefix(cfg.OpenAIEmbeddingAPIKey, "sk-") && !cfg.DemoMode {
		slog.Info("blueprint embeddings: configured with OpenAI upstream", "model", cfg.EmbeddingModel, "dims", cfg.EmbeddingDims)
		openAIClient = embedding.NewOpenAIEmbeddingClient(cfg.OpenAIEmbeddingAPIKey, cfg.EmbeddingDims).
			WithModel(cfg.EmbeddingModel).
			WithBaseURL(cfg.EmbeddingBaseURL)
	} else {
		slog.Info("blueprint embeddings: configured with local deterministic semantic engine", "dims", cfg.EmbeddingDims)
	}
	docketSvc.SetEmbedder(embedding.NewFallbackEmbeddingClient(openAIClient, cfg.EmbeddingDims), logger)
	candidateSvc := service.NewCandidateService(candidateRepo, jobRepo)
	inviteSvc := service.NewInviteService(inviteRepo, candidateRepo, docketRepo, redisClient, cfg.AppDomain)

	// ── Handlers and router ──────────────────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	r := newRouter(routeHandlers{
		auth:      handler.NewAuthHandler(authSvc),
		job:       handler.NewJobHandler(jobSvc),
		docket:    handler.NewDocketHandler(docketSvc),
		candidate: handler.NewCandidateHandler(candidateSvc),
		invite:    handler.NewInviteHandler(inviteSvc),
	}, cfg.JWTSecret, cfg.InternalAPIKey)

	// ── HTTP server with graceful shutdown ────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("job-candidate-service starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe failed", "error", err)
			os.Exit(1)
		}
	}()

	// ── Embedded turn projector (Railway outbox mode) ───────────────────────
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	projRepo := projector.NewTurnRepository(pool)
	proj := projector.NewTurnProjector(projRepo, logger, 500*time.Millisecond)
	go func() {
		logger.Info("turn projector started (embedded, outbox mode)")
		proj.RunOutboxMode(serverCtx)
		logger.Info("turn projector stopped")
	}()

	// Block until SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gracefully...")
	serverCancel()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
