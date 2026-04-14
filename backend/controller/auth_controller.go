package controller

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"strconv"
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

	// csrf tokenを保存するcookie。
	csrfCookieName = "csrf_token"

	// csrf tokenをのせるcookieのpath。
	csrfCookiePath = "/auth"
)

// 認証系HTTPendpointの入口で、依存先はusecase.AuthUCのみ。
type AuthCtl struct {
	uc usecase.AuthUC
	rl usecase.RateLimiter
}

// AuthCtlを生成。
func NewAuthCtl(uc usecase.AuthUC, rl usecase.RateLimiter) *AuthCtl {
	return &AuthCtl{
		uc: uc,
		rl: rl,
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

// verify mail再送のrequest body 。
type ResendVerifyReq struct {
	Email string `json:"email"`
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

type CsrfRes struct {
	Token string `json:"token"`
}

// GET /auth/csrf
// 未認証・認証済みを問わずcsrf tokenを出す。
// cookieとresponse bodyの両方に同じtokenを返す。
func (ctl *AuthCtl) Csrf(c echo.Context) error {
	token, err := newCSRFToken()
	if err != nil {
		return writeErr(c, err)
	}

	c.SetCookie(&http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     csrfCookiePath,
		MaxAge:   authCSRFCookieMaxAgeSec(),
		HttpOnly: false,
		Secure:   authCookieSecure(),
		SameSite: http.SameSiteLaxMode,
		Domain:   authCookieDomain(),
	})

	return c.JSON(http.StatusOK, CsrfRes{
		Token: token,
	})
}

// POST /auth/signup
func (ctl *AuthCtl) Signup(c echo.Context) error {
	var req SignupReq

	// request bodyのbindに失敗したら、HTTP境界の入力不正。
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	allowed, retryAfterSec, err := ctl.rl.AllowSignup(clientIP(c))
	if err != nil {
		return writeErr(c, err)
	}
	if !allowed {
		return writeRateLimited(c, retryAfterSec)
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

// POST /auth/verify-email/resend
func (ctl *AuthCtl) ResendVerify(c echo.Context) error {
	var req ResendVerifyReq

	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	ip := clientIP(c)

	// 1段目: IP 単位の制限。
	allowed, retryAfterSec, err := ctl.rl.AllowResendIP(ip)
	if err != nil {
		return writeErr(c, err)
	}
	if !allowed {
		return writeRateLimited(c, retryAfterSec)
	}

	// 2段目: email 単位の制限。
	emailHash := hashEmailForRateLimit(req.Email)

	allowed, retryAfterSec, err = ctl.rl.AllowResendMail(emailHash)
	if err != nil {
		return writeErr(c, err)
	}
	if !allowed {
		return writeRateLimited(c, retryAfterSec)
	}

	err = ctl.uc.ResendVerify(usecase.ResendVerifyIn{
		Email: req.Email,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, MsgRes{
		Message: "if the email exists, a verify mail has been sent",
	})
}

// POST /auth/login
func (ctl *AuthCtl) Login(c echo.Context) error {
	var req LoginReq

	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	ip := clientIP(c)

	// 1:IP単位の制限。
	allowed, retryAfterSec, err := ctl.rl.AllowLoginIP(ip)
	if err != nil {
		return writeErr(c, err)
	}
	if !allowed {
		return writeRateLimited(c, retryAfterSec)
	}

	// 2:email単位の制限。
	emailHash := hashEmailForRateLimit(req.Email)

	allowed, retryAfterSec, err = ctl.rl.AllowLogin(emailHash)
	if err != nil {
		return writeErr(c, err)
	}
	if !allowed {
		return writeRateLimited(c, retryAfterSec)
	}

	out, err := ctl.uc.Login(usecase.LoginIn{
		Email:    req.Email,
		Password: req.Password,
		UA:       userAgent(c),
		IP:       ip,
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
	// cookieからrefresh tokenを取る。
	refreshToken := cookieValue(c, refreshCookieName)
	refreshTokenHash := hashTokenForRateLimit(refreshToken)

	allowed, retryAfterSec, err := ctl.rl.AllowRefreshToken(refreshTokenHash)
	if err != nil {
		return writeErr(c, err)
	}
	if !allowed {
		return writeRateLimited(c, retryAfterSec)
	}

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

	// request bodyのbind失敗はinvalid_request 。
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	ip := clientIP(c)
	// 1:IP。
	allowed, retryAfterSec, err := ctl.rl.AllowForgotIP(ip)
	if err != nil {
		return writeErr(c, err)
	}
	if !allowed {
		return writeRateLimited(c, retryAfterSec)
	}

	// 2:email
	// emailの生値はRedisキーに使わず、hash化してから使う。
	emailHash := hashEmailForRateLimit(req.Email)

	allowed, retryAfterSec, err = ctl.rl.AllowForgotMail(emailHash)
	if err != nil {
		return writeErr(c, err)
	}
	if !allowed {
		return writeRateLimited(c, retryAfterSec)
	}

	err = ctl.uc.ForgotPw(usecase.ForgotPwIn{
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

// csrf token用のランダム値を生成。
func newCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

// csrf cookieのMaxAge秒数。
func authCSRFCookieMaxAgeSec() int {
	return 24 * 60 * 60
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

// rate limit超過時のレスポンスを返す。
// Retry-Afterが1以上ならheaderにも載せる。
func writeRateLimited(c echo.Context, retryAfterSec int) error {
	if retryAfterSec > 0 {
		c.Response().Header().Set("Retry-After", strconv.Itoa(retryAfterSec))
	}

	return c.JSON(http.StatusTooManyRequests, map[string]string{
		"error": "rate_limited",
	})
}

// loginでemailをそのままRedisキーへ使わないよう、正規化してハッシュ化。
func hashEmailForRateLimit(email string) string {
	v := strings.ToLower(strings.TrimSpace(email))
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

// キーにはhash済みの文字列だけ使う
func hashTokenForRateLimit(token string) string {
	v := strings.TrimSpace(token)
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}
