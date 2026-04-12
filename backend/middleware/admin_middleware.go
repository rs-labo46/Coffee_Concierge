package middleware

import (
	"coffee-spa/entity"

	"github.com/labstack/echo/v4"
)

// actor.role=adminのみ通す。
func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			v := c.Get("actor")
			if v == nil {
				return writeForbidden(c)
			}

			actor, ok := v.(*entity.Actor)
			if !ok || actor == nil {
				return writeForbidden(c)
			}

			if actor.Role != entity.RoleAdmin {
				return writeForbidden(c)
			}
			return next(c)
		}
	}
}
