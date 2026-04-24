package repository

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/apperr"
	"coffee-spa/entity"
)

func TestUserRepository_CreateGetUpdateLifecycle(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)

	user := &entity.User{Email: "user@example.com", PassHash: "hash", Role: entity.RoleUser, TokenVer: 1}
	if err := repo.Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	got, err := repo.GetByEmail("user@example.com")
	if err != nil {
		t.Fatalf("get by email: %v", err)
	}
	if got.ID == 0 || got.Email != "user@example.com" {
		t.Fatalf("unexpected user: %#v", got)
	}
	got.EmailVerified = true
	got.TokenVer = 2
	if err := repo.Update(got); err != nil {
		t.Fatalf("update user: %v", err)
	}
	if err := repo.UpdateTokenVer(got.ID, 3); err != nil {
		t.Fatalf("update token version: %v", err)
	}
	updated, err := repo.GetByID(got.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if updated.TokenVer != 3 || !updated.EmailVerified {
		t.Fatalf("user was not updated: %#v", updated)
	}
}

func TestUserRepository_DuplicateEmail_ReturnsConflict(t *testing.T) {
	db := openTestDB(t)
	repo := NewUserRepository(db)
	first := &entity.User{Email: "dup@example.com", PassHash: "hash", Role: entity.RoleUser, TokenVer: 1}
	second := &entity.User{Email: "dup@example.com", PassHash: "hash", Role: entity.RoleUser, TokenVer: 1}
	if err := repo.Create(first); err != nil {
		t.Fatalf("create first user: %v", err)
	}
	if err := repo.Create(second); !errors.Is(err, apperr.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestEmailVerifyRepository_MarkUsedIsOneTime(t *testing.T) {
	db := openTestDB(t)
	user := seedUser(t, db, "verify@example.com", entity.RoleUser)
	repo := NewEmailVerifyRepository(db)
	token := &entity.EmailVerify{UserID: user.ID, TokenHash: "verify-hash", ExpiresAt: time.Now().Add(time.Hour)}
	if err := repo.Create(token); err != nil {
		t.Fatalf("create verify token: %v", err)
	}
	usedAt := time.Now()
	if err := repo.MarkUsed(token.ID, usedAt); err != nil {
		t.Fatalf("mark used: %v", err)
	}
	if err := repo.MarkUsed(token.ID, usedAt); !errors.Is(err, apperr.ErrInvalidState) {
		t.Fatalf("second mark should be invalid state, got %v", err)
	}
}

func TestPwResetRepository_MarkUsedIsOneTime(t *testing.T) {
	db := openTestDB(t)
	user := seedUser(t, db, "pw@example.com", entity.RoleUser)
	repo := NewPwResetRepository(db)
	token := &entity.PwReset{UserID: user.ID, TokenHash: "pw-hash", ExpiresAt: time.Now().Add(time.Hour)}
	if err := repo.Create(token); err != nil {
		t.Fatalf("create pw token: %v", err)
	}
	usedAt := time.Now()
	if err := repo.MarkUsed(token.ID, usedAt); err != nil {
		t.Fatalf("mark used: %v", err)
	}
	if err := repo.MarkUsed(token.ID, usedAt); !errors.Is(err, apperr.ErrInvalidState) {
		t.Fatalf("second mark should be invalid state, got %v", err)
	}
}

func TestRtRepository_RevokeFamily(t *testing.T) {
	db := openTestDB(t)
	user := seedUser(t, db, "rt@example.com", entity.RoleUser)
	repo := NewRtRepository(db)
	rt := &entity.Rt{UserID: user.ID, FamilyID: "family", TokenHash: "rt-hash", ExpiresAt: time.Now().Add(time.Hour)}
	if err := repo.Create(rt); err != nil {
		t.Fatalf("create refresh token: %v", err)
	}
	revokedAt := time.Now()
	if err := repo.RevokeFamily("family", revokedAt); err != nil {
		t.Fatalf("revoke family: %v", err)
	}
	got, err := repo.GetByTokenHash("rt-hash")
	if err != nil {
		t.Fatalf("get refresh token: %v", err)
	}
	if got.RevokedAt == nil {
		t.Fatalf("refresh token was not revoked")
	}
}
