package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// middlewareから返す共通エラーレスポンス。
type ErrRes struct {
	Error string `json:"error"`
}

// CSRF 失敗時の固定レスポンス。
type CsrfErrRes struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// writeUnauthorizedは401。
func writeUnauthorized(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, ErrRes{
		Error: "unauthorized",
	})
}

// writeForbiddenは403。
func writeForbidden(c echo.Context) error {
	return c.JSON(http.StatusForbidden, ErrRes{
		Error: "forbidden",
	})
}

// writeCSRFMismatchはCSRFエラー。
func writeCSRFMismatch(c echo.Context) error {
	return c.JSON(http.StatusForbidden, CsrfErrRes{
		Error:   "forbidden",
		Message: "csrf token mismatch",
	})
}
