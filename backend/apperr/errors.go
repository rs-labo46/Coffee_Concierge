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

type RateLimitedError struct {
	RetryAfterSec int
}

func (e RateLimitedError) Error() string {
	return ErrRateLimited.Error()
}

// Unwrapにより、errors.Is(err, ErrRateLimited)をtrueにできる。
func (e RateLimitedError) Unwrap() error {
	return ErrRateLimited
}
