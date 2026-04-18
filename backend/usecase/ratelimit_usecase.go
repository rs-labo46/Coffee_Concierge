package usecase

import (
	"fmt"
	"time"
)

// Token Bucket1件分のルール
type RateRule struct {
	Rate     float64
	Capacity float64
	Cost     float64
}

// usecaseはRedisの実装を知らず、repositoryに依存する。
type RateLimitStore interface {
	Allow(
		key string,
		rate float64,
		capacity float64,
		cost float64,
		now time.Time,
	) (allowed bool, retryAfterSec int, err error)
}

// controllerやmiddlewareから見たusecaseの入口でどの種類の制限かを表す。
type RateLimiter interface {
	AllowSignup(ip string) (bool, int, error)
	AllowLoginIP(ip string) (bool, int, error)
	AllowLogin(emailHash string) (bool, int, error)
	AllowRefreshToken(tokenHash string) (bool, int, error)
	AllowResendIP(ip string) (bool, int, error)
	AllowResendMail(emailHash string) (bool, int, error)
	AllowForgotIP(ip string) (bool, int, error)
	AllowForgotMail(emailHash string) (bool, int, error)
	AllowWS(key string) (bool, int, error)
}

type RateLimitUC struct {
	// usecaseは実装詳細を持たず、Allowだけ呼ぶ。
	store RateLimitStore
	// signupをIPで制限するためのルール。
	signupIP RateRule
	// loginをemailで制限するためのルール。
	loginIP   RateRule
	loginMail RateRule
	// refreshをrefreshtokenのhashで制限するためのルール。
	refreshToken RateRule

	// verify再送をIPで制限するためのルール。
	resendIP RateRule

	// verify再送をemailで制限するためのルール。
	resendMail RateRule

	// forgot passwordをIPで制限するためのルール。
	forgotIP RateRule

	// forgot passwordをemailで制限するためのルール。
	forgotMail RateRule

	// WebSocket接続を制限するためのルール。
	wsConnect RateRule
}

// rate limitのusecaseを生成する。
func NewRateLimitUC(
	store RateLimitStore,
	signupIP RateRule,
	loginIP RateRule,
	loginMail RateRule,
	refreshToken RateRule,
	resendIP RateRule,
	resendMail RateRule,
	forgotIP RateRule,
	forgotMail RateRule,
	wsConnect RateRule,
) RateLimiter {
	return &RateLimitUC{
		store:        store,
		signupIP:     signupIP,
		loginIP:      loginIP,
		loginMail:    loginMail,
		refreshToken: refreshToken,
		resendIP:     resendIP,
		resendMail:   resendMail,
		forgotIP:     forgotIP,
		forgotMail:   forgotMail,
		wsConnect:    wsConnect,
	}
}

// 判定単位はIP。
func (u *RateLimitUC) AllowSignup(ip string) (bool, int, error) {
	return u.allow("rl:signup:ip:"+ip, u.signupIP)
}

// loginの1段目。IP単位
func (u *RateLimitUC) AllowLoginIP(ip string) (bool, int, error) {
	return u.allow("rl:login:ip:"+ip, u.loginIP)
}

// 判定単位はemail側(emailのhash)。
func (u *RateLimitUC) AllowLogin(emailHash string) (bool, int, error) {
	return u.allow("rl:login:mail:"+emailHash, u.loginMail)
}

// 判定単位はrefreshのtokenをhash。
func (u *RateLimitUC) AllowRefreshToken(tokenHash string) (bool, int, error) {
	return u.allow("rl:refresh:token:"+tokenHash, u.refreshToken)
}

// 判定単位はIP。
func (u *RateLimitUC) AllowResendIP(ip string) (bool, int, error) {
	return u.allow("rl:resend:ip:"+ip, u.resendIP)
}

// 判定単位はemail
func (u *RateLimitUC) AllowResendMail(emailHash string) (bool, int, error) {
	return u.allow("rl:resend:mail:"+emailHash, u.resendMail)
}

// 判定単位はIP。
func (u *RateLimitUC) AllowForgotIP(ip string) (bool, int, error) {
	return u.allow("rl:forgot:ip:"+ip, u.forgotIP)
}

// 判定単位はemail
func (u *RateLimitUC) AllowForgotMail(emailHash string) (bool, int, error) {
	return u.allow("rl:forgot:mail:"+emailHash, u.forgotMail)
}

// WebSocket接続専用の制限判定
func (u *RateLimitUC) AllowWS(key string) (bool, int, error) {
	return u.allow(key, u.wsConnect)
}

// allow は共通処理。
func (u *RateLimitUC) allow(key string, rule RateRule) (bool, int, error) {
	if u == nil || u.store == nil {
		return false, 0, fmt.Errorf("rate limit store is nil")
	}

	if key == "" {
		return false, 0, fmt.Errorf("rate limit key is empty")
	}

	if rule.Rate <= 0 {
		return false, 0, fmt.Errorf("invalid rule rate")
	}
	if rule.Capacity <= 0 {
		return false, 0, fmt.Errorf("invalid rule capacity")
	}
	if rule.Cost <= 0 {
		return false, 0, fmt.Errorf("invalid rule cost")
	}

	allowed, retryAfterSec, err := u.store.Allow(
		key,
		rule.Rate,
		rule.Capacity,
		rule.Cost,
		time.Now(),
	)
	if err != nil {
		return false, 0, err
	}

	if allowed {
		return true, 0, nil
	}

	return false, retryAfterSec, nil
}
