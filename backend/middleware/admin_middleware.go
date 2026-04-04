package middleware

import "github.com/labstack/echo/v4"

// role=adminのみ通す。
// roleの不一致は403。
func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, _ := c.Get("role").(string)
			if role != "admin" {
				return writeForbidden(c)
			}

			return next(c)
		}
	}
}
