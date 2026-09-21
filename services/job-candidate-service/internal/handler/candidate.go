package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/middleware"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/service"
)

// CandidateHandler handles candidate-related routes.
type CandidateHandler struct {
	svc      *service.CandidateService
	validate *validator.Validate
}

// NewCandidateHandler constructs a CandidateHandler.
func NewCandidateHandler(svc *service.CandidateService) *CandidateHandler {
	return &CandidateHandler{svc: svc, validate: validator.New()}
}

// Create handles POST /api/v1/jobs/:job_id/candidates.
//
//	Request:  { email, full_name }
//	Response: 201 Candidate
func (h *CandidateHandler) Create(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	var req model.CreateCandidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondValidationError(c, err)
		return
	}

	orgID := middleware.GetOrgID(c)

	candidate, err := h.svc.CreateCandidate(c.Request.Context(), orgID, jobID, &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, candidate)
}

// List handles GET /api/v1/jobs/:job_id/candidates.
//
//	Response: 200 []CandidateWithStatus
func (h *CandidateHandler) List(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	orgID := middleware.GetOrgID(c)

	candidates, err := h.svc.ListCandidates(c.Request.Context(), orgID, jobID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"candidates": candidates})
}

// GetByIDInternal handles GET /api/v1/internal/candidates/:candidate_id.
//
//	Response: 200 Candidate
func (h *CandidateHandler) GetByIDInternal(c *gin.Context) {
	candidateID, err := uuid.Parse(c.Param("candidate_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	candidate, err := h.svc.GetByIDInternal(c.Request.Context(), candidateID)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, candidate)
}
