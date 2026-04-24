package usecase

import (
	"testing"
	"time"

	"coffee-spa/entity"
)

func TestAuthUsecase_Signup(t *testing.T) {
	now := time.Date(2026, 4, 23, 9, 0, 0, 0, time.UTC)
	mailer := &mailerMock{}
	users := userRepoMock{createFn: func(user *entity.User) error { user.ID = 1; return nil }}
	verifies := emailVerifyRepoMock{createFn: func(v *entity.EmailVerify) error {
		if v.UserID != 1 { t.Fatalf("unexpected verify:%+v", v) }
		return nil
	}}
	uc := NewAuthUsecase(users, verifies, pwResetRepoMock{}, rtRepoMock{}, &auditRepoMock{}, authValMock{}, hasherMock{hash: "hashed"}, tokenSvcMock{}, refreshSvcMock{}, fixedClock{now: now}, fixedIDGen{id: "plain-token"}, mailer, time.Hour, time.Hour, time.Hour)
	out, err := uc.Signup(SignupIn{Email: "u@example.com", Password: "pw"})
	if err != nil { t.Fatalf("unexpected error:%v", err) }
	if out.User.ID != 1 || out.User.PassHash != "hashed" { t.Fatalf("unexpected user:%+v", out.User) }
	if mailer.verifyTo != "u@example.com" || mailer.verifyToken != "plain-token" { t.Fatalf("mailer not called:%+v", mailer) }
}

func TestAuthUsecase_Login_PasswordMismatch(t *testing.T) {
	uc := NewAuthUsecase(
		userRepoMock{getByEmailFn: func(email string) (*entity.User, error) { return &entity.User{ID: 1, Email: email, PassHash: "hashed", Role: entity.RoleUser, TokenVer: 1}, nil }},
		emailVerifyRepoMock{},
		pwResetRepoMock{},
		rtRepoMock{},
		&auditRepoMock{},
		authValMock{},
		hasherMock{compareErr: errDummy("mismatch")},
		tokenSvcMock{},
		refreshSvcMock{},
		fixedClock{now: time.Now()},
		fixedIDGen{id: "family"},
		nil,
		time.Hour, time.Hour, time.Hour,
	)
	_, err := uc.Login(LoginIn{Email: "u@example.com", Password: "bad", IP: "1.1.1.1", UA: "ua"})
	if err != ErrUnauthorized { t.Fatalf("expected ErrUnauthorized, got:%v", err) }
}

func TestAuthUsecase_Me(t *testing.T) {
	uc := NewAuthUsecase(userRepoMock{getByIDFn: func(id uint) (*entity.User, error) { return &entity.User{ID: id, Email: "u@example.com"}, nil }}, emailVerifyRepoMock{}, pwResetRepoMock{}, rtRepoMock{}, &auditRepoMock{}, authValMock{}, hasherMock{}, tokenSvcMock{}, refreshSvcMock{}, fixedClock{}, fixedIDGen{}, nil, time.Hour, time.Hour, time.Hour)
	out, err := uc.Me(entity.Actor{UserID: 9})
	if err != nil || out.ID != 9 { t.Fatalf("unexpected output:%+v err=%v", out, err) }
}
