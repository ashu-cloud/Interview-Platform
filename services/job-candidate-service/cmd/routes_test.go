package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/handler"
)

func TestRegisteredProtectedRoutesRejectMissingCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newRouter(routeHandlers{
		auth:      handler.NewAuthHandler(nil),
		job:       handler.NewJobHandler(nil),
		docket:    handler.NewDocketHandler(nil),
		candidate: handler.NewCandidateHandler(nil),
		invite:    handler.NewInviteHandler(nil),
	}, "jwt-secret", "internal-secret")

	jwtRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/tracks"},
		{http.MethodGet, "/api/v1/tracks"},
		{http.MethodGet, "/api/v1/tracks/job-id"},
		{http.MethodPatch, "/api/v1/tracks/job-id"},
		{http.MethodDelete, "/api/v1/tracks/job-id"},
		{http.MethodGet, "/api/v1/tracks/job-id/blueprints"},
		{http.MethodPost, "/api/v1/tracks/job-id/blueprints"},
		{http.MethodGet, "/api/v1/blueprints/docket-id"},
		{http.MethodPatch, "/api/v1/blueprints/docket-id"},
		{http.MethodDelete, "/api/v1/blueprints/docket-id"},
		{http.MethodGet, "/api/v1/tracks/job-id/participants"},
		{http.MethodPost, "/api/v1/tracks/job-id/participants"},
		{http.MethodGet, "/api/v1/participants/candidate-id"},
		{http.MethodPost, "/api/v1/participants/candidate-id/invite"},
		{http.MethodPost, "/api/v1/sessions"},
		{http.MethodGet, "/api/v1/sessions/interview-id"},
		{http.MethodGet, "/api/v1/sessions/interview-id/turns"},
		{http.MethodPost, "/api/v1/sessions/interview-id/evaluation"},
		{http.MethodGet, "/api/v1/jobs/job-id/dockets"},
	}
	for _, route := range jwtRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)
			if resp.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", resp.Code, http.StatusUnauthorized)
			}
		})
	}

	internalRoutes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/dockets/docket-id"},
		{http.MethodGet, "/api/v1/candidates/candidate-id"},
		{http.MethodPatch, "/api/v1/interviews/interview-id/status"},
		{http.MethodPatch, "/api/v1/invites/invite-id/use"},
		{http.MethodGet, "/internal/blueprints/blueprint-id/criteria-embeddings"},
	}
	for _, route := range internalRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)
			if resp.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", resp.Code, http.StatusUnauthorized)
			}
		})
	}
}
