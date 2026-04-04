package usecase

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

func TestAuthUCForgotPw_UserNotFound_ReturnsNilAndAudits(t *testing.T) {
	t.Parallel()

	var gotAuditType string

	uc := &AuthUC{
		user: &authUserRepoMock{
			getByEmailFn: func(email string) (entity.User, error) {
				return entity.User{}, repository.ErrNotFound
			},
		},
		ev: &authEvRepoMock{},
		pw: &authPwRepoMock{},
		rt: &authRtRepoMock{},
		audit: &authAuditRepoMock{
			createFn: func(a entity.AuditLog) error {
				gotAuditType = a.Type
				return nil
			},
		},
		val: &authValMock{
			emailOnlyFn: func(email string) error { return nil },
		},
		ph:   &pwHashMock{},
		tk:   &tokMock{},
		mail: &mailerMock{},
		rl: &rateLimMock{
			allowForgotIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowForgotMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	err := uc.ForgotPw(ForgotPwIn{Email: "none@example.com", IP: "127.0.0.1", UA: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuditType != "auth.password.forgot" {
		t.Fatalf("audit type = %q, want auth.password.forgot", gotAuditType)
	}
}

func TestAuthUCResetPw_OK_BumpsAndRevokesAll(t *testing.T) {
	t.Parallel()

	var updatedUserID int64
	var updatedHash string
	var usedID int64
	var bumpedUserID int64
	var revokedAllUserID int64
	var gotAuditType string

	uc := &AuthUC{
		user: &authUserRepoMock{
			createFn:           func(u entity.User) (entity.User, error) { return entity.User{}, nil },
			getByEmailFn:       func(email string) (entity.User, error) { return entity.User{}, nil },
			getByIDFn:          func(id int64) (entity.User, error) { return entity.User{}, nil },
			setEmailVerifiedFn: func(userID int64) error { return nil },
			updatePassHashFn: func(userID int64, newHash string) error {
				updatedUserID = userID
				updatedHash = newHash
				return nil
			},
			bumpTokenVerFn: func(userID int64) (int, error) {
				bumpedUserID = userID
				return 2, nil
			},
		},
		ev: &authEvRepoMock{},
		pw: &authPwRepoMock{
			createFn: func(pw entity.PwReset) error { return nil },
			getByTokenHashFn: func(hash string) (entity.PwReset, error) {
				if hash != sha256Hex("reset-raw") {
					t.Fatalf("unexpected token hash: %q", hash)
				}
				return entity.PwReset{
					ID:        40,
					UserID:    12,
					ExpiresAt: time.Now().Add(3 * time.Minute),
				}, nil
			},
			useFn:              func(id int64) error { usedID = id; return nil },
			revokeUnusedByUser: func(userID int64) error { return nil },
		},
		rt: &authRtRepoMock{
			createFn:         func(rt entity.RefreshToken) (entity.RefreshToken, error) { return entity.RefreshToken{}, nil },
			getByTokenHashFn: func(hash string) (entity.RefreshToken, error) { return entity.RefreshToken{}, nil },
			revokeFn:         func(id int64) error { return nil },
			markUsedFn:       func(id int64) error { return nil },
			setReplacedByFn:  func(id int64, newID int64) error { return nil },
			revokeByFamilyFn: func(familyID string) error { return nil },
			revokeAllByUserFn: func(userID int64) error {
				revokedAllUserID = userID
				return nil
			},
		},
		audit: &authAuditRepoMock{
			createFn: func(a entity.AuditLog) error {
				gotAuditType = a.Type
				return nil
			},
		},
		val: &authValMock{
			newPwFn: func(pw string) error { return nil },
		},
		ph: &pwHashMock{
			hashFn:    func(pw string) (string, error) { return "new-hash", nil },
			compareFn: func(hash string, pw string) error { return nil },
		},
		tk:   &tokMock{},
		mail: &mailerMock{},
		rl:   &rateLimMock{},
	}

	err := uc.ResetPw(ResetPwIn{Token: "reset-raw", NewPw: "NewPW123!@#X", IP: "127.0.0.1", UA: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedUserID != 12 || updatedHash != "new-hash" {
		t.Fatalf("update pass hash args = (%d,%q)", updatedUserID, updatedHash)
	}
	if usedID != 40 {
		t.Fatalf("used id = %d, want 40", usedID)
	}
	if bumpedUserID != 12 {
		t.Fatalf("bumped user id = %d, want 12", bumpedUserID)
	}
	if revokedAllUserID != 12 {
		t.Fatalf("revoke all user id = %d, want 12", revokedAllUserID)
	}
	if gotAuditType != "auth.password.reset" {
		t.Fatalf("audit type = %q, want auth.password.reset", gotAuditType)
	}
}

func TestAuthUCForgotPw_RateLimited_ReturnsNil(t *testing.T) {
	t.Parallel()

	uc := &AuthUC{
		user:  &authUserRepoMock{},
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
			allowForgotIPFn:   func(ip string) (bool, int, error) { return false, 2, nil },
			allowForgotMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	if err := uc.ForgotPw(ForgotPwIn{Email: "test@example.com", IP: "127.0.0.1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthUCResetPw_UsedToken_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	usedAt := time.Now().Add(-1 * time.Minute)

	uc := &AuthUC{
		user: &authUserRepoMock{},
		ev:   &authEvRepoMock{},
		pw: &authPwRepoMock{
			getByTokenHashFn: func(hash string) (entity.PwReset, error) {
				return entity.PwReset{
					ID:        1,
					UserID:    2,
					ExpiresAt: time.Now().Add(3 * time.Minute),
					UsedAt:    &usedAt,
				}, nil
			},
		},
		rt:    &authRtRepoMock{},
		audit: &authAuditRepoMock{createFn: func(a entity.AuditLog) error { return nil }},
		val: &authValMock{
			newPwFn: func(pw string) error { return nil },
		},
		ph:   &pwHashMock{},
		tk:   &tokMock{},
		mail: &mailerMock{},
		rl:   &rateLimMock{},
	}

	err := uc.ResetPw(ResetPwIn{Token: "used-token", NewPw: "NewPW123!@#X"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}
