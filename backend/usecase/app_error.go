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

// 429用の詳細付きエラー。
type RateLimitedError = apperr.RateLimitedError
