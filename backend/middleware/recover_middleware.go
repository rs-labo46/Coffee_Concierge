package middleware

import (
	"github.com/labstack/echo/v4"
)

// panicを拾ってinternalを返す。
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = writeInternal(c)
				}
			}()

			err = next(c)
			return err
		}
	}
}
