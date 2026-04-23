package usecase

import "coffee-spa/apperr"

var (
	ErrNotFound     = apperr.ErrNotFound
	ErrConflict     = apperr.ErrConflict
	ErrForbidden    = apperr.ErrForbidden
	ErrUnauthorized = apperr.ErrUnauthorized
	ErrInvalidState = apperr.ErrInvalidState
	ErrRateLimited  = apperr.ErrRateLimited
	ErrInternal     = apperr.ErrInternal
)
