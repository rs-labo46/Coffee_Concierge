package e2e

import (
	"net/http"
	"testing"
)

type csrfRes struct {
	Token string `json:"token"`
}

type authRes struct {
	AccessToken string `json:"access_token"`
	User        struct {
		ID            uint   `json:"id"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		TokenVer      int    `json:"token_ver"`
		EmailVerified bool   `json:"email_verified"`
	} `json:"user"`
}

// 管理者ユーザーでログインし、access token・refresh cookie・CSRFを使った認証フローが通るか確認
func TestE2E_Auth_AdminLoginMeRefreshLogout(t *testing.T) {
	c := newAPIClient(t)
	email, password := adminCreds()

	csrfStatus, csrfBody, _ := c.getJSON(t, "/auth/csrf", nil)
	requireStatus(t, csrfStatus, http.StatusOK, csrfBody)

	var csrf csrfRes
	decodeJSON(t, csrfBody, &csrf)
	if csrf.Token == "" {
		t.Fatalf("csrf token is empty")
	}

	loginStatus, loginBody, _ := c.doJSON(t, http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, nil)
	requireStatus(t, loginStatus, http.StatusOK, loginBody)

	var login authRes
	decodeJSON(t, loginBody, &login)
	if login.AccessToken == "" {
		t.Fatalf("access token is empty")
	}
	if login.User.Email != email {
		t.Fatalf("unexpected login email: got=%s want=%s", login.User.Email, email)
	}
	if login.User.Role != "admin" {
		t.Fatalf("unexpected login role: %s", login.User.Role)
	}
	if !login.User.EmailVerified {
		t.Fatalf("seed admin should be email verified")
	}

	meStatus, meBody, _ := c.getJSON(t, "/me", bearer(login.AccessToken))
	requireStatus(t, meStatus, http.StatusOK, meBody)

	refreshStatus, refreshBody, _ := c.doJSON(t, http.MethodPost, "/auth/refresh", nil, map[string]string{
		"X-CSRF-Token": csrf.Token,
	})
	requireStatus(t, refreshStatus, http.StatusOK, refreshBody)

	var refresh authRes
	decodeJSON(t, refreshBody, &refresh)
	if refresh.AccessToken == "" {
		t.Fatalf("refreshed access token is empty")
	}
	logoutStatus, logoutBody, _ := c.doJSON(t, http.MethodPost, "/auth/logout", nil, bearer(refresh.AccessToken))
	requireStatus(t, logoutStatus, http.StatusOK, logoutBody)
}

// refresh cookieがあっても、CSRF headerがない場合にrefreshを拒否できるか確認
func TestE2E_Auth_RefreshWithoutCSRF_ReturnsForbidden(t *testing.T) {
	c := newAPIClient(t)
	email, password := adminCreds()

	loginStatus, loginBody, _ := c.doJSON(t, http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, nil)
	requireStatus(t, loginStatus, http.StatusOK, loginBody)

	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/refresh", nil, nil)
	requireStatus(t, status, http.StatusForbidden, body)

	var errRes struct {
		Error string `json:"error"`
	}
	decodeJSON(t, body, &errRes)
	if errRes.Error != "forbidden" {
		t.Fatalf("unexpected error code: %s", errRes.Error)
	}
}
