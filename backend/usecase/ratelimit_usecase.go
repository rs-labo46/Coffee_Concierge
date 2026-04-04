package usecase

import (
	"fmt"
	"time"
)

// ratelimit1件分のルール
type RateRule struct {
	Limit  int64
	Window time.Duration
}

type RateLimitStore interface {
	Hit(key string, windowSec int64) (count int64, ttlSec int64, err error)
}

type RateLimiter interface {
	AllowSignup(ip string) (bool, int, error)
	AllowLogin(emailHash string) (bool, int, error)
	AllowRefresh(userID int64) (bool, int, error)
	AllowResendIP(ip string) (bool, int, error)
	AllowResendMail(emailHash string) (bool, int, error)
	AllowForgotIP(ip string) (bool, int, error)
	AllowForgotMail(emailHash string) (bool, int, error)
}

type RateLimitUC struct {
	store      RateLimitStore
	signupIP   RateRule
	loginMail  RateRule
	refreshUID RateRule
	resendIP   RateRule
	resendMail RateRule
	forgotIP   RateRule
	forgotMail RateRule
}

func NewRateLimitUC(
	store RateLimitStore,
	signupIP RateRule,
	loginMail RateRule,
	refreshUID RateRule,
	resendIP RateRule,
	resendMail RateRule,
	forgotIP RateRule,
	forgotMail RateRule,
) RateLimiter {
	return &RateLimitUC{
		store:      store,
		signupIP:   signupIP,
		loginMail:  loginMail,
		refreshUID: refreshUID,
		resendIP:   resendIP,
		resendMail: resendMail,
		forgotIP:   forgotIP,
		forgotMail: forgotMail,
	}
}

func (u *RateLimitUC) AllowSignup(ip string) (bool, int, error) {
	return u.allow("rl:signup:ip:"+ip, u.signupIP)
}

func (u *RateLimitUC) AllowLogin(emailHash string) (bool, int, error) {
	return u.allow("rl:login:mail:"+emailHash, u.loginMail)
}

func (u *RateLimitUC) AllowRefresh(userID int64) (bool, int, error) {
	return u.allow(fmt.Sprintf("rl:refresh:uid:%d", userID), u.refreshUID)
}

func (u *RateLimitUC) AllowResendIP(ip string) (bool, int, error) {
	return u.allow("rl:resend:ip:"+ip, u.resendIP)
}

func (u *RateLimitUC) AllowResendMail(emailHash string) (bool, int, error) {
	return u.allow("rl:resend:mail:"+emailHash, u.resendMail)
}

func (u *RateLimitUC) AllowForgotIP(ip string) (bool, int, error) {
	return u.allow("rl:forgot:ip:"+ip, u.forgotIP)
}

func (u *RateLimitUC) AllowForgotMail(emailHash string) (bool, int, error) {
	return u.allow("rl:forgot:mail:"+emailHash, u.forgotMail)
}

func (u *RateLimitUC) allow(key string, rule RateRule) (bool, int, error) {
	sec := int64(rule.Window / time.Second)
	if sec <= 0 {
		sec = 1
	}

	count, ttl, err := u.store.Hit(key, sec)
	if err != nil {
		return false, 0, err
	}

	if count <= rule.Limit {
		return true, 0, nil
	}

	return false, int(ttl), nil
}
