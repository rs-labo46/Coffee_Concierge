package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

const (
	// guest再開やguest操作で使うヘッダ名。
	HeaderSessionKey = "X-Session-Key"

	// middlewareがecho.Contextにactorを入れるときのキー
	ContextActorKey = "actor"
)

// 成功時の簡易メッセージレスポンス
type MsgRes struct {
	Message string `json:"message"`
}

// middlewareがcontextに入れたactorを取り出す。
func actorFromCtx(c echo.Context) *entity.Actor {
	v := c.Get(ContextActorKey)
	if v == nil {
		return nil
	}

	if a, ok := v.(*entity.Actor); ok {
		if a == nil {
			return nil
		}
		return a
	}

	if a, ok := v.(entity.Actor); ok {
		cp := a
		return &cp
	}

	return nil
}

// 認証必須
func requireActor(c echo.Context) (*entity.Actor, error) {
	a := actorFromCtx(c)
	if a == nil {
		return nil, usecase.ErrUnauthorized
	}
	return a, nil
}

// X-Session-Keyを取り出す(空白は空文字)
func sessionKeyFromHeader(c echo.Context) string {
	v := strings.TrimSpace(c.Request().Header.Get(HeaderSessionKey))
	if v == "" {
		return ""
	}
	return v
}

// guest必須
func requireSessionKey(c echo.Context) (string, error) {
	key := sessionKeyFromHeader(c)
	if key == "" {
		return "", ErrInvalidRequest
	}
	return key, nil
}

// path paramをuintに変換。
func pUint(c echo.Context, name string) (uint, error) {
	raw := strings.TrimSpace(c.Param(name))
	if raw == "" {
		return 0, ErrInvalidRequest
	}

	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, ErrInvalidRequest
	}
	if n == 0 {
		return 0, ErrInvalidRequest
	}

	return uint(n), nil
}

// query stringの整数値を取得。
func qInt(c echo.Context, name string, def int) (int, error) {
	raw := strings.TrimSpace(c.QueryParam(name))
	if raw == "" {
		return def, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, ErrInvalidRequest
	}

	return n, nil
}

// query stringのuintを取得。
func qUint(c echo.Context, name string) (*uint, error) {
	raw := strings.TrimSpace(c.QueryParam(name))
	if raw == "" {
		return nil, nil
	}

	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return nil, ErrInvalidRequest
	}
	if n == 0 {
		return nil, ErrInvalidRequest
	}

	v := uint(n)
	return &v, nil
}

// query stringのboolを取得。
func qBool(c echo.Context, name string) (*bool, error) {
	raw := strings.TrimSpace(c.QueryParam(name))
	if raw == "" {
		return nil, nil
	}

	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, ErrInvalidRequest
	}

	return &v, nil
}

// 未指定なら空文字
func qStr(c echo.Context, name string) string {
	return strings.TrimSpace(c.QueryParam(name))
}

// アクセス元のIPを返す。
func clientIP(c echo.Context) string {
	return strings.TrimSpace(c.RealIP())
}

// User-Agentを返す。
func userAgent(c echo.Context) string {
	return strings.TrimSpace(c.Request().UserAgent())
}

// キャッシュさせたくないレスポンスの
func setNoStore(c echo.Context) {
	h := c.Response().Header()
	h.Set(echo.HeaderCacheControl, "no-store")
	h.Set("Pragma", "no-cache")
	h.Set("Expires", "0")
}

// 共通cookie設定の
func setCookie(
	c echo.Context,
	name string,
	value string,
	maxAgeSec int,
	httpOnly bool,
	secure bool,
	sameSite http.SameSite,
	path string,
	domain string,
) {
	ck := new(http.Cookie)
	ck.Name = name
	ck.Value = value
	ck.MaxAge = maxAgeSec

	if path == "" {
		ck.Path = "/"
	} else {
		ck.Path = path
	}

	if domain != "" {
		ck.Domain = domain
	}

	ck.HttpOnly = httpOnly
	ck.Secure = secure
	ck.SameSite = sameSite

	if maxAgeSec > 0 {
		ck.Expires = time.Now().Add(time.Duration(maxAgeSec) * time.Second)
	}

	http.SetCookie(c.Response(), ck)
}

// cookieの削除 (PathとDomainはset時と揃える)
func clearCookie(
	c echo.Context,
	name string,
	path string,
	domain string,
	httpOnly bool,
	secure bool,
	sameSite http.SameSite,
) {
	ck := new(http.Cookie)
	ck.Name = name
	ck.Value = ""
	ck.MaxAge = -1
	ck.Expires = time.Unix(0, 0)

	if path == "" {
		ck.Path = "/"
	} else {
		ck.Path = path
	}

	if domain != "" {
		ck.Domain = domain
	}

	ck.HttpOnly = httpOnly
	ck.Secure = secure
	ck.SameSite = sameSite

	http.SetCookie(c.Response(), ck)
}

// request cookieを読む。
func cookieValue(c echo.Context, name string) string {
	ck, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	if ck == nil {
		return ""
	}
	return strings.TrimSpace(ck.Value)
}

// refresh token
func setRefreshCookie(
	c echo.Context,
	cookieName string,
	token string,
	maxAgeSec int,
	secure bool,
	sameSite http.SameSite,
	path string,
	domain string,
) {
	setCookie(
		c,
		cookieName,
		token,
		maxAgeSec,
		true,
		secure,
		sameSite,
		path,
		domain,
	)
}

// refresh token用削除
func clearRefreshCookie(
	c echo.Context,
	cookieName string,
	secure bool,
	sameSite http.SameSite,
	path string,
	domain string,
) {
	clearCookie(
		c,
		cookieName,
		path,
		domain,
		true,
		secure,
		sameSite,
	)
}
