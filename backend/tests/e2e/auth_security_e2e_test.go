package e2e

import (
	"net/http"
	"testing"
)

// TestE2E45_Auth_LoginInvalidEmail_ReturnsBadRequest は、email形式不正が400になることを検証する。
func TestE2E45_Auth_LoginInvalidEmail_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)
	_, password := adminCreds()

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/login", map[string]string{
		"email":    "not-email",
		"password": password,
	}, map[string]string{
		"X-Forwarded-For": e2e45UniqueIP(),
	})
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}

// TestE2E45_Auth_LoginMissingEmail_ReturnsBadRequest は、email未指定が400になることを検証する。
func TestE2E45_Auth_LoginMissingEmail_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)
	_, password := adminCreds()

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/login", map[string]string{
		"password": password,
	}, map[string]string{
		"X-Forwarded-For": e2e45UniqueIP(),
	})
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}

// TestE2E45_Auth_LoginMissingPassword_ReturnsBadRequest は、password未指定が400になることを検証する。
func TestE2E45_Auth_LoginMissingPassword_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/login", map[string]string{
		"email": e2e45UniqueEmail("missing_pw"),
	}, map[string]string{
		"X-Forwarded-For": e2e45UniqueIP(),
	})
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}

// TestE2E45_Auth_LoginWrongPassword_ReturnsUnauthorized は、認証失敗時が401になることを検証する。
func TestE2E45_Auth_LoginWrongPassword_ReturnsUnauthorized(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/login", map[string]string{
		"email":    e2e45UniqueEmail("wrong_pw"),
		"password": "WrongPass123!",
	}, map[string]string{
		"X-Forwarded-For": e2e45UniqueIP(),
	})
	requireStatus(t, status, http.StatusUnauthorized, body)
	e2e45RequireErrorCode(t, body, "unauthorized")
}

// TestE2E45_Auth_MeWithoutToken_ReturnsUnauthorized は、private routeがtokenなしを拒否することを検証する。
func TestE2E45_Auth_MeWithoutToken_ReturnsUnauthorized(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.getJSON(t, "/me", nil)
	requireStatus(t, status, http.StatusUnauthorized, body)
	e2e45RequireErrorCode(t, body, "unauthorized")
}

// TestE2E45_Auth_MeWithBrokenToken_ReturnsUnauthorized は、壊れたBearer tokenが401になることを検証する。
func TestE2E45_Auth_MeWithBrokenToken_ReturnsUnauthorized(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.getJSON(t, "/me", bearer("broken.jwt.token"))
	requireStatus(t, status, http.StatusUnauthorized, body)
	e2e45RequireErrorCode(t, body, "unauthorized")
}

// TestE2E45_Auth_LogoutWithoutToken_ReturnsUnauthorized は、logoutが認証なしでは通らないことを検証する。
func TestE2E45_Auth_LogoutWithoutToken_ReturnsUnauthorized(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/logout", nil, nil)
	requireStatus(t, status, http.StatusUnauthorized, body)
	e2e45RequireErrorCode(t, body, "unauthorized")
}

// TestE2E45_Auth_LogoutWithBrokenToken_ReturnsUnauthorized は、壊れたtokenでlogoutできないことを検証する。
func TestE2E45_Auth_LogoutWithBrokenToken_ReturnsUnauthorized(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/logout", nil, bearer("broken.jwt.token"))
	requireStatus(t, status, http.StatusUnauthorized, body)
	e2e45RequireErrorCode(t, body, "unauthorized")
}

// TestE2E45_Auth_RefreshWithoutRefreshCookie_ReturnsUnauthorized は、csrfが正しくてもrefresh cookieなしなら401になることを検証する。
func TestE2E45_Auth_RefreshWithoutRefreshCookie_ReturnsUnauthorized(t *testing.T) {
	c := newAPIClient(t)
	csrf := e2e45GetCSRF(t, c)

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/refresh", nil, map[string]string{
		"X-CSRF-Token": csrf,
	})
	requireStatus(t, status, http.StatusUnauthorized, body)
	e2e45RequireErrorCode(t, body, "unauthorized")
}

// TestE2E45_Auth_RefreshCSRFVariants_ReturnsForbidden は、CSRF cookie/headerの不一致系を1回のloginでまとめて検証する。
// 成功loginを増やしすぎるとlogin RateLimitに巻き込まれるため、同じrefresh cookieを使って異常系だけを確認する。
func TestE2E45_Auth_RefreshCSRFVariants_ReturnsForbidden(t *testing.T) {
	loginClient := newAPIClient(t)
	e2e45GetCSRF(t, loginClient)
	e2e45LoginAdmin(t, loginClient)

	refreshCookie := e2e45Cookie(t, loginClient, "refresh_token")
	csrfCookie := e2e45Cookie(t, loginClient, "csrf_token")

	t.Run("header but no csrf cookie", func(t *testing.T) {
		c := newAPIClient(t)
		e2e45SetCookie(t, c, refreshCookie)

		status, body, _ := c.doJSON(t, http.MethodPost, "/auth/refresh", nil, map[string]string{
			"X-CSRF-Token": csrfCookie.Value,
		})
		requireStatus(t, status, http.StatusForbidden, body)
		e2e45RequireErrorCode(t, body, "forbidden")
	})

	t.Run("invalid csrf", func(t *testing.T) {
		c := newAPIClient(t)
		e2e45SetCookie(t, c, refreshCookie)
		e2e45SetCookie(t, c, csrfCookie)

		status, body, _ := c.doJSON(t, http.MethodPost, "/auth/refresh", nil, map[string]string{
			"X-CSRF-Token": "invalid-csrf-token",
		})
		requireStatus(t, status, http.StatusForbidden, body)
		e2e45RequireErrorCode(t, body, "forbidden")
	})
}

// TestE2E45_Auth_SignupInvalidEmail_ReturnsBadRequest は、signupのemail形式不正が400になることを検証する。
func TestE2E45_Auth_SignupInvalidEmail_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/signup", map[string]string{
		"email":    "not-email",
		"password": "ValidPass123!",
	}, nil)
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}

// TestE2E45_Auth_SignupWeakPassword_ReturnsBadRequest は、signupの弱いpasswordが400になることを検証する。
func TestE2E45_Auth_SignupWeakPassword_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/signup", map[string]string{
		"email":    e2e45UniqueEmail("weak_pw"),
		"password": "short",
	}, nil)
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}
