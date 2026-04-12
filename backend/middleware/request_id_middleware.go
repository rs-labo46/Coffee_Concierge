package middleware

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
)

func RequestID() echo.MiddlewareFunc {
	return emw.RequestIDWithConfig(emw.RequestIDConfig{
		TargetHeader: echo.HeaderXRequestID,
		Generator: func() string {
			return uuid.NewString()
		},
		RequestIDHandler: func(c echo.Context, id string) {
			id = strings.TrimSpace(id)
			if id == "" {
				return
			}
			c.Set("request_id", id)
		},
	})
}
