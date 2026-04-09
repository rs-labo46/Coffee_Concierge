package controller

import (
	"net/http"
	"os"
	"strings"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

const (
	// refresh tokenを保存するcookie名。
	refreshCookieName = "refresh_token"

	// refresh tokenをのせるcookieのpath。refresh/logoutなど認証でcookieを使う。
	refreshCookiePath = "/auth"
)

// 認証系HTTPendpointの入口で、依存先はusecase.AuthUCのみ。
type AuthCtl struct {
	uc usecase.AuthUC
}

// AuthCtlを生成。
func NewAuthCtl(uc usecase.AuthUC) *AuthCtl {
	return &AuthCtl{
		uc: uc,
	}
}

// signupのrequestbody 。
type SignupReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// verifyemailのrequest body 。
type VerifyEmailReq struct {
	Token string `json:"token"`
}

// loginのrequest body 。
type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// password forgotのrequestbody 。
type ForgotPwReq struct {
	Email string `json:"email"`
}

// password resetのrequest body 。
type ResetPwReq struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// 認証系のresponseで返すuserの形。
type AuthUserRes struct {
	ID            uint   `json:"id"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	TokenVer      int    `json:"token_ver"`
	EmailVerified bool   `json:"email_verified"`
}

// signup成功レスポンス。
type SignupRes struct {
	User AuthUserRes `json:"user"`
}

// login/refresh成功時のレスポンス。
type AuthRes struct {
	AccessToken string      `json:"access_token"`
	User        AuthUserRes `json:"user"`
}

// /meのレスポンス。
type MeRes struct {
	User AuthUserRes `json:"user"`
}

// POST /auth/signup
func (ctl *AuthCtl) Signup(c echo.Context) error {
	var req SignupReq

	// request bodyのbindに失敗したら、HTTP境界の入力不正。
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	out, err := ctl.uc.Signup(usecase.SignupIn{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusCreated, SignupRes{
		User: toAuthUserRes(out.User),
	})
}

// POST /auth/verify-email
func (ctl *AuthCtl) VerifyEmail(c echo.Context) error {
	var req VerifyEmailReq

	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	err := ctl.uc.VerifyEmail(usecase.VerifyEmailIn{
		Token: req.Token,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, MsgRes{
		Message: "email verified",
	})
}

// POST /auth/login
func (ctl *AuthCtl) Login(c echo.Context) error {
	var req LoginReq

	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	out, err := ctl.uc.Login(usecase.LoginIn{
		Email:    req.Email,
		Password: req.Password,
		UA:       userAgent(c),
		IP:       clientIP(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	// refresh tokenはHttpOnly cookieに保存。
	setRefreshCookie(
		c,
		refreshCookieName,
		out.RefreshToken,
		authRefreshCookieMaxAgeSec(),
		authCookieSecure(),
		http.SameSiteLaxMode,
		refreshCookiePath,
		authCookieDomain(),
	)

	return c.JSON(http.StatusOK, AuthRes{
		AccessToken: out.AccessToken,
		User:        toAuthUserRes(out.User),
	})
}

// POST /auth/refresh
// cookieのrefresh tokenを使う。
func (ctl *AuthCtl) Refresh(c echo.Context) error {
	// cookie から refresh token を取ります。
	refreshToken := cookieValue(c, refreshCookieName)

	out, err := ctl.uc.Refresh(usecase.RefreshIn{
		RefreshToken: refreshToken,
		UA:           userAgent(c),
		IP:           clientIP(c),
	})
	if err != nil {
		return writeErr(c, err)
	}

	// rotate後の新しいrefresh tokenをcookieに再設定。
	setRefreshCookie(
		c,
		refreshCookieName,
		out.RefreshToken,
		authRefreshCookieMaxAgeSec(),
		authCookieSecure(),
		http.SameSiteLaxMode,
		refreshCookiePath,
		authCookieDomain(),
	)

	return c.JSON(http.StatusOK, AuthRes{
		AccessToken: out.AccessToken,
		User:        toAuthUserRes(out.User),
	})
}

// POST /auth/logout
func (ctl *AuthCtl) Logout(c echo.Context) error {
	// 認証必須だからactorを要求。
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}

	//cookieから取る。
	refreshToken := cookieValue(c, refreshCookieName)

	err = ctl.uc.Logout(*actor, refreshToken)
	if err != nil {
		return writeErr(c, err)
	}

	// logoutの成功時はrefresh cookieを削除。
	clearRefreshCookie(
		c,
		refreshCookieName,
		authCookieSecure(),
		http.SameSiteLaxMode,
		refreshCookiePath,
		authCookieDomain(),
	)

	return c.JSON(http.StatusOK, MsgRes{
		Message: "logged out",
	})
}

// POST /auth/password/forgot
func (ctl *AuthCtl) ForgotPw(c echo.Context) error {
	var req ForgotPwReq

	// request body の bind 失敗は invalid_request 。
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	err := ctl.uc.ForgotPw(usecase.ForgotPwIn{
		Email: req.Email,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, MsgRes{
		Message: "if the email exists, a reset mail has been sent",
	})
}

// POST /auth/password/reset
func (ctl *AuthCtl) ResetPw(c echo.Context) error {
	var req ResetPwReq

	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	err := ctl.uc.ResetPw(usecase.ResetPwIn{
		Token:    req.Token,
		Password: req.Password,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, MsgRes{
		Message: "password reset completed",
	})
}

// GET /me
// actorから自分のuserを取得して返す。
// キャッシュしてはいけないからno-store。
func (ctl *AuthCtl) Me(c echo.Context) error {
	// 認証必須。
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}

	u, err := ctl.uc.Me(*actor)
	if err != nil {
		return writeErr(c, err)
	}

	// 認証情報に紐づくレスポンスだかエアキャッシュ禁止。
	setNoStore(c)

	return c.JSON(http.StatusOK, MeRes{
		User: toAuthUserRes(u),
	})
}

// entity.Userを認証系レスポンス用のshapeに変換。
func toAuthUserRes(u entity.User) AuthUserRes {
	return AuthUserRes{
		ID:            u.ID,
		Email:         u.Email,
		Role:          string(u.Role),
		TokenVer:      u.TokenVer,
		EmailVerified: u.EmailVerified,
	}
}

// refresh cookieにSecureを付けるかを返す。
func authCookieSecure() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE")))
	return v == "1" || v == "true" || v == "yes"
}

// cookieのdomainを返す。未指定なら空文字のままに。
func authCookieDomain() string {
	return strings.TrimSpace(os.Getenv("COOKIE_DOMAIN"))
}

// refresh cookieのMaxAge秒数。
func authRefreshCookieMaxAgeSec() int {
	return 24 * 60 * 60
}
