package middleware

import (
	"os"

	"github.com/labstack/echo/v4"
	emw "github.com/labstack/echo/v4/middleware"
)

func Logger() echo.MiddlewareFunc {
	return emw.LoggerWithConfig(emw.LoggerConfig{
		Output: os.Stdout,
		Format: `{"time":"${time_rfc3339}","request_id":"${header:X-Request-Id}","remote_ip":"${remote_ip}","host":"${host}","method":"${method}","uri":"${uri}","status":${status},"latency":"${latency_human}","bytes_in":${bytes_in},"bytes_out":${bytes_out},"user_agent":"${user_agent}"}` + "\n",
	})
}
