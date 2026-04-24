package middleware

import (
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSignupRateLimit_Allowed_CallsNextWithIP(t *testing.T) {
	rl := &allRateLimiterMock{allow: true}
	c, rec := newTestContext(http.MethodPost, "/auth/signup", `{"email":"a@example.com"}`)
	c.Request().RemoteAddr = "192.0.2.1:1234"
	err := SignupRateLimit(rl)(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
	if len(rl.signupIPs) != 1 || rl.signupIPs[0] == "" {
		t.Fatalf("AllowSignup was not called correctly: %#v", rl.signupIPs)
	}
}

func TestSignupRateLimit_Denied_ReturnsTooManyRequests(t *testing.T) {
	rl := &allRateLimiterMock{allow: false, retry: 9}
	c, rec := newTestContext(http.MethodPost, "/auth/signup", `{}`)
	err := SignupRateLimit(rl)(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusTooManyRequests)
	if rec.Header().Get("Retry-After") != "9" {
		t.Fatalf("Retry-After mismatch: %s", rec.Header().Get("Retry-After"))
	}
}

func TestSignupRateLimit_StoreError_ReturnsInternal(t *testing.T) {
	rl := &allRateLimiterMock{allow: false, err: errors.New("redis down")}
	c, rec := newTestContext(http.MethodPost, "/auth/signup", `{}`)
	err := SignupRateLimit(rl)(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusInternalServerError)
}

func TestLoginRateLimit_Allowed_RestoresBodyAndHashesEmail(t *testing.T) {
	rl := &allRateLimiterMock{allow: true}
	c, rec := newTestContext(http.MethodPost, "/auth/login", `{"email":" User@Test.COM ","password":"password"}`)
	err := LoginRateLimit(rl)(func(c echo.Context) error {
		email, err := readEmailAndRestoreBody(c)
		if err != nil {
			t.Fatalf("body was not restored: %v", err)
		}
		if email != "User@Test.COM" {
			t.Fatalf("email mismatch after restore: %s", email)
		}
		return okHandler(c)
	})(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
	if len(rl.loginIPs) != 1 || len(rl.loginKeys) != 1 {
		t.Fatalf("login limit calls mismatch: ip=%#v mail=%#v", rl.loginIPs, rl.loginKeys)
	}
	want := hashEmailForRateLimit("user@test.com")
	if rl.loginKeys[0] != want {
		t.Fatalf("email hash mismatch: want=%s got=%s", want, rl.loginKeys[0])
	}
}

func TestLoginRateLimit_InvalidJSON_ReturnsBadRequestAndDoesNotCallEmailLimit(t *testing.T) {
	rl := &allRateLimiterMock{allow: true}
	c, rec := newTestContext(http.MethodPost, "/auth/login", `{invalid-json`)
	err := LoginRateLimit(rl)(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusBadRequest)
	if len(rl.loginKeys) != 0 {
		t.Fatalf("email rate limit should not be called: %#v", rl.loginKeys)
	}
}

func TestRefreshRateLimit_MissingCookie_ReturnsUnauthorized(t *testing.T) {
	rl := &allRateLimiterMock{allow: true}
	c, rec := newTestContext(http.MethodPost, "/auth/refresh", "")
	err := RefreshRateLimit(rl, "refresh_token")(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestRefreshRateLimit_Allowed_UsesHashedToken(t *testing.T) {
	rl := &allRateLimiterMock{allow: true}
	c, rec := newTestContext(http.MethodPost, "/auth/refresh", "")
	c.Request().AddCookie(&http.Cookie{Name: "refresh_token", Value: " refresh-token "})
	err := RefreshRateLimit(rl, "refresh_token")(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
	if len(rl.refreshes) != 1 || rl.refreshes[0] != hashTokenForRateLimit("refresh-token") {
		t.Fatalf("refresh hash mismatch: %#v", rl.refreshes)
	}
}

func TestWsRateLimit_Denied_ReturnsTooManyRequests(t *testing.T) {
	rl := &allRateLimiterMock{allow: false, retry: 4}
	c, rec := newTestContext(http.MethodGet, "/ws/search/sessions/1", "")
	err := WsRateLimit(rl)(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusTooManyRequests)
	if rec.Header().Get("Retry-After") != "4" {
		t.Fatalf("Retry-After mismatch: %s", rec.Header().Get("Retry-After"))
	}
}
