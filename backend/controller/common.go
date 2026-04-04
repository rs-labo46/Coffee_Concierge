package controller

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type ErrRes struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// middlewareがcontextに入れたuser_id / roleをusecaseが受け取れるActorに変換。
func actorFromCtx(c echo.Context) usecase.Actor {
	userID, _ := c.Get("user_id").(int64)
	role, _ := c.Get("role").(string)

	return usecase.Actor{
		UserID: userID,
		Role:   role,
		IP:     realIP(c),
		UA:     userAgent(c),
	}
}

// middlewareがcontextに入れたuser_idを取り出す。
func userIDFromCtx(c echo.Context) int64 {
	userID, _ := c.Get("user_id").(int64)
	return userID
}

// リクエスト元IPを返す。
func realIP(c echo.Context) string {
	x := strings.TrimSpace(c.RealIP())
	if x != "" {
		return x
	}
	return ""
}

// User-Agentを返す。
func userAgent(c echo.Context) string {
	return c.Request().UserAgent()
}

func bindJSON(c echo.Context, dst interface{}) error {
	if err := c.Bind(dst); err != nil {
		return usecase.ErrInvalidRequest
	}
	return nil
}

// query parameterをintに変換。
// 値が無いときはdefを返し、不正値ならErrInvalidRequestを返す。
func qInt(c echo.Context, key string, def int) (int, error) {
	s := strings.TrimSpace(c.QueryParam(key))
	if s == "" {
		return def, nil
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, usecase.ErrInvalidRequest
	}

	return n, nil
}

// cookieにSecureを付けるかを環境変数から判定。
func secureCookie() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE")))
	return v == "1" || v == "true" || v == "yes"
}

// refresh_token cookieをセット。
// Pathは/authに固定。
func setRefreshCookie(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   secureCookie(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   24 * 60 * 60,
	})
}

// csrf_token cookieをセット。
// Pathは/に固定。
func setCSRFCookie(c echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false,
		Secure:   secureCookie(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   24 * 60 * 60,
	})
}

// refresh_token cookieを削除。
func clearRefreshCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		Secure:   secureCookie(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   0,
		Expires:  time.Unix(0, 0),
	})
}

// csrf_token cookieを削除。
func clearCSRFCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   secureCookie(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   0,
		Expires:  time.Unix(0, 0),
	})
}

// refresh_token cookieの値を返す。
// cookieが無ければ空文字を返す。
func refreshCookie(c echo.Context) string {
	x, err := c.Cookie("refresh_token")
	if err != nil {
		return ""
	}
	return x.Value
}

// usecaseのエラーをHTTPステータスへ変換。
func writeErr(c echo.Context, err error) error {
	if err == nil {
		return nil
	}

	var rl usecase.ErrRateLimited

	switch {
	case errors.Is(err, usecase.ErrInvalidRequest):
		return c.JSON(http.StatusBadRequest, ErrRes{
			Error:   "invalid_request",
			Message: "request body or query is invalid",
		})

	case errors.Is(err, usecase.ErrUnauthorized):
		return c.JSON(http.StatusUnauthorized, ErrRes{
			Error:   "unauthorized",
			Message: "authentication failed",
		})

	case errors.Is(err, usecase.ErrForbidden):
		return c.JSON(http.StatusForbidden, ErrRes{
			Error:   "forbidden",
			Message: "permission denied",
		})

	case errors.Is(err, usecase.ErrNotFound):
		return c.JSON(http.StatusNotFound, ErrRes{
			Error:   "not_found",
			Message: "resource not found",
		})

	case errors.Is(err, usecase.ErrConflict):
		return c.JSON(http.StatusConflict, ErrRes{
			Error:   "conflict",
			Message: "すでに登録されています。",
		})

	case errors.As(err, &rl):
		c.Response().Header().Set("Retry-After", strconv.Itoa(rl.RetryAfterSec))
		return c.JSON(http.StatusTooManyRequests, ErrRes{
			Error:   "rate_limited",
			Message: "too many requests",
		})

	default:
		return c.JSON(http.StatusInternalServerError, ErrRes{
			Error:   "internal",
			Message: "internal server error",
		})
	}
}
