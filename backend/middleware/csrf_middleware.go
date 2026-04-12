package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
)

// Double Submit Cookieを検証。
// cookie csrf_token と header X-CSRF-Token が両方存在し、かつ一致する必要がある。
func CsrfCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			csrfCookie, err := c.Cookie(csrfCookieName)
			if err != nil || csrfCookie == nil {
				return writeCSRFMismatch(c)
			}

			cookieVal := strings.TrimSpace(csrfCookie.Value)
			if cookieVal == "" {
				return writeCSRFMismatch(c)
			}

			headerVal := strings.TrimSpace(c.Request().Header.Get(csrfHeaderName))
			if headerVal == "" {
				return writeCSRFMismatch(c)
			}

			if cookieVal != headerVal {
				return writeCSRFMismatch(c)
			}

			return next(c)
		}
	}
}
