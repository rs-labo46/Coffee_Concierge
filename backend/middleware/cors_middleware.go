package middleware

import (
	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
)

// CORSはSPA用のCORSを設定。
// Allow-Originは呼び出し側から固定値で渡す。
func CORS(feURL string) echo.MiddlewareFunc {
	return emw.CORSWithConfig(emw.CORSConfig{
		AllowOrigins:     []string{feURL},
		AllowCredentials: true,
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderAuthorization,
			echo.HeaderContentType,
			"X-CSRF-Token",
		},
	})
}
