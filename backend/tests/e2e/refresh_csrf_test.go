package e2e

import (
	"net/http"
	"testing"
)

func Test_Auth_Refresh_CSRFMismatch_ReturnsForbidden(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	u := env.createUserDirect(t, "csrf-refresh@example.com", testUserPW, "user", true)
	st := env.loginOK(t, u.Email, testUserPW)

	t.Run("missing csrf cookie", func(t *testing.T) {
		res, body := env.doEmpty(
			t,
			http.MethodPost,
			"/auth/refresh",
			addCSRFHeader(nil, st.CSRFToken),
			&http.Cookie{Name: "refresh_token", Value: st.Rt, Path: "/auth"},
		)
		mustStatus(t, res, body, http.StatusForbidden)
		mustErrCode(t, body, "forbidden")
	})

	t.Run("missing csrf header", func(t *testing.T) {
		res, body := env.doEmpty(
			t,
			http.MethodPost,
			"/auth/refresh",
			nil,
			authCookies(st)...,
		)
		mustStatus(t, res, body, http.StatusForbidden)
		mustErrCode(t, body, "forbidden")
	})

	t.Run("csrf mismatch", func(t *testing.T) {
		res, body := env.doEmpty(
			t,
			http.MethodPost,
			"/auth/refresh",
			addCSRFHeader(nil, "different-csrf"),
			authCookies(st)...,
		)
		mustStatus(t, res, body, http.StatusForbidden)
		mustErrCode(t, body, "forbidden")
	})
}
