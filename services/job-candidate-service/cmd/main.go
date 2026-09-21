package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/cache"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/config"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/db"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/handler"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/middleware"
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
	candidateSvc := service.NewCandidateService(candidateRepo, jobRepo)
	inviteSvc := service.NewInviteService(inviteRepo, candidateRepo, docketRepo, redisClient, cfg.AppDomain)

	// ── Handlers ─────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(authSvc)
	jobHandler := handler.NewJobHandler(jobSvc)
	docketHandler := handler.NewDocketHandler(docketSvc)
	candidateHandler := handler.NewCandidateHandler(candidateSvc)
	inviteHandler := handler.NewInviteHandler(inviteSvc)

	// ── Gin router ───────────────────────────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// Structured request logging
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// ── Public routes (no authentication) ────────────────────────────────────
	v1.POST("/auth/register", authHandler.Register)
	v1.POST("/auth/login", authHandler.Login)
	v1.GET("/invites/validate/:token", inviteHandler.ValidateToken)

	// ── JWT-protected routes ──────────────────────────────────────────────────
	authed := v1.Group("")
	authed.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Job roles
		authed.POST("/jobs", jobHandler.Create)
		authed.GET("/jobs", jobHandler.List)
		authed.GET("/jobs/:job_id", jobHandler.GetByID)
		authed.PATCH("/jobs/:job_id", jobHandler.Patch)

		// Dockets (scoped to job)
		authed.POST("/jobs/:job_id/dockets", docketHandler.Create)
		authed.GET("/jobs/:job_id/dockets/:docket_id", docketHandler.GetByID)

		// Candidates
		authed.POST("/jobs/:job_id/candidates", candidateHandler.Create)
		authed.GET("/jobs/:job_id/candidates", candidateHandler.List)

		// Invites
		authed.POST("/candidates/:candidate_id/invite", inviteHandler.Create)

		// Interviews + feedback
		authed.GET("/interviews", inviteHandler.ListInterviews)
		authed.GET("/interviews/:interview_id", inviteHandler.GetInterview)
		authed.GET("/interviews/:interview_id/feedback", inviteHandler.GetFeedback)
	}

	// ── Internal routes (X-Internal-Key protected) ────────────────────────────
	// Called by Interview Orchestrator service-to-service.
	internal := v1.Group("")
	internal.Use(middleware.InternalAuthMiddleware(cfg.InternalAPIKey))
	{
		// Docket fetch — orchestrator uses at session start
		internal.GET("/dockets/:docket_id", docketHandler.GetByIDInternal)

		// Candidate fetch - feedback worker uses it to get candidate name
		internal.GET("/candidates/:candidate_id", candidateHandler.GetByIDInternal)

		// Spec gap resolutions: orchestrator needs these but spec Section 7 omits them
		internal.PATCH("/interviews/:interview_id/status", inviteHandler.UpdateStatus)
		internal.PATCH("/invites/:invite_id/use", inviteHandler.MarkInviteUsed)
	}

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

	// Block until SIGINT or SIGTERM
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
