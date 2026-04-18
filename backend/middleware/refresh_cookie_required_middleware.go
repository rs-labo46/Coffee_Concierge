package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type refreshErrRes struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// 実際のratelimit判定はusecase側で行う。
func RequireRefreshCookie() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("refresh_token")
			if err != nil || cookie == nil {
				return c.JSON(http.StatusUnauthorized, refreshErrRes{
					Error:   "unauthorized",
					Message: "refresh token is missing",
				})
			}

			if strings.TrimSpace(cookie.Value) == "" {
				return c.JSON(http.StatusUnauthorized, refreshErrRes{
					Error:   "unauthorized",
					Message: "refresh token is missing",
				})
			}

			return next(c)
		}
	}
}
