package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/middleware"
	"coffee-spa/testutil/middlewaremock"

	"github.com/labstack/echo/v4"
)

func runMW(t *testing.T, mw echo.MiddlewareFunc, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	next := func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]string{"ok": "true"}) }
	if err := mw(next)(c); err != nil {
		t.Fatalf("middleware error: %v", err)
	}
	return rec
}

func TestRefreshAndSecurityMiddleware_RequiredCoverage(t *testing.T) {
	t.Run("RequireRefreshCookie rejects missing cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		rec := runMW(t, middleware.RequireRefreshCookie(), req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("RequireRefreshCookie rejects blank cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "   "})
		rec := runMW(t, middleware.RequireRefreshCookie(), req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("RequireRefreshCookie passes present cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh"})
		rec := runMW(t, middleware.RequireRefreshCookie(), req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("RefreshRateLimit rejects missing cookie", func(t *testing.T) {
		rl := &middlewaremock.RateLimiter{Allowed: true}
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		rec := runMW(t, middleware.RefreshRateLimit(rl, "refresh_token"), req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("RefreshRateLimit returns 429 with Retry-After", func(t *testing.T) {
		rl := &middlewaremock.RateLimiter{Allowed: false, RetryAfter: 42}
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh"})
		rec := runMW(t, middleware.RefreshRateLimit(rl, "refresh_token"), req)
		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("status=%d", rec.Code)
		}
		if got := rec.Header().Get("Retry-After"); got != "42" {
			t.Fatalf("Retry-After=%q", got)
		}
	})

	t.Run("ForgotPwRateLimit rejects malformed json", func(t *testing.T) {
		rl := &middlewaremock.RateLimiter{Allowed: true}
		req := httptest.NewRequest(http.MethodPost, "/auth/password/forgot", strings.NewReader(`{bad`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := runMW(t, middleware.ForgotPwRateLimit(rl), req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("ResendVerifyRateLimit store error returns 500", func(t *testing.T) {
		rl := &middlewaremock.RateLimiter{Err: errors.New("redis down")}
		req := httptest.NewRequest(http.MethodPost, "/auth/verify-email/resend", strings.NewReader(`{"email":"a@example.com"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := runMW(t, middleware.ResendVerifyRateLimit(rl), req)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("WsRateLimit uses user key when actor exists", func(t *testing.T) {
		rl := &middlewaremock.RateLimiter{Allowed: true}
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/ws/search/sessions/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("actor", &entity.Actor{UserID: 1})
		next := func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }
		if err := middleware.WsRateLimit(rl)(next)(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if !strings.HasPrefix(rl.LastKey, "rl:ws:user:") {
			t.Fatalf("key=%q", rl.LastKey)
		}
	})
}
