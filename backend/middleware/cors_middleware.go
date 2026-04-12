package middleware

import (
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
)

// CORSを設定。
// Allow-Originは呼び出し側から固定値で渡す。
func CORS(feURL string) echo.MiddlewareFunc {
	return emw.CORSWithConfig(emw.CORSConfig{
		AllowOrigins:     []string{feURL},
		AllowCredentials: true,
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PATCH,
			echo.PUT,
			echo.DELETE,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderAuthorization,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderOrigin,
			echo.HeaderXRequestedWith,
			echo.HeaderXRequestID,
			"X-CSRF-Token",
			"X-Session-Key",
		},
		ExposeHeaders: []string{
			echo.HeaderXRequestID,
		},
	})
}
