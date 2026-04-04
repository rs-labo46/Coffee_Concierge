package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

// auth usecaseの代役。
// controllerがusecaseをどう使うかだけを確認
type mockAuthUC struct {
	signupFn      func(in usecase.SignupIn) (entity.User, error)
	verifyEmailFn func(in usecase.VerifyEmailIn) error
	resendFn      func(in usecase.ResendVerifyIn) error
	loginFn       func(in usecase.LoginIn) (usecase.AuthOut, error)
	refreshFn     func(in usecase.RefreshIn) (usecase.AuthOut, error)
	logoutFn      func(in usecase.LogoutIn) error
	forgotFn      func(in usecase.ForgotPwIn) error
	resetFn       func(in usecase.ResetPwIn) error
	meFn          func(userID int64) (entity.User, error)
}

func (m *mockAuthUC) Signup(in usecase.SignupIn) (entity.User, error) {
	return m.signupFn(in)
}
func (m *mockAuthUC) VerifyEmail(in usecase.VerifyEmailIn) error {
	return m.verifyEmailFn(in)
}
func (m *mockAuthUC) ResendVerify(in usecase.ResendVerifyIn) error {
	return m.resendFn(in)
}
func (m *mockAuthUC) Login(in usecase.LoginIn) (usecase.AuthOut, error) {
	return m.loginFn(in)
}
func (m *mockAuthUC) Refresh(in usecase.RefreshIn) (usecase.AuthOut, error) {
	return m.refreshFn(in)
}
func (m *mockAuthUC) Logout(in usecase.LogoutIn) error {
	return m.logoutFn(in)
}
func (m *mockAuthUC) ForgotPw(in usecase.ForgotPwIn) error {
	return m.forgotFn(in)
}
func (m *mockAuthUC) ResetPw(in usecase.ResetPwIn) error {
	return m.resetFn(in)
}
func (m *mockAuthUC) Me(userID int64) (entity.User, error) {
	return m.meFn(userID)
}

// /meのレスポンスにCache-Control:no-storeが付き、contextのuser_id がusecaseに渡ることを確認。
func TestAuthCtlMe_NoStore(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// middlewareがuser_id をcontextに入れた想定を作る。
	c.Set("user_id", int64(1))

	called := false
	ctl := NewAuthCtl(&mockAuthUC{
		meFn: func(userID int64) (entity.User, error) {
			called = true
			if userID != 1 {
				t.Fatalf("userID = %d, want 1", userID)
			}
			return entity.User{
				ID:            1,
				Email:         "a@test.com",
				Role:          "user",
				TokenVer:      2,
				EmailVerified: true,
			}, nil
		},
	})

	if err := ctl.Me(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("usecase.Me was not called")
	}

	// 正常終了なので200を期待。
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	// 個人情報レスポンスはキャッシュさせない。
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}

	var body MeRes
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if body.User.ID != 1 || body.User.Email != "a@test.com" || body.User.Role != "user" || body.User.TokenVer != 2 || !body.User.EmailVerified {
		t.Fatalf("unexpected user body: %+v", body.User)
	}
}
