package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/handler"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/middleware"
)

type routeHandlers struct {
	auth      *handler.AuthHandler
	job       *handler.JobHandler
	docket    *handler.DocketHandler
	candidate *handler.CandidateHandler
	invite    *handler.InviteHandler
}

func newRouter(handlers routeHandlers, jwtSecret, internalAPIKey string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	// CORS middleware for web clients (e.g. Cloudflare Pages)
	r.Use(func(c *gin.Context) {
		origin := os.Getenv("APP_DOMAIN")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Internal-Key")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

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

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "platform-service"})
	})
	registerRoutes(r, handlers, jwtSecret, internalAPIKey)
	return r
}

func registerRoutes(r *gin.Engine, handlers routeHandlers, jwtSecret, internalAPIKey string) {
	v1 := r.Group("/api/v1")

	v1.POST("/auth/register", handlers.auth.Register)
	v1.POST("/auth/login", handlers.auth.Login)
	// Deprecated frontend compatibility route. Remove when frontend is updated.
	v1.GET("/invites/validate/:token", handlers.invite.ValidateToken)

	authed := v1.Group("")
	authed.Use(middleware.AuthMiddleware(jwtSecret))
	{
		authed.POST("/tracks", handlers.job.Create)
		authed.GET("/tracks", handlers.job.ListTracks)
		authed.GET("/tracks/:job_id", handlers.job.GetByID)
		authed.PATCH("/tracks/:job_id", handlers.job.Patch)
		authed.DELETE("/tracks/:job_id", handlers.job.Delete)

		authed.GET("/tracks/:job_id/blueprints", handlers.docket.List)
		authed.POST("/tracks/:job_id/blueprints", handlers.docket.Create)
		authed.GET("/blueprints/:docket_id", handlers.docket.GetByID)
		authed.PATCH("/blueprints/:docket_id", handlers.docket.Update)
		authed.DELETE("/blueprints/:docket_id", handlers.docket.Delete)

		authed.GET("/tracks/:job_id/participants", handlers.candidate.ListParticipants)
		authed.POST("/tracks/:job_id/participants", handlers.candidate.Create)
		authed.GET("/participants/:candidate_id", handlers.candidate.GetByID)
		authed.POST("/participants/:candidate_id/invite", handlers.invite.Create)

		authed.POST("/sessions", handlers.invite.CreateSession)
		authed.GET("/sessions/:interview_id", handlers.invite.GetSession)
		authed.GET("/sessions/:interview_id/turns", handlers.invite.ListSessionTurns)
		authed.POST("/sessions/:interview_id/evaluation", handlers.invite.EnqueueEvaluation)

		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.POST("/jobs", handlers.job.Create)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.GET("/jobs", handlers.job.List)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.GET("/jobs/:job_id", handlers.job.GetByID)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.PATCH("/jobs/:job_id", handlers.job.Patch)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.GET("/jobs/:job_id/dockets", handlers.docket.ListCompatibility)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.POST("/jobs/:job_id/dockets", handlers.docket.Create)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.GET("/jobs/:job_id/dockets/:docket_id", handlers.docket.GetByID)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.POST("/jobs/:job_id/candidates", handlers.candidate.Create)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.GET("/jobs/:job_id/candidates", handlers.candidate.List)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.POST("/candidates/:candidate_id/invite", handlers.invite.Create)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.GET("/interviews", handlers.invite.ListInterviews)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.GET("/interviews/:interview_id", handlers.invite.GetInterview)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		authed.GET("/interviews/:interview_id/feedback", handlers.invite.GetFeedback)
	}

	internal := v1.Group("")
	internal.Use(middleware.InternalAuthMiddleware(internalAPIKey))
	{
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		internal.GET("/dockets/:docket_id", handlers.docket.GetByIDInternal)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		internal.GET("/candidates/:candidate_id", handlers.candidate.GetByIDInternal)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		internal.PATCH("/interviews/:interview_id/status", handlers.invite.UpdateStatus)
		// Deprecated frontend compatibility route. Remove when frontend is updated.
		internal.PATCH("/invites/:invite_id/use", handlers.invite.MarkInviteUsed)
	}

	// Service-to-service routes live under /internal and are not exposed via /api/v1.
	svc := r.Group("/internal")
	svc.Use(middleware.InternalAuthMiddleware(internalAPIKey))
	svc.GET("/blueprints/:blueprint_id/criteria-embeddings", handlers.docket.GetCriteriaEmbeddings)
}
