package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	HeaderSessionKey = "X-Session-Key"
	QuerySessionKey  = "session_key"
)

// HTTP guestでX-Session-Keyを要求。
func SessionKeyHeaderRequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := strings.TrimSpace(c.Request().Header.Get(HeaderSessionKey))
			if key == "" {
				return writeBadRequest(c)
			}

			c.Set("session_key", key)
			return next(c)
		}
	}
}

// WSのguestでqueryのsession_keyを要求。
func SessionKeyQueryRequired() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := strings.TrimSpace(c.QueryParam(QuerySessionKey))
			if key == "" {
				return writeBadRequest(c)
			}

			c.Set("session_key", key)
			return next(c)
		}
	}
}
