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

// JobHandler handles /api/v1/jobs routes.
type JobHandler struct {
	svc      *service.JobService
	validate *validator.Validate
}

// NewJobHandler constructs a JobHandler.
func NewJobHandler(svc *service.JobService) *JobHandler {
	return &JobHandler{svc: svc, validate: validator.New()}
}

// Create handles POST /api/v1/jobs.
//
//	Request:  { title, description, required_skills }
//	Response: 201 JobRole
func (h *JobHandler) Create(c *gin.Context) {
	var req model.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondValidationError(c, err)
		return
	}

	orgID := middleware.GetOrgID(c)
	userID := middleware.GetUserID(c)

	job, err := h.svc.CreateJob(c.Request.Context(), orgID, userID, &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, job)
}

// List handles GET /api/v1/jobs?status=&page=&limit=.
func (h *JobHandler) List(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resp, err := h.svc.ListJobs(c.Request.Context(), orgID, status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListTracks returns the canonical tracks list envelope.
func (h *JobHandler) ListTracks(c *gin.Context) {
	orgID := middleware.GetOrgID(c)
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resp, err := h.svc.ListJobs(c.Request.Context(), orgID, status, page, limit)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"tracks": resp.Jobs,
		"total":  resp.Total,
		"page":   resp.Page,
		"limit":  resp.Limit,
	})
}

// GetByID handles GET /api/v1/jobs/:job_id.
func (h *JobHandler) GetByID(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}
	orgID := middleware.GetOrgID(c)

	job, err := h.svc.GetJob(c.Request.Context(), orgID, jobID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, job)
}

// Patch handles PATCH /api/v1/jobs/:job_id.
//
//	Request:  any subset of { title, description, required_skills, status }
//	Response: 200 updated JobRole
func (h *JobHandler) Patch(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	var req model.PatchJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondValidationError(c, err)
		return
	}

	orgID := middleware.GetOrgID(c)

	job, err := h.svc.PatchJob(c.Request.Context(), orgID, jobID, &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, job)
}

// Delete handles DELETE of a canonical track.
func (h *JobHandler) Delete(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}
	if err := h.svc.DeleteJob(c.Request.Context(), middleware.GetOrgID(c), jobID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
