package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type loginLimiterForAdditionalTest struct {
	ipAllowed    bool
	emailAllowed bool
	retryAfter   int
}

func (m loginLimiterForAdditionalTest) AllowLoginIP(ip string) (bool, int, error) {
	return m.ipAllowed, m.retryAfter, nil
}
func (m loginLimiterForAdditionalTest) AllowLogin(emailHash string) (bool, int, error) {
	return m.emailAllowed, m.retryAfter, nil
}

func TestLoginRateLimit_AdditionalCoverage(t *testing.T) {
	// M-RL-ADD-01
	t.Run("IP deny returns 429", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := LoginRateLimit(loginLimiterForAdditionalTest{ipAllowed: false, emailAllowed: true, retryAfter: 9})(func(c echo.Context) error {
			t.Fatal("next should not be called")
			return nil
		})
		if err := h(c); err != nil {
			t.Fatalf("handler error = %v", err)
		}
		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
		}
		if got := rec.Header().Get("Retry-After"); got != "9" {
			t.Fatalf("Retry-After = %q, want 9", got)
		}
	})

	// M-RL-ADD-02
	t.Run("email deny returns 429 after body restore", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"email":"user@example.com"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := LoginRateLimit(loginLimiterForAdditionalTest{ipAllowed: true, emailAllowed: false, retryAfter: 11})(func(c echo.Context) error {
			t.Fatal("next should not be called")
			return nil
		})
		if err := h(c); err != nil {
			t.Fatalf("handler error = %v", err)
		}
		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
		}
		if got := rec.Header().Get("Retry-After"); got != "11" {
			t.Fatalf("Retry-After = %q, want 11", got)
		}
	})
}
