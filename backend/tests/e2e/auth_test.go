package e2e

import (
	"net/http"
	"testing"
)

func Test_Auth_Signup_Verify_Login_Me(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	email := "user1@example.com"

	//signupする
	signupBody := signupReq{
		Email:    email,
		Password: testUserPW,
	}

	res, body := postJSON(t, env, "/auth/signup", &signupBody, nil)
	mustStatus(t, res, body, http.StatusCreated)

	var su signupRes
	mustJSON(t, body, &su)

	if su.User.Email != email {
		t.Fatalf("signup email=%q want=%q", su.User.Email, email)
	}
	if su.User.EmailVerified {
		t.Fatalf("signup should create unverified user")
	}

	//verify前loginはunauthorized
	loginBody := loginReq{
		Email:    email,
		Password: testUserPW,
	}

	res, body = postJSON(t, env, "/auth/login", &loginBody, nil)
	mustStatus(t, res, body, http.StatusUnauthorized)
	mustErrCode(t, body, "unauthorized")
	env.resetRedis(t)

	// 既に知っているtokenをDBに直接投入する
	user := env.findUserByEmail(t, email)
	verifyToken := "verify-token-e2e-001"
	env.insertVerifyTokenDirect(t, user.ID, verifyToken)

	verifyBody := verifyEmailReq{
		Token: verifyToken,
	}

	res, body = postJSON(t, env, "/auth/verify-email", &verifyBody, nil)
	mustStatus(t, res, body, http.StatusNoContent)

	//verify後はlogin成功
	st := env.loginOK(t, email, testUserPW)
	if st.AccessToken == "" {
		t.Fatalf("access token is empty")
	}
	if st.RefreshToken == "" {
		t.Fatalf("refresh token cookie is empty")
	}
	if st.CSRFToken == "" {
		t.Fatalf("csrf cookie is empty")
	}

	///me で本人情報が取れる
	res, body = env.doEmpty(t, http.MethodGet, "/me", bearerHeader(st.AccessToken))
	mustStatus(t, res, body, http.StatusOK)
	mustNoStore(t, res)

	var me meRes
	mustJSON(t, body, &me)

	if me.User.Email != email {
		t.Fatalf("me email=%q want=%q", me.User.Email, email)
	}
	if !me.User.EmailVerified {
		t.Fatalf("/me should return verified user")
	}
}

// verify token の再使用は401を返す。
func Test_Auth_VerifyEmail_TokenReuse_IsUnauthorized_CurrentBehavior(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	u := env.createUserDirect(t, "verify-reuse@example.com", testUserPW, "user", false)
	token := "verify-reuse-token"
	env.insertVerifyTokenDirect(t, u.ID, token)

	verifyBody := verifyEmailReq{
		Token: token,
	}

	//1回目は成功
	res, body := postJSON(t, env, "/auth/verify-email", &verifyBody, nil)
	mustStatus(t, res, body, http.StatusNoContent)

	//2回目は401
	res, body = postJSON(t, env, "/auth/verify-email", &verifyBody, nil)
	mustStatus(t, res, body, http.StatusUnauthorized)
	mustErrCode(t, body, "unauthorized")
}
