package usecase

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

func TestAuthUCSignup_OK(t *testing.T) {
	t.Parallel()

	var gotAuditType string
	var gotVerifyEmail string
	var gotVerifyToken string

	uc := &AuthUC{
		user: &authUserRepoMock{
			createFn: func(u entity.User) (entity.User, error) {
				if u.Email != "test@example.com" {
					t.Fatalf("email = %q, want test@example.com", u.Email)
				}
				if u.Role != string(entity.RoleUser) {
					t.Fatalf("role = %q, want user", u.Role)
				}
				if u.PassHash != "hashed-pw" {
					t.Fatalf("pass hash = %q, want hashed-pw", u.PassHash)
				}
				u.ID = 7
				return u, nil
			},
			getByEmailFn:       func(email string) (entity.User, error) { return entity.User{}, nil },
			getByIDFn:          func(id int64) (entity.User, error) { return entity.User{}, nil },
			setEmailVerifiedFn: func(userID int64) error { return nil },
			updatePassHashFn:   func(userID int64, newHash string) error { return nil },
			bumpTokenVerFn:     func(userID int64) (int, error) { return 0, nil },
		},
		ev: &authEvRepoMock{
			createFn: func(ev entity.EmailVerify) error {
				if ev.UserID != 7 {
					t.Fatalf("userID = %d, want 7", ev.UserID)
				}
				if ev.TokenHash != sha256Hex("verify-raw-token") {
					t.Fatalf("unexpected token hash: %q", ev.TokenHash)
				}
				if time.Until(ev.ExpiresAt) <= 0 {
					t.Fatal("expires_at should be in the future")
				}
				return nil
			},
			getByTokenHashFn:   func(hash string) (entity.EmailVerify, error) { return entity.EmailVerify{}, nil },
			useFn:              func(id int64) error { return nil },
			revokeUnusedByUser: func(userID int64) error { return nil },
		},
		pw: &authPwRepoMock{
			createFn:           func(pw entity.PwReset) error { return nil },
			getByTokenHashFn:   func(hash string) (entity.PwReset, error) { return entity.PwReset{}, nil },
			useFn:              func(id int64) error { return nil },
			revokeUnusedByUser: func(userID int64) error { return nil },
		},
		rt: &authRtRepoMock{
			createFn:          func(rt entity.Rt) (entity.Rt, error) { return entity.Rt{}, nil },
			getByTokenHashFn:  func(hash string) (entity.Rt, error) { return entity.Rt{}, nil },
			revokeFn:          func(id int64) error { return nil },
			markUsedFn:        func(id int64) error { return nil },
			setReplacedByFn:   func(id int64, newID int64) error { return nil },
			revokeByFamilyFn:  func(familyID string) error { return nil },
			revokeAllByUserFn: func(userID int64) error { return nil },
		},
		audit: &authAuditRepoMock{
			createFn: func(a entity.AuditLog) error {
				gotAuditType = a.Type
				return nil
			},
		},
		val: &authValMock{
			signupFn:    func(email string, pw string) error { return nil },
			loginFn:     func(email string, pw string) error { return nil },
			emailOnlyFn: func(email string) error { return nil },
			newPwFn:     func(pw string) error { return nil },
		},
		ph: &pwHashMock{
			hashFn:    func(pw string) (string, error) { return "hashed-pw", nil },
			compareFn: func(hash string, pw string) error { return nil },
		},
		tk: &tokMock{
			newAccessFn:   func(userID int64, role string, tokenVer int) (string, error) { return "", nil },
			newCSRFFn:     func() (string, error) { return "", nil },
			newOpaqueFn:   func() (string, error) { return "verify-raw-token", nil },
			newFamilyIDFn: func() (string, error) { return "", nil },
		},
		mail: &mailerMock{
			sendVerifyFn: func(email string, token string) error {
				gotVerifyEmail = email
				gotVerifyToken = token
				return nil
			},
			sendResetFn: func(email string, token string) error { return nil },
		},
		rl: &rateLimMock{
			allowSignupFn:     func(ip string) (bool, int, error) { return true, 0, nil },
			allowLoginFn:      func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowRefreshFn:    func(userID int64) (bool, int, error) { return true, 0, nil },
			allowResendIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowResendMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowForgotIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowForgotMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	got, err := uc.Signup(SignupIn{
		Email: "  Test@Example.com ",
		Pw:    "CorrectPW123!",
		IP:    "127.0.0.1",
		UA:    "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 7 {
		t.Fatalf("user id = %d, want 7", got.ID)
	}
	if gotVerifyEmail != "test@example.com" {
		t.Fatalf("verify email = %q, want test@example.com", gotVerifyEmail)
	}
	if gotVerifyToken != "verify-raw-token" {
		t.Fatalf("verify token = %q, want verify-raw-token", gotVerifyToken)
	}
	if gotAuditType != "auth.signup" {
		t.Fatalf("audit type = %q, want auth.signup", gotAuditType)
	}
}

func TestAuthUCSignup_RateLimited(t *testing.T) {
	t.Parallel()

	uc := &AuthUC{
		user:  &authUserRepoMock{},
		ev:    &authEvRepoMock{},
		pw:    &authPwRepoMock{},
		rt:    &authRtRepoMock{},
		audit: &authAuditRepoMock{createFn: func(a entity.AuditLog) error { return nil }},
		val: &authValMock{
			signupFn:    func(email string, pw string) error { return nil },
			loginFn:     func(email string, pw string) error { return nil },
			emailOnlyFn: func(email string) error { return nil },
			newPwFn:     func(pw string) error { return nil },
		},
		ph:   &pwHashMock{},
		tk:   &tokMock{},
		mail: &mailerMock{},
		rl: &rateLimMock{
			allowSignupFn:     func(ip string) (bool, int, error) { return false, 5, nil },
			allowLoginFn:      func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowRefreshFn:    func(userID int64) (bool, int, error) { return true, 0, nil },
			allowResendIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowResendMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowForgotIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowForgotMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	_, err := uc.Signup(SignupIn{Email: "a@test.com", Pw: "CorrectPW123!", IP: "127.0.0.1"})
	var rl ErrRateLimited
	if !errors.As(err, &rl) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
	if rl.RetryAfterSec != 5 {
		t.Fatalf("retry_after = %d, want 5", rl.RetryAfterSec)
	}
}

func TestAuthUCVerifyEmail_OK(t *testing.T) {
	t.Parallel()

	var setVerifiedID int64
	var usedID int64
	var gotAuditType string

	uc := &AuthUC{
		user: &authUserRepoMock{
			createFn:           func(u entity.User) (entity.User, error) { return entity.User{}, nil },
			getByEmailFn:       func(email string) (entity.User, error) { return entity.User{}, nil },
			getByIDFn:          func(id int64) (entity.User, error) { return entity.User{}, nil },
			setEmailVerifiedFn: func(userID int64) error { setVerifiedID = userID; return nil },
			updatePassHashFn:   func(userID int64, newHash string) error { return nil },
			bumpTokenVerFn:     func(userID int64) (int, error) { return 0, nil },
		},
		ev: &authEvRepoMock{
			createFn: func(ev entity.EmailVerify) error { return nil },
			getByTokenHashFn: func(hash string) (entity.EmailVerify, error) {
				if hash != sha256Hex("verify-token") {
					t.Fatalf("unexpected hash: %q", hash)
				}
				return entity.EmailVerify{
					ID:        11,
					UserID:    9,
					ExpiresAt: time.Now().Add(3 * time.Minute),
				}, nil
			},
			useFn:              func(id int64) error { usedID = id; return nil },
			revokeUnusedByUser: func(userID int64) error { return nil },
		},
		pw: &authPwRepoMock{},
		rt: &authRtRepoMock{},
		audit: &authAuditRepoMock{
			createFn: func(a entity.AuditLog) error {
				gotAuditType = a.Type
				return nil
			},
		},
		val:  &authValMock{},
		ph:   &pwHashMock{},
		tk:   &tokMock{},
		mail: &mailerMock{},
		rl:   &rateLimMock{},
	}

	err := uc.VerifyEmail(VerifyEmailIn{Token: "verify-token", IP: "127.0.0.1", UA: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if setVerifiedID != 9 {
		t.Fatalf("set email verified id = %d, want 9", setVerifiedID)
	}
	if usedID != 11 {
		t.Fatalf("used id = %d, want 11", usedID)
	}
	if gotAuditType != "auth.verify_email" {
		t.Fatalf("audit type = %q, want auth.verify_email", gotAuditType)
	}
}

func TestAuthUCResendVerify_RateLimited_ReturnsNilAndAudits(t *testing.T) {
	t.Parallel()

	var gotAuditType string

	uc := &AuthUC{
		user: &authUserRepoMock{},
		ev:   &authEvRepoMock{},
		pw:   &authPwRepoMock{},
		rt:   &authRtRepoMock{},
		audit: &authAuditRepoMock{
			createFn: func(a entity.AuditLog) error {
				gotAuditType = a.Type
				return nil
			},
		},
		val: &authValMock{
			signupFn:    func(email string, pw string) error { return nil },
			loginFn:     func(email string, pw string) error { return nil },
			emailOnlyFn: func(email string) error { return nil },
			newPwFn:     func(pw string) error { return nil },
		},
		ph:   &pwHashMock{},
		tk:   &tokMock{},
		mail: &mailerMock{},
		rl: &rateLimMock{
			allowSignupFn:     func(ip string) (bool, int, error) { return true, 0, nil },
			allowLoginFn:      func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowRefreshFn:    func(userID int64) (bool, int, error) { return true, 0, nil },
			allowResendIPFn:   func(ip string) (bool, int, error) { return false, 2, nil },
			allowResendMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowForgotIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowForgotMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	err := uc.ResendVerify(ResendVerifyIn{Email: "test@example.com", IP: "127.0.0.1", UA: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuditType != "auth.email.resend.rate_limited" {
		t.Fatalf("audit type = %q, want auth.email.resend.rate_limited", gotAuditType)
	}
}

func TestAuthUCVerifyEmail_Expired_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	uc := &AuthUC{
		user: &authUserRepoMock{},
		ev: &authEvRepoMock{
			createFn: func(ev entity.EmailVerify) error { return nil },
			getByTokenHashFn: func(hash string) (entity.EmailVerify, error) {
				return entity.EmailVerify{
					ID:        1,
					UserID:    2,
					ExpiresAt: time.Now().Add(-1 * time.Minute),
				}, nil
			},
			useFn:              func(id int64) error { return nil },
			revokeUnusedByUser: func(userID int64) error { return nil },
		},
		pw:    &authPwRepoMock{},
		rt:    &authRtRepoMock{},
		audit: &authAuditRepoMock{createFn: func(a entity.AuditLog) error { return nil }},
		val:   &authValMock{},
		ph:    &pwHashMock{},
		tk:    &tokMock{},
		mail:  &mailerMock{},
		rl:    &rateLimMock{},
	}

	err := uc.VerifyEmail(VerifyEmailIn{Token: "expired"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestAuthUCResendVerify_UserNotFound_ReturnsNil(t *testing.T) {
	t.Parallel()

	uc := &AuthUC{
		user: &authUserRepoMock{
			getByEmailFn: func(email string) (entity.User, error) {
				return entity.User{}, repository.ErrNotFound
			},
		},
		ev:    &authEvRepoMock{},
		pw:    &authPwRepoMock{},
		rt:    &authRtRepoMock{},
		audit: &authAuditRepoMock{createFn: func(a entity.AuditLog) error { return nil }},
		val: &authValMock{
			emailOnlyFn: func(email string) error { return nil },
		},
		ph:   &pwHashMock{},
		tk:   &tokMock{},
		mail: &mailerMock{},
		rl: &rateLimMock{
			allowResendIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowResendMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	if err := uc.ResendVerify(ResendVerifyIn{Email: "none@example.com"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
