package apperr

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidState = errors.New("invalid state")
	ErrRateLimited  = errors.New("rate limited")
	ErrInternal     = errors.New("internal")
)
