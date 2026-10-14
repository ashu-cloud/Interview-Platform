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

// DocketHandler handles docket-related routes.
type DocketHandler struct {
	svc      *service.DocketService
	validate *validator.Validate
}

// NewDocketHandler constructs a DocketHandler.
func NewDocketHandler(svc *service.DocketService) *DocketHandler {
	return &DocketHandler{svc: svc, validate: validator.New()}
}

// Create handles POST /api/v1/jobs/:job_id/dockets.
//
//	Request:  { title, instructions, duration_mins, questions[], rubric[] }
//	Response: 201 Docket (with nested questions and rubric)
func (h *DocketHandler) Create(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	var req model.CreateDocketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondValidationError(c, err)
		return
	}

	orgID := middleware.GetOrgID(c)

	docket, err := h.svc.CreateDocket(c.Request.Context(), orgID, jobID, &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, docket)
}

// List handles the canonical blueprint list route.
func (h *DocketHandler) List(c *gin.Context) {
	h.list(c, "blueprints")
}

// ListCompatibility preserves the legacy dockets list envelope.
func (h *DocketHandler) ListCompatibility(c *gin.Context) {
	h.list(c, "dockets")
}

func (h *DocketHandler) list(c *gin.Context, envelope string) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}
	dockets, err := h.svc.ListDockets(c.Request.Context(), middleware.GetOrgID(c), jobID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{envelope: dockets})
}

// Update handles a canonical partial blueprint update.
func (h *DocketHandler) Update(c *gin.Context) {
	docketID, err := uuid.Parse(c.Param("docket_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}
	var req model.UpdateDocketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		RespondValidationError(c, err)
		return
	}
	docket, err := h.svc.UpdateDocket(c.Request.Context(), middleware.GetOrgID(c), docketID, &req)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, docket)
}

// Delete handles canonical blueprint deletion.
func (h *DocketHandler) Delete(c *gin.Context) {
	docketID, err := uuid.Parse(c.Param("docket_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}
	if err := h.svc.DeleteDocket(c.Request.Context(), middleware.GetOrgID(c), docketID); err != nil {
		RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetByID handles GET /api/v1/jobs/:job_id/dockets/:docket_id (JWT-protected).
func (h *DocketHandler) GetByID(c *gin.Context) {
	docketID, err := uuid.Parse(c.Param("docket_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	orgID := middleware.GetOrgID(c)

	docket, err := h.svc.GetDocket(c.Request.Context(), orgID, docketID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, docket)
}

// GetByIDInternal handles GET /api/v1/dockets/:docket_id (X-Internal-Key protected).
// This is the endpoint called by the Interview Orchestrator at session start.
func (h *DocketHandler) GetByIDInternal(c *gin.Context) {
	docketID, err := uuid.Parse(c.Param("docket_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}

	docket, err := h.svc.GetDocketInternal(c.Request.Context(), docketID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, docket)
}

// GetCriteriaEmbeddings handles GET /internal/blueprints/:blueprint_id/criteria-embeddings
// (X-Internal-Key protected). It returns {"embeddings": {criterion_id: [floats]}};
// the map is empty when the blueprint predates embedding generation.
func (h *DocketHandler) GetCriteriaEmbeddings(c *gin.Context) {
	blueprintID, err := uuid.Parse(c.Param("blueprint_id"))
	if err != nil {
		RespondError(c, apperror.ErrBadRequest)
		return
	}
	embeddings, err := h.svc.GetCriteriaEmbeddings(c.Request.Context(), blueprintID)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"embeddings": embeddings})
}
