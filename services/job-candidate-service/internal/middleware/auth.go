package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// Context keys for values set by AuthMiddleware.
const (
	ContextKeyUserID = "user_id"
	ContextKeyOrgID  = "org_id"
	ContextKeyEmail  = "email"
)

// AuthMiddleware validates a Bearer JWT and injects user_id, org_id, email into the Gin context.
// Returns 401 UNAUTHORIZED on any validation failure.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			respondUnauthorized(c, "missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respondUnauthorized(c, "invalid Authorization header format")
			return
		}
		tokenStr := parts[1]

		claims := &model.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			slog.Debug("AuthMiddleware: invalid token", "error", err)
			respondUnauthorized(c, "invalid or expired token")
			return
		}

		c.Set(ContextKeyUserID, claims.Subject)
		c.Set(ContextKeyOrgID, claims.OrgID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Next()
	}
}

// InternalAuthMiddleware validates the X-Internal-Key header for service-to-service calls.
// Returns 401 UNAUTHORIZED if the key is missing or incorrect.
// Uses constant-time comparison to prevent timing attacks.
func InternalAuthMiddleware(internalAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Internal-Key")
		if internalAPIKey == "" || key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(internalAPIKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
				Error: model.ErrorBody{Code: "UNAUTHORIZED", Message: "invalid or missing X-Internal-Key"},
			})
			return
		}
		c.Next()
	}
}

// GetOrgID extracts org_id from the Gin context (set by AuthMiddleware).
// Returns uuid.Nil if not present.
func GetOrgID(c *gin.Context) uuid.UUID {
	val, _ := c.Get(ContextKeyOrgID)
	s, _ := val.(string)
	id, _ := uuid.Parse(s)
	return id
}

// GetUserID extracts user_id from the Gin context (set by AuthMiddleware).
// Returns uuid.Nil if not present.
func GetUserID(c *gin.Context) uuid.UUID {
	val, _ := c.Get(ContextKeyUserID)
	s, _ := val.(string)
	id, _ := uuid.Parse(s)
	return id
}

func respondUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
		Error: model.ErrorBody{Code: "UNAUTHORIZED", Message: message},
	})
}
