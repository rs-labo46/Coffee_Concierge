package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type SignupRateLimiter interface {
	//IP単位でsignupを許可するか判定。
	AllowSignup(ip string) (allowed bool, retryAfterSec int, err error)
}

// IPとemailの二段階制限が必要。
type LoginRateLimiter interface {
	//loginのIP。
	AllowLoginIP(ip string) (allowed bool, retryAfterSec int, err error)

	//hash済みの文字列を受け取る。
	AllowLogin(emailHash string) (allowed bool, retryAfterSec int, err error)
}

type ResendVerifyRateLimiter interface {
	// resendのIP。
	AllowResendIP(ip string) (allowed bool, retryAfterSec int, err error)

	//resendのemail。
	AllowResendMail(emailHash string) (allowed bool, retryAfterSec int, err error)
}

// password forgotのrate limit判定
type ForgotPwRateLimiter interface {
	// forgot passwordのIP。
	AllowForgotIP(ip string) (allowed bool, retryAfterSec int, err error)

	// forgot passwordのemail。
	AllowForgotMail(emailHash string) (allowed bool, retryAfterSec int, err error)
}

type RefreshRateLimiter interface {
	// refresh token単位でrefreshを許可するか判定。
	// hash化済み文字列を受け取る。
	AllowRefreshToken(tokenHash string) (allowed bool, retryAfterSec int, err error)
}

func SignupRateLimit(rl SignupRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return signupRateLimitHandler(rl, next)
	}
}

func signupRateLimitHandler(rl SignupRateLimiter, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if rl == nil {
			return next(c)
		}

		//ipを取得
		ip := normalizedRealIP(c)
		allowed, retryAfterSec, err := rl.AllowSignup(ip)
		if err != nil {
			return writeInternal(c)
		}
		if !allowed {
			return writeRateLimited(c, retryAfterSec)
		}

		return next(c)
	}
}

// IP→emailの二段階制限
func LoginRateLimit(rl LoginRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl == nil {
				return next(c)
			}

			// IPを取得。
			ip := normalizedRealIP(c)

			// IP単位の制限をかける。
			allowed, retryAfterSec, err := rl.AllowLoginIP(ip)
			if err != nil {
				return writeInternal(c)
			}
			if !allowed {
				return writeRateLimited(c, retryAfterSec)
			}

			// bodyからemailを読む。
			email, err := readEmailAndRestoreBody(c)
			if err != nil {
				return writeInvalidRequest(c)
			}

			// 2段目: email の生値はそのまま使わず、正規化して hash 化する。
			emailHash := hashEmailForRateLimit(email)

			allowed, retryAfterSec, err = rl.AllowLogin(emailHash)
			if err != nil {
				return writeInternal(c)
			}
			if !allowed {
				return writeRateLimited(c, retryAfterSec)
			}

			return next(c)
		}
	}
}

// resendもと同じくIP→emailの二段階制限。
func ResendVerifyRateLimit(rl ResendVerifyRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl == nil {
				return next(c)
			}

			// 送信元のIPを取得する。
			ip := normalizedRealIP(c)

			// IPの制限。
			allowed, retryAfterSec, err := rl.AllowResendIP(ip)
			if err != nil {
				return writeInternal(c)
			}
			if !allowed {
				return writeRateLimited(c, retryAfterSec)
			}

			// bodyからemailを読み、必ずbodyを元に戻す。
			email, err := readEmailAndRestoreBody(c)
			if err != nil {
				return writeInvalidRequest(c)
			}

			// emailをhash化して使う。
			emailHash := hashEmailForRateLimit(email)

			allowed, retryAfterSec, err = rl.AllowResendMail(emailHash)
			if err != nil {
				return writeInternal(c)
			}
			if !allowed {
				return writeRateLimited(c, retryAfterSec)
			}

			return next(c)
		}
	}
}

// forgotもIP→emailの二段階制限。
func ForgotPwRateLimit(rl ForgotPwRateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl == nil {
				return next(c)
			}

			// 送信元のIPを取得。
			ip := normalizedRealIP(c)

			// IP制限。
			allowed, retryAfterSec, err := rl.AllowForgotIP(ip)
			if err != nil {
				return writeInternal(c)
			}
			if !allowed {
				return writeRateLimited(c, retryAfterSec)
			}

			// bodyからemailを読む。
			email, err := readEmailAndRestoreBody(c)
			if err != nil {
				return writeInvalidRequest(c)
			}

			// emailをhash化してemail単位の制限へ使う。
			emailHash := hashEmailForRateLimit(email)

			allowed, retryAfterSec, err = rl.AllowForgotMail(emailHash)
			if err != nil {
				return writeInternal(c)
			}
			if !allowed {
				return writeRateLimited(c, retryAfterSec)
			}

			return next(c)
		}
	}
}

// refresh tokenはcookieを読んで、hash化して制限。
func RefreshRateLimit(rl RefreshRateLimiter, refreshCookieName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl == nil {
				return next(c)
			}
			// cookieからrefresh tokenを取得。
			cookie, err := c.Cookie(refreshCookieName)
			if err != nil {
				return writeUnauthorized(c)
			}
			tokenHash := hashTokenForRateLimit(cookie.Value)

			// refresh tokenでrate limit判定。
			allowed, retryAfterSec, err := rl.AllowRefreshToken(tokenHash)
			if err != nil {
				return writeInternal(c)
			}
			if !allowed {
				return writeRateLimited(c, retryAfterSec)
			}

			return next(c)
		}
	}
}

type emailOnlyBody struct {
	Email string `json:"email"`
}

// request bodyからemailを読み、その後、controller側でも再度Bindできるようbodyを元に戻す。
func readEmailAndRestoreBody(c echo.Context) (string, error) {
	// request自体を取得する。
	req := c.Request()

	if req.Body == nil {
		return "", io.EOF
	}

	// 一度bodyを全部読む。
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}

	// 読み終わるとbodyは空になるので、controllerでも読めるように戻す。
	req.Body = io.NopCloser(bytes.NewBuffer(raw))

	// JSONをstructにdecodeする。
	var in emailOnlyBody
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", err
	}

	// 前後空白を除去して返す。
	return strings.TrimSpace(in.Email), nil
}

// Echoから送信元IPを取り、空文字を避けるため
func normalizedRealIP(c echo.Context) string {
	// X-Forwarded-Forなどを見た上でIPを返す。
	ip := strings.TrimSpace(c.RealIP())

	// keyが空になるのを防ぐ。
	if ip == "" {
		return "unknown"
	}

	return ip
}

// emailを正規化してSHA-256でhash化する。
// email生値をRedis keyに残さない。
func hashEmailForRateLimit(email string) string {
	v := strings.ToLower(strings.TrimSpace(email))

	// SHA-256を計算。
	sum := sha256.Sum256([]byte(v))

	// 16進文字列で返す。
	return hex.EncodeToString(sum[:])
}

// tokenをtrimしてSHA-256でhash化する。
// refresh tokenをRedis keyに残さない。
func hashTokenForRateLimit(token string) string {
	v := strings.TrimSpace(token)
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

func intToString(v int) string {
	return strconv.Itoa(v)
}

func writeInvalidRequest(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, map[string]string{
		"error": "invalid_request",
	})
}
