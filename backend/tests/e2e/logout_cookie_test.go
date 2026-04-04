package e2e

import (
	"net/http"
	"testing"
	"time"
)

func Test_Auth_Logout_ClearsRefreshAndCSRFCookies(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	u := env.createUserDirect(t, "logout-cookie@example.com", testUserPW, "user", true)
	st := env.loginOK(t, u.Email, testUserPW)

	res, body := env.doEmpty(
		t,
		http.MethodPost,
		"/auth/logout",
		bearerHeader(st.AccessToken),
		&http.Cookie{Name: "refresh_token", Value: st.RefreshToken, Path: "/auth"},
	)
	mustStatus(t, res, body, http.StatusNoContent)

	refresh := mustCookie(t, res, "refresh_token")
	csrf := mustCookie(t, res, "csrf_token")

	if refresh.MaxAge != 0 {
		t.Fatalf("refresh Max-Age = %d, want 0", refresh.MaxAge)
	}
	if csrf.MaxAge != 0 {
		t.Fatalf("csrf Max-Age = %d, want 0", csrf.MaxAge)
	}
	if refresh.Path != "/auth" {
		t.Fatalf("refresh path = %q, want /auth", refresh.Path)
	}
	if csrf.Path != "/" {
		t.Fatalf("csrf path = %q, want /", csrf.Path)
	}
	if !refresh.Expires.Before(time.Now().Add(1 * time.Second)) {
		t.Fatalf("refresh expires should be in the past: %v", refresh.Expires)
	}
	if !csrf.Expires.Before(time.Now().Add(1 * time.Second)) {
		t.Fatalf("csrf expires should be in the past: %v", csrf.Expires)
	}
}
