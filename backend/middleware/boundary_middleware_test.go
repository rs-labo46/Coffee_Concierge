package middleware

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCsrfCheck_MatchedCookieAndHeader_CallsNext(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/auth/refresh", "")
	c.Request().AddCookie(&http.Cookie{Name: csrfCookieName, Value: "token"})
	c.Request().Header.Set(csrfHeaderName, "token")
	err := CsrfCheck()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
}

func TestCsrfCheck_Mismatch_ReturnsForbidden(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/auth/refresh", "")
	c.Request().AddCookie(&http.Cookie{Name: csrfCookieName, Value: "cookie-token"})
	c.Request().Header.Set(csrfHeaderName, "header-token")
	err := CsrfCheck()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusForbidden)
}

func TestRequireRefreshCookie_Present_CallsNext(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/auth/refresh", "")
	c.Request().AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh"})
	err := RequireRefreshCookie()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
}

func TestRequireRefreshCookie_Missing_ReturnsUnauthorized(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/auth/refresh", "")
	err := RequireRefreshCookie()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestSessionKeyHeaderRequired_Present_SetsContext(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/search/guest/sessions/1", "")
	c.Request().Header.Set(HeaderSessionKey, "guest-key")
	err := SessionKeyHeaderRequired()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
	if got := c.Get("session_key"); got != "guest-key" {
		t.Fatalf("session_key mismatch: %#v", got)
	}
}

func TestSessionKeyHeaderRequired_Missing_ReturnsBadRequest(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/search/guest/sessions/1", "")
	err := SessionKeyHeaderRequired()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusBadRequest)
}

func TestWsUpgradeCheck_ValidHeaders_CallsNext(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/ws/search/sessions/1", "")
	c.Request().Header.Set("Upgrade", "websocket")
	c.Request().Header.Set("Connection", "keep-alive, Upgrade")
	err := WsUpgradeCheck()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
}

func TestWsUpgradeCheck_MissingHeaders_ReturnsBadRequest(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/ws/search/sessions/1", "")
	err := WsUpgradeCheck()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusBadRequest)
}

func TestSecurityHeaders_SetsExpectedHeaders(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/health", "")
	err := SecurityHeaders()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("security header was not set")
	}
}

func TestRecover_Panic_ReturnsInternal(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/panic", "")
	err := Recover()(func(c echo.Context) error {
		panic("boom")
	})(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusInternalServerError)
}
