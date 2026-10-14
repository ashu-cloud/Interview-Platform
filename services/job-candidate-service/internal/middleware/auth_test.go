package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

func TestCanonicalRoutesRequireJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "test-secret"

	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(AuthMiddleware(secret))
	handler := func(c *gin.Context) { c.Status(http.StatusNoContent) }

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/tracks"},
		{http.MethodGet, "/tracks"},
		{http.MethodGet, "/tracks/:job_id"},
		{http.MethodPatch, "/tracks/:job_id"},
		{http.MethodDelete, "/tracks/:job_id"},
		{http.MethodGet, "/tracks/:job_id/blueprints"},
		{http.MethodPost, "/tracks/:job_id/blueprints"},
		{http.MethodGet, "/blueprints/:docket_id"},
		{http.MethodPatch, "/blueprints/:docket_id"},
		{http.MethodDelete, "/blueprints/:docket_id"},
		{http.MethodGet, "/tracks/:job_id/participants"},
		{http.MethodPost, "/tracks/:job_id/participants"},
		{http.MethodGet, "/participants/:candidate_id"},
		{http.MethodPost, "/participants/:candidate_id/invite"},
		{http.MethodPost, "/sessions"},
		{http.MethodGet, "/sessions/:interview_id"},
		{http.MethodGet, "/sessions/:interview_id/turns"},
		{http.MethodPost, "/sessions/:interview_id/evaluation"},
	}
	for _, route := range routes {
		group.Handle(route.method, route.path, handler)
	}

	claims := model.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		OrgID: uuid.NewString(),
		Email: "manager@example.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	jobID := uuid.NewString()
	resourceID := uuid.NewString()
	for _, route := range routes {
		path := route.path
		path = replaceParam(path, ":job_id", jobID)
		path = replaceParam(path, ":docket_id", resourceID)
		path = replaceParam(path, ":candidate_id", resourceID)
		path = replaceParam(path, ":interview_id", resourceID)

		t.Run(route.method+" "+route.path+" rejects missing JWT", func(t *testing.T) {
			req := httptest.NewRequest(route.method, "/api/v1"+path, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", resp.Code, http.StatusUnauthorized)
			}
		})

		t.Run(route.method+" "+route.path+" accepts JWT", func(t *testing.T) {
			req := httptest.NewRequest(route.method, "/api/v1"+path, nil)
			req.Header.Set("Authorization", "Bearer "+signed)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			if resp.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", resp.Code, http.StatusNoContent)
			}
		})
	}
}

func replaceParam(path, param, value string) string {
	for i := 0; i+len(param) <= len(path); i++ {
		if path[i:i+len(param)] == param {
			return path[:i] + value + path[i+len(param):]
		}
	}
	return path
}
