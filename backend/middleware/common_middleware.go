package middleware

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// middlewareから返す共通エラーレスポンス。
type ErrRes struct {
	Error string `json:"error"`
}

// CSRF 失敗時の固定レスポンス。
type ErrMsgRes struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// 400
func writeBadRequest(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, ErrRes{
		Error: "invalid_request",
	})
}

// 401
func writeUnauthorized(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, ErrRes{
		Error: "unauthorized",
	})
}

// 403
func writeForbidden(c echo.Context) error {
	return c.JSON(http.StatusForbidden, ErrRes{
		Error: "forbidden",
	})
}

// 404
func writeNotFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, ErrRes{
		Error: "not_found",
	})
}

// 409
func writeConflict(c echo.Context) error {
	return c.JSON(http.StatusConflict, ErrRes{
		Error: "conflict",
	})
}

// 429
func writeRateLimited(c echo.Context, retryAfterSec int) error {
	if retryAfterSec > 0 {
		c.Response().Header().Set("Retry-After", strconv.Itoa(retryAfterSec))
	}

	return c.JSON(http.StatusTooManyRequests, ErrRes{
		Error: "rate_limited",
	})
}

// 500
func writeInternal(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, ErrRes{
		Error: "internal",
	})
}

// CSRF不一致
func writeCSRFMismatch(c echo.Context) error {
	return c.JSON(http.StatusForbidden, ErrMsgRes{
		Error:   "forbidden",
		Message: "csrf token mismatch",
	})
}
