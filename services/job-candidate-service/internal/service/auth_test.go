package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// TestGenerateToken verifies JWT structure and claims. No DB required.
func TestGenerateToken(t *testing.T) {
	svc := &AuthService{jwtSecret: "test-secret-32-bytes-xxxxxxxxxxxxxxx"}

	userID := "user-uuid-1234"
	orgID := "org-uuid-5678"
	email := "test@example.com"

	tokenStr, err := svc.generateToken(userID, orgID, email)
	if err != nil {
		t.Fatalf("generateToken() error: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("generateToken() returned empty string")
	}

	// Parse and verify claims
	parsed, err := jwt.ParseWithClaims(tokenStr, &model.Claims{}, func(tok *jwt.Token) (interface{}, error) {
		return []byte(svc.jwtSecret), nil
	})
	if err != nil {
		t.Fatalf("ParseWithClaims() error: %v", err)
	}

	claims, ok := parsed.Claims.(*model.Claims)
	if !ok || !parsed.Valid {
		t.Fatal("token is invalid or claims type mismatch")
	}

	if claims.Subject != userID {
		t.Errorf("sub = %q, want %q", claims.Subject, userID)
	}
	if claims.OrgID != orgID {
		t.Errorf("org_id = %q, want %q", claims.OrgID, orgID)
	}
	if claims.Email != email {
		t.Errorf("email = %q, want %q", claims.Email, email)
	}

	// Verify expiry is ~24h from now
	exp := claims.ExpiresAt.Time
	if exp.Before(time.Now().Add(23 * time.Hour)) {
		t.Errorf("exp %v is less than 23h from now", exp)
	}
	if exp.After(time.Now().Add(25 * time.Hour)) {
		t.Errorf("exp %v is more than 25h from now", exp)
	}
}

func TestGenerateToken_WrongSecret(t *testing.T) {
	svc := &AuthService{jwtSecret: "correct-secret-xxxxxxxxxxxxxxxxxxxxxxxx"}
	tokenStr, _ := svc.generateToken("uid", "oid", "e@example.com")

	_, err := jwt.ParseWithClaims(tokenStr, &model.Claims{}, func(tok *jwt.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}
