package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/service"
)

// AuthHandler handles POST /api/v1/auth/register and POST /api/v1/auth/login.
type AuthHandler struct {
	svc      *service.AuthService
	validate *validator.Validate
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc, validate: validator.New()}
}

// Register handles POST /api/v1/auth/register.
//
//	Request:  { org_name, org_slug, full_name, email, password }
//	Response: 201 { user_id, org_id, token }
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondValidationError(c, err)
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login handles POST /api/v1/auth/login.
//
//	Request:  { email, password }
//	Response: 200 { token, user_id, org_id }
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondValidationError(c, err)
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
