package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// WebSocket upgradeの要求だけを通す。
func WsUpgradeCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			upgrade := strings.ToLower(strings.TrimSpace(c.Request().Header.Get("Upgrade")))
			if upgrade != "websocket" {
				return writeBadRequest(c)
			}

			connHdr := strings.ToLower(c.Request().Header.Get("Connection"))
			if !strings.Contains(connHdr, "upgrade") {
				return writeBadRequest(c)
			}

			return next(c)
		}
	}
}
