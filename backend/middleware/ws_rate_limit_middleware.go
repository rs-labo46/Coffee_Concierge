package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// middlewareは業務ロジックを持たず、許可/拒否だけをusecaseに委譲する。
type WsRateLimiter interface {
	AllowWS(key string) (bool, int, error)
}

// WebSocket用の接続制限で、HTTP APIの制限とは分けて、WS接続専用のkeyで判定する。
func WsRateLimit(rl WsRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl == nil {
				return next(c)
			}

			key := wsRateLimitKey(c)

			allowed, retryAfterSec, err := rl.AllowWS(key)
			if err != nil {
				return writeInternal(c)
			}
			if !allowed {
				return writeRateLimited(c, retryAfterSec)
			}

			return next(c)
		}
	}
}

// WS 接続制限用のkeyを組みたて、private/guestとIPを使い分ける。
func wsRateLimitKey(c echo.Context) string {
	ip := strings.TrimSpace(c.RealIP())
	if ip == "" {
		ip = "unknown"
	}

	// JWTAuth済みならuser側key、それ以外はguest側keyに分ける。
	if c.Get("actor") != nil {
		return "rl:ws:user:" + ip
	}

	return "rl:ws:guest:" + ip
}
