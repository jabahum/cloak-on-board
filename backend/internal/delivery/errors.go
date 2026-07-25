package delivery

import "errors"

var (
	ErrNotFound    = errors.New("resource not found")
	ErrConflict    = errors.New("invalid or stale workflow state")
	ErrValidation  = errors.New("validation failed")
	ErrForbidden   = errors.New("operation requires approval")
	ErrUnavailable = errors.New("target realm unavailable")
	ErrGone        = errors.New("secret delivery expired or consumed")
)
