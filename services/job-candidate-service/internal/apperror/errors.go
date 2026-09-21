package apperror

import "errors"

// Sentinel errors used to translate service/repository errors into HTTP responses.
// Handlers use errors.Is to map these to the appropriate HTTP status codes.
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBadRequest   = errors.New("bad request")
)
