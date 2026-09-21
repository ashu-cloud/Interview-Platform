package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/middleware"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/service"
)

// InviteHandler handles invite, interview, and feedback routes.
type InviteHandler struct {
	svc      *service.InviteService
	validate *validator.Validate
}

// NewInviteHandler constructs an InviteHandler.
func NewInviteHandler(svc *service.InviteService) *InviteHandler {
	return &InviteHandler{svc: svc, validate: validator.New()}
}

// Create handles POST /api/v1/candidates/:candidate_id/invite.
//
//	Request:  { docket_id, expires_in_days }
//	Response: 201 { invite_id, invite_link, token, expires_at }
func (h *InviteHandler) Create(c *gin.Context) {
	candidateID, err := uuid.Parse(c.Param("candidate_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	var req model.CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondValidationError(c, err)
		return
	}

	orgID := middleware.GetOrgID(c)

	resp, err := h.svc.CreateInvite(c.Request.Context(), orgID, candidateID, &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ValidateToken handles GET /api/v1/invites/validate/:token (no JWT required).
//
//	Response: 200 { valid, candidate, job_role, docket_title, expires_at }
func (h *InviteHandler) ValidateToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	resp, err := h.svc.ValidateToken(c.Request.Context(), token)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListInterviews handles GET /api/v1/interviews?status=&job_id=&page=&limit=.
func (h *InviteHandler) ListInterviews(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	status := c.Query("status")
	jobID := c.Query("job_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resp, err := h.svc.ListInterviews(c.Request.Context(), orgID, status, jobID, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetInterview handles GET /api/v1/interviews/:interview_id.
func (h *InviteHandler) GetInterview(c *gin.Context) {
	interviewID, err := uuid.Parse(c.Param("interview_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	orgID := middleware.GetOrgID(c)

	detail, err := h.svc.GetInterviewByID(c.Request.Context(), orgID, interviewID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

// GetFeedback handles GET /api/v1/interviews/:interview_id/feedback.
func (h *InviteHandler) GetFeedback(c *gin.Context) {
	interviewID, err := uuid.Parse(c.Param("interview_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	orgID := middleware.GetOrgID(c)

	resp, err := h.svc.GetFeedback(c.Request.Context(), orgID, interviewID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ─── Internal endpoints (X-Internal-Key protected) ────────────────────────────

// UpdateStatus handles PATCH /api/v1/interviews/:interview_id/status.
// NOTE: Spec gap resolution — the spec requires this for the orchestrator but
// does not define the endpoint in Section 7.
//
//	Request:  { status: "in_progress|completed|feedback_ready" }
//	Response: 200 { message: "ok" }
func (h *InviteHandler) UpdateStatus(c *gin.Context) {
	interviewID, err := uuid.Parse(c.Param("interview_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	var req model.UpdateInterviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondValidationError(c, err)
		return
	}

	changed, err := h.svc.UpdateInterviewStatus(c.Request.Context(), interviewID, &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "changed": changed})
}

// MarkInviteUsed handles PATCH /api/v1/invites/:invite_id/use.
// NOTE: Spec gap resolution — needed by the orchestrator to mark an invite consumed.
//
//	Response: 200 { message: "ok" }
func (h *InviteHandler) MarkInviteUsed(c *gin.Context) {
	inviteID, err := uuid.Parse(c.Param("invite_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	if err := h.svc.MarkInviteUsed(c.Request.Context(), inviteID); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
