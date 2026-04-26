package e2e

import (
	"net/http"
	"testing"
)

// TestE2E45_Cookie_CSRFBodyMatchesCookie は、/auth/csrf のbody tokenとcsrf_token cookieが一致することを検証する。
func TestE2E45_Cookie_CSRFBodyMatchesCookie(t *testing.T) {
	c := newAPIClient(t)
	csrf := e2e45GetCSRF(t, c)
	ck := e2e45Cookie(t, c, "csrf_token")

	if ck.Value == "" {
		t.Fatalf("csrf_token cookie is empty")
	}
	if ck.Value != csrf {
		t.Fatalf("csrf body token and cookie token mismatch")
	}
}

// TestE2E45_Cookie_LoginRefreshLogoutFlow は、refresh_token cookieの発行・rotation・削除を1回のloginで検証する。
// Cookie系で複数回admin loginするとlogin RateLimitに巻き込まれるため、cookie関連の正常系を1本にまとめる。
func TestE2E45_Cookie_LoginRefreshLogoutFlow(t *testing.T) {
	c := newAPIClient(t)
	csrf := e2e45GetCSRF(t, c)
	access := e2e45LoginAdmin(t, c)

	before := e2e45Cookie(t, c, "refresh_token")
	if before.Value == "" {
		t.Fatalf("refresh_token cookie is empty")
	}

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/refresh", nil, map[string]string{
		"X-CSRF-Token": csrf,
	})
	requireStatus(t, status, http.StatusOK, body)

	after := e2e45Cookie(t, c, "refresh_token")
	if after.Value == "" {
		t.Fatalf("rotated refresh_token cookie is empty")
	}
	if before.Value == after.Value {
		t.Fatalf("refresh_token was not rotated")
	}

	var refreshRes authRes
	decodeJSON(t, body, &refreshRes)
	if refreshRes.AccessToken != "" {
		access = refreshRes.AccessToken
	}

	status, body, _ = c.doJSON(t, http.MethodPost, "/auth/logout", nil, bearer(access))
	requireStatus(t, status, http.StatusOK, body)

	if e2e45CookieExists(t, c, "refresh_token") {
		t.Fatalf("refresh_token cookie still exists after logout")
	}
}
