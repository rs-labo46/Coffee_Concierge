package usecase

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

func TestAuthUCLogin_OK(t *testing.T) {
	t.Parallel()

	var gotAuditType string

	uc := &AuthUC{
		user: &authUserRepoMock{
			createFn: func(u entity.User) (entity.User, error) { return entity.User{}, nil },
			getByEmailFn: func(email string) (entity.User, error) {
				if email != "login@example.com" {
					t.Fatalf("email = %q, want login@example.com", email)
				}
				return entity.User{
					ID:            3,
					Email:         email,
					PassHash:      "stored-hash",
					Role:          "user",
					TokenVer:      4,
					EmailVerified: true,
				}, nil
			},
			getByIDFn:          func(id int64) (entity.User, error) { return entity.User{}, nil },
			setEmailVerifiedFn: func(userID int64) error { return nil },
			updatePassHashFn:   func(userID int64, newHash string) error { return nil },
			bumpTokenVerFn:     func(userID int64) (int, error) { return 0, nil },
		},
		ev: &authEvRepoMock{},
		pw: &authPwRepoMock{},
		rt: &authRtRepoMock{
			createFn: func(rt entity.RefreshToken) (entity.RefreshToken, error) {
				if rt.UserID != 3 {
					t.Fatalf("user_id = %d, want 3", rt.UserID)
				}
				if rt.FamilyID != "family-1" {
					t.Fatalf("family_id = %q, want family-1", rt.FamilyID)
				}
				if rt.TokenHash != sha256Hex("refresh-raw") {
					t.Fatalf("unexpected token hash: %q", rt.TokenHash)
				}
				rt.ID = 10
				return rt, nil
			},
			getByTokenHashFn:  func(hash string) (entity.RefreshToken, error) { return entity.RefreshToken{}, nil },
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
			loginFn: func(email string, pw string) error { return nil },
		},
		ph: &pwHashMock{
			hashFn: func(pw string) (string, error) { return "", nil },
			compareFn: func(hash string, pw string) error {
				if hash != "stored-hash" || pw != "CorrectPW123!" {
					t.Fatalf("unexpected compare args: hash=%q pw=%q", hash, pw)
				}
				return nil
			},
		},
		tk: &tokMock{
			newAccessFn: func(userID int64, role string, tokenVer int) (string, error) {
				if userID != 3 || role != "user" || tokenVer != 4 {
					t.Fatalf("unexpected access args: %d %s %d", userID, role, tokenVer)
				}
				return "access-1", nil
			},
			newCSRFFn:     func() (string, error) { return "csrf-1", nil },
			newOpaqueFn:   func() (string, error) { return "refresh-raw", nil },
			newFamilyIDFn: func() (string, error) { return "family-1", nil },
		},
		mail: &mailerMock{},
		rl: &rateLimMock{
			allowLoginFn: func(emailHash string) (bool, int, error) {
				if emailHash == "" {
					t.Fatal("email hash should not be empty")
				}
				return true, 0, nil
			},
			allowSignupFn:     func(ip string) (bool, int, error) { return true, 0, nil },
			allowRefreshFn:    func(userID int64) (bool, int, error) { return true, 0, nil },
			allowResendIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowResendMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowForgotIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowForgotMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	out, err := uc.Login(LoginIn{Email: " Login@Example.com ", Pw: "CorrectPW123!", IP: "127.0.0.1", UA: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.AccessToken != "access-1" || out.RefreshToken != "refresh-raw" || out.CsrfToken != "csrf-1" {
		t.Fatalf("unexpected auth out: %+v", out)
	}
	if gotAuditType != "auth.login.success" {
		t.Fatalf("audit type = %q, want auth.login.success", gotAuditType)
	}
}

func TestAuthUCRefresh_OK_Rotates(t *testing.T) {
	t.Parallel()

	calls := make([]string, 0, 6)

	uc := &AuthUC{
		user: &authUserRepoMock{
			createFn:     func(u entity.User) (entity.User, error) { return entity.User{}, nil },
			getByEmailFn: func(email string) (entity.User, error) { return entity.User{}, nil },
			getByIDFn: func(id int64) (entity.User, error) {
				return entity.User{ID: 8, Role: "user", TokenVer: 2}, nil
			},
			setEmailVerifiedFn: func(userID int64) error { return nil },
			updatePassHashFn:   func(userID int64, newHash string) error { return nil },
			bumpTokenVerFn:     func(userID int64) (int, error) { return 0, nil },
		},
		ev: &authEvRepoMock{},
		pw: &authPwRepoMock{},
		rt: &authRtRepoMock{
			createFn: func(rt entity.RefreshToken) (entity.RefreshToken, error) {
				calls = append(calls, "create")
				rt.ID = 30
				return rt, nil
			},
			getByTokenHashFn: func(hash string) (entity.RefreshToken, error) {
				return entity.RefreshToken{
					ID:        20,
					UserID:    8,
					FamilyID:  "fam-1",
					ExpiresAt: time.Now().Add(10 * time.Minute),
				}, nil
			},
			revokeFn: func(id int64) error {
				calls = append(calls, "revoke")
				return nil
			},
			markUsedFn: func(id int64) error {
				calls = append(calls, "mark_used")
				return nil
			},
			setReplacedByFn: func(id int64, newID int64) error {
				calls = append(calls, "set_replaced_by")
				if id != 20 || newID != 30 {
					t.Fatalf("unexpected set replaced by args: %d %d", id, newID)
				}
				return nil
			},
			revokeByFamilyFn:  func(familyID string) error { return nil },
			revokeAllByUserFn: func(userID int64) error { return nil },
		},
		audit: &authAuditRepoMock{createFn: func(a entity.AuditLog) error { return nil }},
		val:   &authValMock{},
		ph:    &pwHashMock{},
		tk: &tokMock{
			newAccessFn:   func(userID int64, role string, tokenVer int) (string, error) { return "access-2", nil },
			newCSRFFn:     func() (string, error) { return "csrf-2", nil },
			newOpaqueFn:   func() (string, error) { return "refresh-2", nil },
			newFamilyIDFn: func() (string, error) { return "", nil },
		},
		mail: &mailerMock{},
		rl: &rateLimMock{
			allowRefreshFn:    func(userID int64) (bool, int, error) { return true, 0, nil },
			allowSignupFn:     func(ip string) (bool, int, error) { return true, 0, nil },
			allowLoginFn:      func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowResendIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowResendMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowForgotIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowForgotMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	out, err := uc.Refresh(RefreshIn{RefreshToken: "refresh-1", IP: "127.0.0.1", UA: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.AccessToken != "access-2" || out.RefreshToken != "refresh-2" || out.CsrfToken != "csrf-2" {
		t.Fatalf("unexpected refresh out: %+v", out)
	}

	want := []string{"create", "mark_used", "revoke", "set_replaced_by"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestAuthUCRefresh_UsedToken_DetectsReuse(t *testing.T) {
	t.Parallel()

	var revokedFamily string
	var bumpedUserID int64

	usedAt := time.Now().Add(-1 * time.Minute)

	uc := &AuthUC{
		user: &authUserRepoMock{
			createFn:           func(u entity.User) (entity.User, error) { return entity.User{}, nil },
			getByEmailFn:       func(email string) (entity.User, error) { return entity.User{}, nil },
			getByIDFn:          func(id int64) (entity.User, error) { return entity.User{}, nil },
			setEmailVerifiedFn: func(userID int64) error { return nil },
			updatePassHashFn:   func(userID int64, newHash string) error { return nil },
			bumpTokenVerFn: func(userID int64) (int, error) {
				bumpedUserID = userID
				return 2, nil
			},
		},
		ev: &authEvRepoMock{},
		pw: &authPwRepoMock{},
		rt: &authRtRepoMock{
			createFn: func(rt entity.RefreshToken) (entity.RefreshToken, error) { return entity.RefreshToken{}, nil },
			getByTokenHashFn: func(hash string) (entity.RefreshToken, error) {
				return entity.RefreshToken{
					ID:        100,
					UserID:    55,
					FamilyID:  "family-reuse",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					UsedAt:    &usedAt,
				}, nil
			},
			revokeFn:        func(id int64) error { return nil },
			markUsedFn:      func(id int64) error { return nil },
			setReplacedByFn: func(id int64, newID int64) error { return nil },
			revokeByFamilyFn: func(familyID string) error {
				revokedFamily = familyID
				return nil
			},
			revokeAllByUserFn: func(userID int64) error { return nil },
		},
		audit: &authAuditRepoMock{createFn: func(a entity.AuditLog) error { return nil }},
		val:   &authValMock{},
		ph:    &pwHashMock{},
		tk:    &tokMock{},
		mail:  &mailerMock{},
		rl: &rateLimMock{
			allowRefreshFn:    func(userID int64) (bool, int, error) { return true, 0, nil },
			allowSignupFn:     func(ip string) (bool, int, error) { return true, 0, nil },
			allowLoginFn:      func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowResendIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowResendMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowForgotIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowForgotMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	_, err := uc.Refresh(RefreshIn{RefreshToken: "raw-old", IP: "127.0.0.1", UA: "test"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
	if revokedFamily != "family-reuse" {
		t.Fatalf("revoked family = %q, want family-reuse", revokedFamily)
	}
	if bumpedUserID != 55 {
		t.Fatalf("bumped user id = %d, want 55", bumpedUserID)
	}
}

func TestAuthUCLogout_OK(t *testing.T) {
	t.Parallel()

	var bumpedUserID int64
	var revokedFamily string
	var gotAuditType string

	uc := &AuthUC{
		user: &authUserRepoMock{
			createFn:           func(u entity.User) (entity.User, error) { return entity.User{}, nil },
			getByEmailFn:       func(email string) (entity.User, error) { return entity.User{}, nil },
			getByIDFn:          func(id int64) (entity.User, error) { return entity.User{}, nil },
			setEmailVerifiedFn: func(userID int64) error { return nil },
			updatePassHashFn:   func(userID int64, newHash string) error { return nil },
			bumpTokenVerFn: func(userID int64) (int, error) {
				bumpedUserID = userID
				return 3, nil
			},
		},
		ev: &authEvRepoMock{},
		pw: &authPwRepoMock{},
		rt: &authRtRepoMock{
			createFn: func(rt entity.RefreshToken) (entity.RefreshToken, error) { return entity.RefreshToken{}, nil },
			getByTokenHashFn: func(hash string) (entity.RefreshToken, error) {
				if hash != sha256Hex("refresh-on-logout") {
					t.Fatalf("unexpected token hash: %q", hash)
				}
				return entity.RefreshToken{FamilyID: "logout-family"}, nil
			},
			revokeFn:        func(id int64) error { return nil },
			markUsedFn:      func(id int64) error { return nil },
			setReplacedByFn: func(id int64, newID int64) error { return nil },
			revokeByFamilyFn: func(familyID string) error {
				revokedFamily = familyID
				return nil
			},
			revokeAllByUserFn: func(userID int64) error { return nil },
		},
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

	err := uc.Logout(LogoutIn{UserID: 9, RefreshToken: "refresh-on-logout", IP: "127.0.0.1", UA: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bumpedUserID != 9 {
		t.Fatalf("bumped user id = %d, want 9", bumpedUserID)
	}
	if revokedFamily != "logout-family" {
		t.Fatalf("revoked family = %q, want logout-family", revokedFamily)
	}
	if gotAuditType != "auth.logout" {
		t.Fatalf("audit type = %q, want auth.logout", gotAuditType)
	}
}

func TestAuthUCRefresh_MarkUsedConflict_BecomesUnauthorized(t *testing.T) {
	t.Parallel()

	uc := &AuthUC{
		user: &authUserRepoMock{
			createFn:     func(u entity.User) (entity.User, error) { return entity.User{}, nil },
			getByEmailFn: func(email string) (entity.User, error) { return entity.User{}, nil },
			getByIDFn: func(id int64) (entity.User, error) {
				return entity.User{ID: 8, Role: "user", TokenVer: 2}, nil
			},
			setEmailVerifiedFn: func(userID int64) error { return nil },
			updatePassHashFn:   func(userID int64, newHash string) error { return nil },
			bumpTokenVerFn:     func(userID int64) (int, error) { return 3, nil },
		},
		ev: &authEvRepoMock{},
		pw: &authPwRepoMock{},
		rt: &authRtRepoMock{
			createFn: func(rt entity.RefreshToken) (entity.RefreshToken, error) {
				rt.ID = 30
				return rt, nil
			},
			getByTokenHashFn: func(hash string) (entity.RefreshToken, error) {
				return entity.RefreshToken{
					ID:        20,
					UserID:    8,
					FamilyID:  "fam-1",
					ExpiresAt: time.Now().Add(10 * time.Minute),
				}, nil
			},
			revokeFn:          func(id int64) error { return nil },
			markUsedFn:        func(id int64) error { return repository.ErrConflict },
			setReplacedByFn:   func(id int64, newID int64) error { return nil },
			revokeByFamilyFn:  func(familyID string) error { return nil },
			revokeAllByUserFn: func(userID int64) error { return nil },
		},
		audit: &authAuditRepoMock{createFn: func(a entity.AuditLog) error { return nil }},
		val:   &authValMock{},
		ph:    &pwHashMock{},
		tk: &tokMock{
			newAccessFn:   func(userID int64, role string, tokenVer int) (string, error) { return "access-2", nil },
			newCSRFFn:     func() (string, error) { return "csrf-2", nil },
			newOpaqueFn:   func() (string, error) { return "refresh-2", nil },
			newFamilyIDFn: func() (string, error) { return "", nil },
		},
		mail: &mailerMock{},
		rl: &rateLimMock{
			allowRefreshFn:    func(userID int64) (bool, int, error) { return true, 0, nil },
			allowSignupFn:     func(ip string) (bool, int, error) { return true, 0, nil },
			allowLoginFn:      func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowResendIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowResendMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
			allowForgotIPFn:   func(ip string) (bool, int, error) { return true, 0, nil },
			allowForgotMailFn: func(emailHash string) (bool, int, error) { return true, 0, nil },
		},
	}

	_, err := uc.Refresh(RefreshIn{RefreshToken: "refresh-1", IP: "127.0.0.1", UA: "test"})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}
