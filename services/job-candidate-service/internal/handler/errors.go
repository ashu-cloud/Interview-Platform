package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// RespondError maps a service/repository error to the appropriate HTTP status code
// and the spec Section 16 error envelope: {"error": {"code": "...", "message": "..."}}.
func RespondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error: model.ErrorBody{Code: "NOT_FOUND", Message: err.Error()},
		})
	case errors.Is(err, apperror.ErrConflict):
		c.JSON(http.StatusConflict, model.ErrorResponse{
			Error: model.ErrorBody{Code: "CONFLICT", Message: err.Error()},
		})
	case errors.Is(err, apperror.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error: model.ErrorBody{Code: "UNAUTHORIZED", Message: err.Error()},
		})
	case errors.Is(err, apperror.ErrForbidden):
		c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error: model.ErrorBody{Code: "FORBIDDEN", Message: err.Error()},
		})
	case errors.Is(err, apperror.ErrBadRequest):
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorBody{Code: "BAD_REQUEST", Message: err.Error()},
		})
	default:
		slog.Error("unhandled server error", "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorBody{Code: "INTERNAL_ERROR", Message: "an unexpected error occurred"},
		})
	}
}

// RespondValidationError formats validator.ValidationErrors into the spec error envelope.
func RespondValidationError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, model.ErrorResponse{
		Error: model.ErrorBody{Code: "VALIDATION_ERROR", Message: err.Error()},
	})
}
