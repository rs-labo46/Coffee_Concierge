package usecase_test

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/apperr"
	"coffee-spa/entity"
	"coffee-spa/testutil/repomock"
	"coffee-spa/testutil/servicemock"
	"coffee-spa/testutil/valmock"
	"coffee-spa/usecase"
)

func newAuthUCTest(users *repomock.User, hasher servicemock.Hasher, val usecase.AuthVal) usecase.AuthUC {
	if users == nil {
		users = &repomock.User{}
	}
	if val == nil {
		val = valmock.Auth{}
	}
	return usecase.NewAuthUsecase(
		users,
		&repomock.EmailVerify{},
		&repomock.PwReset{},
		&repomock.Rt{},
		&repomock.Audit{},
		val,
		hasher,
		servicemock.Token{},
		servicemock.Refresh{},
		servicemock.Clock{},
		servicemock.IDGen{},
		servicemock.Mailer{},
		time.Hour,
		time.Hour,
		time.Hour,
	)
}

func TestAuthUsecase_AdditionalErrorCoverage(t *testing.T) {
	// U-AUTH-ADD-01
	t.Run("Signup returns validator error", func(t *testing.T) {
		want := errors.New("validation failed")
		uc := newAuthUCTest(nil, servicemock.Hasher{}, valmock.Auth{Err: want})
		_, err := uc.Signup(usecase.SignupIn{Email: "bad", Password: "short"})
		if !errors.Is(err, want) {
			t.Fatalf("Signup() error = %v, want %v", err, want)
		}
	})

	// U-AUTH-ADD-02
	t.Run("Login password mismatch returns unauthorized", func(t *testing.T) {
		uc := newAuthUCTest(&repomock.User{GetByEmailFn: func(email string) (*entity.User, error) {
			return &entity.User{ID: 1, Email: email, PassHash: "hash", Role: entity.RoleUser, TokenVer: 1, EmailVerified: true}, nil
		}}, servicemock.Hasher{CompareFn: func(hash string, raw string) error { return errors.New("mismatch") }}, nil)
		_, err := uc.Login(usecase.LoginIn{Email: "user@example.com", Password: "wrong"})
		if !errors.Is(err, usecase.ErrUnauthorized) {
			t.Fatalf("Login() error = %v, want unauthorized", err)
		}
	})

	// U-AUTH-ADD-03
	t.Run("ForgotPw hides missing user", func(t *testing.T) {
		uc := newAuthUCTest(&repomock.User{GetByEmailFn: func(email string) (*entity.User, error) {
			return nil, apperr.ErrNotFound
		}}, servicemock.Hasher{}, nil)
		if err := uc.ForgotPw(usecase.ForgotPwIn{Email: "none@example.com"}); err != nil {
			t.Fatalf("ForgotPw() error = %v, want nil", err)
		}
	})
}
