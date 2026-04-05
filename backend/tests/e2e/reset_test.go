package e2e

import (
	"net/http"
	"testing"
)

func Test_Auth_Forgot_Reset_LoginWithNewPassword(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	u := env.createUserDirect(t, "reset-user@example.com", testUserPW, "user", true)

	// reset 前に login して refresh token を発行しておく
	oldSt := env.loginOK(t, u.Email, testUserPW)
	if oldSt.Rt == "" {
		t.Fatalf("expected refresh token before password reset")
	}

	// forgot
	forgotBody := forgotPwReq{
		Email: u.Email,
	}

	res, body := postJSON(t, env, "/auth/password/forgot", &forgotBody, nil)
	mustStatus(t, res, body, http.StatusNoContent)

	// reset token を直接投入
	resetToken := "reset-token-e2e-001"
	newPW := "NewPW123!@#X"
	env.insertResetTokenDirect(t, u.ID, resetToken)

	resetBody := resetPwReq{
		Token:       resetToken,
		NewPassword: newPW,
	}

	res, body = postJSON(t, env, "/auth/password/reset", &resetBody, nil)
	mustStatus(t, res, body, http.StatusNoContent)

	// reset 実行時に旧 refresh が revoke されていること
	revoked := env.countRevokedRtsByUser(t, u.ID)
	if revoked == 0 {
		t.Fatalf("expected revoked refresh tokens after password reset")
	}

	// 旧 access は token_ver++ により me で 401
	res, body = env.doEmpty(
		t,
		http.MethodGet,
		"/me",
		bearerHeader(oldSt.AccessToken),
	)
	mustStatus(t, res, body, http.StatusUnauthorized)
	mustErrCode(t, body, "unauthorized")

	// refresh rate limit を避ける
	env.resetRedis(t)

	// 旧 refresh も 401
	res, body = env.doEmpty(
		t,
		http.MethodPost,
		"/auth/refresh",
		addCSRFHeader(nil, oldSt.CSRFToken),
		&http.Cookie{Name: "refresh_token", Value: oldSt.Rt, Path: "/auth"},
		&http.Cookie{Name: "csrf_token", Value: oldSt.CSRFToken, Path: "/"},
	)
	mustStatus(t, res, body, http.StatusUnauthorized)
	mustErrCode(t, body, "unauthorized")

	// 古いPWでは login 失敗
	oldLoginBody := loginReq{
		Email:    u.Email,
		Password: testUserPW,
	}
	res, body = postJSON(t, env, "/auth/login", &oldLoginBody, nil)
	mustStatus(t, res, body, http.StatusUnauthorized)
	mustErrCode(t, body, "unauthorized")

	env.resetRedis(t)

	// 新しいPWでは login 成功
	st := env.loginOK(t, u.Email, newPW)
	if st.AccessToken == "" {
		t.Fatalf("new password login should issue access token")
	}
}

// used reset token の再利用は401を返す。
func Test_Auth_ResetPassword_TokenReuse_IsUnauthorized_CurrentBehavior(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	u := env.createUserDirect(t, "reset-reuse@example.com", testUserPW, "user", true)
	token := "reset-reuse-token"

	env.insertResetTokenDirect(t, u.ID, token)

	firstResetBody := resetPwReq{
		Token:       token,
		NewPassword: "BrandNewPW123!",
	}

	//1回目reset成功
	res, body := postJSON(t, env, "/auth/password/reset", &firstResetBody, nil)
	mustStatus(t, res, body, http.StatusNoContent)

	secondResetBody := resetPwReq{
		Token:       token,
		NewPassword: "AnotherPW123!",
	}

	//2回目は401
	res, body = postJSON(t, env, "/auth/password/reset", &secondResetBody, nil)
	mustStatus(t, res, body, http.StatusUnauthorized)
	mustErrCode(t, body, "unauthorized")
}
