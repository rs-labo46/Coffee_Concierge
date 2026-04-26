package controller

import (
	"coffee-spa/apperr"
	"coffee-spa/usecase"
	"errors"
	"net/http"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
)

// 入力不正
var ErrInvalidRequest = errors.New("invalid request")

// APIの失敗レスポンス
type ErrRes struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// usecase・repository・controllerのエラーをHTTPstatusとAPI・error codeに変換
func mapError(err error) (int, string, string) {
	if err == nil {
		return http.StatusInternalServerError, "internal", "internal server error"
	}

	var validationErrors validation.Errors
	var validationError validation.Error

	switch {
	case errors.Is(err, ErrInvalidRequest):
		return http.StatusBadRequest, "invalid_request", "request body or query is invalid"

	case errors.As(err, &validationErrors):
		return http.StatusBadRequest, "invalid_request", "request value is invalid"

	case errors.As(err, &validationError):
		return http.StatusBadRequest, "invalid_request", "request value is invalid"

	case errors.Is(err, usecase.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized", "authentication failed"

	case errors.Is(err, usecase.ErrForbidden):
		return http.StatusForbidden, "forbidden", "permission denied"

	case errors.Is(err, usecase.ErrNotFound):
		return http.StatusNotFound, "not_found", "resource not found"

	case errors.Is(err, usecase.ErrConflict):
		return http.StatusConflict, "conflict", "resource conflict"

	case errors.Is(err, usecase.ErrInvalidState):
		return http.StatusConflict, "conflict", "resource state is invalid"

	case errors.Is(err, usecase.ErrRateLimited):
		return http.StatusTooManyRequests, "rate_limited", "too many requests"

	default:
		return http.StatusInternalServerError, "internal", "internal server error"
	}
}

// controllerから共通で使う失敗レスポンス関数
func writeErr(c echo.Context, err error) error {
	status, code, msg := mapError(err)
	var rlErr apperr.RateLimitedError
	if errors.As(err, &rlErr) && rlErr.RetryAfterSec > 0 {
		c.Response().Header().Set("Retry-After", strconv.Itoa(rlErr.RetryAfterSec))
	}

	return c.JSON(status, ErrRes{
		Error:   code,
		Message: msg,
	})
}

// bind失敗やparam/query不正のときにErrInvalidRequestを毎回呼ばないようにする。
func writeInvalidRequest(c echo.Context) error {
	return writeErr(c, ErrInvalidRequest)
}
