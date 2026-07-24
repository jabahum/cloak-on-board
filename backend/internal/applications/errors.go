package applications

import "errors"

var (
	ErrConflict       = errors.New("application conflict")
	ErrNotFound       = errors.New("application not found")
	ErrNotProvisioned = errors.New("application is not linked to a Keycloak client")
	ErrValidation     = errors.New("validation failed")
)
