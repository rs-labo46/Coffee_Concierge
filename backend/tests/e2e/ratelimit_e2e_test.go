package e2e

import (
	"net/http"
	"testing"
)

// TestE2E45_RateLimit_Login_ReturnsTooManyRequests は、login連打で429とRetry-Afterが返ることを検証する。
func TestE2E45_RateLimit_Login_ReturnsTooManyRequests(t *testing.T) {
	e2e45RequireRateLimitE2E(t)
	e2e45FlushRedis(t)
	t.Cleanup(func() { e2e45FlushRedis(t) })

	c := newAPIClient(t)
	email, _ := adminCreds()
	e2e45ExpectRateLimited(t, 20, func() (int, []byte, http.Header) {
		return c.doJSON(t, http.MethodPost, "/auth/login", map[string]string{
			"email":    email,
			"password": "WrongPass123!",
		}, nil)
	})
}

// TestE2E45_RateLimit_Signup_ReturnsTooManyRequests は、signup連打で429とRetry-Afterが返ることを検証する。
func TestE2E45_RateLimit_Signup_ReturnsTooManyRequests(t *testing.T) {
	e2e45RequireRateLimitE2E(t)
	e2e45FlushRedis(t)
	t.Cleanup(func() { e2e45FlushRedis(t) })

	c := newAPIClient(t)
	e2e45ExpectRateLimited(t, 10, func() (int, []byte, http.Header) {
		return c.doJSON(t, http.MethodPost, "/auth/signup", map[string]string{
			"email":    e2e45UniqueEmail("signup_rl"),
			"password": "ValidPass123!",
		}, nil)
	})
}

// TestE2E45_RateLimit_ForgotPassword_ReturnsTooManyRequests は、password forgot連打で429とRetry-Afterが返ることを検証する。
func TestE2E45_RateLimit_ForgotPassword_ReturnsTooManyRequests(t *testing.T) {
	e2e45RequireRateLimitE2E(t)
	e2e45FlushRedis(t)
	t.Cleanup(func() { e2e45FlushRedis(t) })

	c := newAPIClient(t)
	e2e45ExpectRateLimited(t, 10, func() (int, []byte, http.Header) {
		return c.doJSON(t, http.MethodPost, "/auth/password/forgot", map[string]string{
			"email": "forgot-rate-limit@example.com",
		}, nil)
	})
}

// TestE2E45_RateLimit_ResendVerify_ReturnsTooManyRequests は、verify resend連打で429とRetry-Afterが返ることを検証する。
func TestE2E45_RateLimit_ResendVerify_ReturnsTooManyRequests(t *testing.T) {
	e2e45RequireRateLimitE2E(t)
	e2e45FlushRedis(t)
	t.Cleanup(func() { e2e45FlushRedis(t) })

	c := newAPIClient(t)
	e2e45ExpectRateLimited(t, 10, func() (int, []byte, http.Header) {
		return c.doJSON(t, http.MethodPost, "/auth/verify-email/resend", map[string]string{
			"email": "resend-rate-limit@example.com",
		}, nil)
	})
}
