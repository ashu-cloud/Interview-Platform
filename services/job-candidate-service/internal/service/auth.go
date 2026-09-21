package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/password"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/repository"
)

// AuthService handles registration, login, and JWT issuance.
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

// NewAuthService constructs an AuthService.
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: jwtSecret}
}

// Register creates a new organization and its first user, returning a JWT.
// Returns ErrConflict if the org slug or email is already taken.
func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.RegisterResponse, error) {
	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Register: hash password: %w", err)
	}

	org, user, err := s.userRepo.CreateOrgAndUser(ctx,
		req.OrgName, req.OrgSlug, req.Email, hash, req.FullName,
	)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Register: %w", err)
	}

	token, err := s.generateToken(user.ID.String(), org.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Register: generate token: %w", err)
	}

	return &model.RegisterResponse{
		UserID: user.ID.String(),
		OrgID:  org.ID.String(),
		Token:  token,
	}, nil
}

// Login authenticates a user by email and password, returning a JWT.
// Returns ErrUnauthorized on credential mismatch.
func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Mask not-found as unauthorized to avoid user enumeration
		return nil, fmt.Errorf("AuthService.Login: %w", apperror.ErrUnauthorized)
	}

	if !password.Verify(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("AuthService.Login: %w", apperror.ErrUnauthorized)
	}

	token, err := s.generateToken(user.ID.String(), user.OrgID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Login: generate token: %w", err)
	}

	return &model.LoginResponse{
		Token:  token,
		UserID: user.ID.String(),
		OrgID:  user.OrgID.String(),
	}, nil
}

// generateToken creates a signed JWT with 24-hour expiry.
// Claims: sub (user_id), org_id, email, exp.
func (s *AuthService) generateToken(userID, orgID, email string) (string, error) {
	claims := model.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
		OrgID: orgID,
		Email: email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("generateToken: sign: %w", err)
	}
	return signed, nil
}
