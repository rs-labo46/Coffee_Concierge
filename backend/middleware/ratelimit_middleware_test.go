package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coffee-spa/testutil/middlewaremock"

	"github.com/labstack/echo/v4"
)

func TestSignupRateLimit_AllowsRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	rl := &middlewaremock.RateLimiter{Allowed: true}
	called := false

	h := SignupRateLimit(rl)(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})
	if err := h(c); err != nil {
		t.Fatalf("handler error = %v", err)
	}
	if !called || rec.Code != http.StatusNoContent {
		t.Fatalf("called/code = %v/%d", called, rec.Code)
	}
}

func TestSignupRateLimit_ReturnsTooManyRequests(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	rl := &middlewaremock.RateLimiter{Allowed: false, RetryAfter: 30}

	h := SignupRateLimit(rl)(func(c echo.Context) error {
		t.Fatal("next should not be called")
		return nil
	})
	if err := h(c); err != nil {
		t.Fatalf("handler error = %v", err)
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if got := rec.Header().Get("Retry-After"); got != "30" {
		t.Fatalf("Retry-After = %q, want 30", got)
	}
}
