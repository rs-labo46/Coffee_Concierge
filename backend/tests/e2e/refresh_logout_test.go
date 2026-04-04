package e2e

import (
	"net/http"
	"testing"
)

func Test_Auth_Refresh_Rotation_And_ReuseRejected(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	u := env.createUserDirect(t, "refresh-user@example.com", testUserPW, "user", true)

	st1 := env.loginOK(t, u.Email, testUserPW)

	oldRefresh := st1.RefreshToken
	oldCSRF := st1.CSRFToken

	// 1回目 refresh 成功
	refreshHeaders := addCSRFHeader(nil, oldCSRF)

	res, body := env.doEmpty(
		t,
		http.MethodPost,
		"/auth/refresh",
		refreshHeaders,
		authCookies(st1)...,
	)
	mustStatus(t, res, body, http.StatusOK)

	var out authRes
	mustJSON(t, body, &out)

	newRefreshCookie := mustCookie(t, res, "refresh_token")
	newCSRFCookie := mustCookie(t, res, "csrf_token")

	if newRefreshCookie.Value == "" || newRefreshCookie.Value == oldRefresh {
		t.Fatalf("refresh token was not rotated")
	}
	if newCSRFCookie.Value == "" {
		t.Fatalf("new csrf token is empty")
	}
	if out.AccessToken == "" {
		t.Fatalf("new access token is empty")
	}

	//rate limit 状態を消す
	env.resetRedis(t)

	// old refresh を再使用 → reuse 検知で 401
	res, body = env.doEmpty(
		t,
		http.MethodPost,
		"/auth/refresh",
		addCSRFHeader(nil, oldCSRF),
		&http.Cookie{Name: "refresh_token", Value: oldRefresh, Path: "/auth"},
		&http.Cookie{Name: "csrf_token", Value: oldCSRF, Path: "/"},
	)
	mustStatus(t, res, body, http.StatusUnauthorized)
	mustErrCode(t, body, "unauthorized")
}

func Test_Auth_Logout_Then_Me_IsUnauthorized(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	u := env.createUserDirect(t, "logout-user@example.com", testUserPW, "user", true)

	st := env.loginOK(t, u.Email, testUserPW)

	//logoutはJWT保護のみ
	res, body := env.doEmpty(
		t,
		http.MethodPost,
		"/auth/logout",
		bearerHeader(st.AccessToken),
		&http.Cookie{Name: "refresh_token", Value: st.RefreshToken, Path: "/auth"},
	)
	mustStatus(t, res, body, http.StatusNoContent)

	//logout後は古いaccess tokenで/meが401
	res, body = env.doEmpty(
		t,
		http.MethodGet,
		"/me",
		bearerHeader(st.AccessToken),
	)
	mustStatus(t, res, body, http.StatusUnauthorized)
	mustErrCode(t, body, "unauthorized")
}
