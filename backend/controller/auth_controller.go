package controller

import (
	"net/http"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type AuthCtl struct {
	uc usecase.AuthUsecase
}

func NewAuthCtl(uc usecase.AuthUsecase) AuthCtl {
	return AuthCtl{
		uc: uc,
	}
}

// signupの入力。
type SignupReq struct {
	Email string `json:"email"`
	Pw    string `json:"password"`
}

// verify emailの入力。
type VerifyEmailReq struct {
	Token string `json:"token"`
}

// verify再送の入力。
type ResendVerifyReq struct {
	Email string `json:"email"`
}

// loginの入力。
type LoginReq struct {
	Email string `json:"email"`
	Pw    string `json:"password"`
}

// forgot passwordの入力。
type ForgotPwReq struct {
	Email string `json:"email"`
}

// reset passwordの入力。
type ResetPwReq struct {
	Token string `json:"token"`
	NewPw string `json:"new_password"`
}

// 認証系レスポンスで返すuser 。
type AuthUserRes struct {
	ID            int64  `json:"id"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	TokenVer      int    `json:"token_ver"`
	EmailVerified bool   `json:"email_verified"`
}

// signup成功レスポンス。
type SignupRes struct {
	User AuthUserRes `json:"user"`
}

// login / refresh成功レスポンス。
type AuthRes struct {
	AccessToken string      `json:"access_token"`
	User        AuthUserRes `json:"user"`
}

// /me成功レスポンス。
type MeRes struct {
	User AuthUserRes `json:"user"`
}

// POST /auth/signupを処理。
func (ctl AuthCtl) Signup(c echo.Context) error {
	var req SignupReq

	if err := bindJSON(c, &req); err != nil {
		return err
	}

	u, err := ctl.uc.Signup(usecase.SignupIn{
		Email: req.Email,
		Pw:    req.Pw,
		IP:    realIP(c),
		UA:    userAgent(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusCreated, SignupRes{
		User: toAuthUserRes(u),
	})
}

// POST /auth/verify-emailを処理。
func (ctl AuthCtl) VerifyEmail(c echo.Context) error {
	var req VerifyEmailReq

	if err := bindJSON(c, &req); err != nil {
		return err
	}

	err := ctl.uc.VerifyEmail(usecase.VerifyEmailIn{
		Token: req.Token,
		IP:    realIP(c),
		UA:    userAgent(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// POST /auth/resend-verifyを処理。
func (ctl AuthCtl) ResendVerify(c echo.Context) error {
	var req ResendVerifyReq

	if err := bindJSON(c, &req); err != nil {
		return err
	}

	err := ctl.uc.ResendVerify(usecase.ResendVerifyIn{
		Email: req.Email,
		IP:    realIP(c),
		UA:    userAgent(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// POST /auth/loginを処理。
func (ctl AuthCtl) Login(c echo.Context) error {
	var req LoginReq

	if err := bindJSON(c, &req); err != nil {
		return err
	}

	loginResult, err := ctl.uc.Login(usecase.LoginIn{
		Email: req.Email,
		Pw:    req.Pw,
		IP:    realIP(c),
		UA:    userAgent(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	setRefreshCookie(c, loginResult.Rt)
	setCSRFCookie(c, loginResult.CsrfToken)

	return c.JSON(http.StatusOK, AuthRes{
		AccessToken: loginResult.AccessToken,
		User:        toAuthUserRes(loginResult.User),
	})
}

// POST /auth/refreshを処理。
func (ctl AuthCtl) Refresh(c echo.Context) error {
	refreshResult, err := ctl.uc.Refresh(usecase.RefreshIn{
		Rt: refreshCookie(c),
		IP: realIP(c),
		UA: userAgent(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	setRefreshCookie(c, refreshResult.Rt)
	setCSRFCookie(c, refreshResult.CsrfToken)

	return c.JSON(http.StatusOK, AuthRes{
		AccessToken: refreshResult.AccessToken,
		User:        toAuthUserRes(refreshResult.User),
	})
}

// POST /auth/logoutを処理。
func (ctl AuthCtl) Logout(c echo.Context) error {
	err := ctl.uc.Logout(usecase.LogoutIn{
		UserID: userIDFromCtx(c),
		Rt:     refreshCookie(c),
		IP:     realIP(c),
		UA:     userAgent(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	clearRefreshCookie(c)
	clearCSRFCookie(c)

	return c.NoContent(http.StatusNoContent)
}

// GET /meを処理。
func (ctl AuthCtl) Me(c echo.Context) error {
	u, err := ctl.uc.Me(userIDFromCtx(c))
	if err != nil {
		return writeErr(c, err)
	}

	c.Response().Header().Set("Cache-Control", "no-store")

	return c.JSON(http.StatusOK, MeRes{
		User: toAuthUserRes(u),
	})
}

// POST /auth/password/forgotを処理。
func (ctl AuthCtl) ForgotPw(c echo.Context) error {
	var req ForgotPwReq

	if err := bindJSON(c, &req); err != nil {
		return err
	}

	err := ctl.uc.ForgotPw(usecase.ForgotPwIn{
		Email: req.Email,
		IP:    realIP(c),
		UA:    userAgent(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// POST /auth/password/resetを処理。
func (ctl AuthCtl) ResetPw(c echo.Context) error {
	var req ResetPwReq

	if err := bindJSON(c, &req); err != nil {
		return err
	}

	err := ctl.uc.ResetPw(usecase.ResetPwIn{
		Token: req.Token,
		NewPw: req.NewPw,
		IP:    realIP(c),
		UA:    userAgent(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// entity.Userを認証レスポンス用に変換。
func toAuthUserRes(u entity.User) AuthUserRes {
	return AuthUserRes{
		ID:            u.ID,
		Email:         u.Email,
		Role:          u.Role,
		TokenVer:      u.TokenVer,
		EmailVerified: u.EmailVerified,
	}
}
