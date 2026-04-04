package usecase

import (
	"errors"
	"strings"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

// Loginはaccess / refresh / csrfを発行する
func (u *AuthUC) Login(input LoginIn) (AuthOut, error) {
	if err := u.val.Login(input.Email, input.Pw); err != nil {
		return AuthOut{}, ErrInvalidRequest
	}

	email := normEmail(input.Email)
	emailHash := sha256Hex(email)

	ok, retry, err := u.rl.AllowLogin(emailHash)
	if err != nil {
		return AuthOut{}, ErrInternal
	}
	if !ok {
		return AuthOut{}, ErrRateLimited{RetryAfterSec: retry}
	}

	user, err := u.user.GetByEmail(email)
	if err != nil {
		return AuthOut{}, u.writeLoginUnauthorized(nil, input)
	}

	if !user.EmailVerified {
		return AuthOut{}, u.writeLoginUnauthorized(int64Pointer(user.ID), input)
	}

	if err := u.ph.Compare(user.PassHash, input.Pw); err != nil {
		return AuthOut{}, u.writeLoginUnauthorized(int64Pointer(user.ID), input)
	}

	familyID, err := u.tk.NewFamilyID()
	if err != nil {
		return AuthOut{}, ErrInternal
	}

	rawRefresh, err := u.tk.NewOpaque()
	if err != nil {
		return AuthOut{}, ErrInternal
	}

	_, err = u.rt.Create(entity.RefreshToken{
		UserID:    user.ID,
		FamilyID:  familyID,
		TokenHash: sha256Hex(rawRefresh),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		return AuthOut{}, mapRepoErr(err)
	}

	access, err := u.tk.NewAccess(user.ID, user.Role, user.TokenVer)
	if err != nil {
		return AuthOut{}, ErrInternal
	}

	csrf, err := u.tk.NewCSRF()
	if err != nil {
		return AuthOut{}, ErrInternal
	}

	if err := u.writeAudit(
		"auth.login.success",
		int64Pointer(user.ID),
		input.IP,
		input.UA,
		refreshOKMeta{
			UserID:   user.ID,
			FamilyID: familyID,
		},
	); err != nil {
		return AuthOut{}, err
	}

	return AuthOut{
		AccessToken:  access,
		RefreshToken: rawRefresh,
		CsrfToken:    csrf,
		User:         user,
	}, nil
}

// refresh tokenをローリングさせる
func (u *AuthUC) Refresh(input RefreshIn) (AuthOut, error) {
	if strings.TrimSpace(input.RefreshToken) == "" {
		_ = u.writeAudit("auth.refresh.fail", nil, input.IP, input.UA, nil)
		return AuthOut{}, ErrUnauthorized
	}

	rt, err := u.rt.GetByTokenHash(sha256Hex(input.RefreshToken))
	if err != nil {
		_ = u.writeAudit("auth.refresh.fail", nil, input.IP, input.UA, nil)
		return AuthOut{}, ErrUnauthorized
	}

	ok, retry, err := u.rl.AllowRefresh(rt.UserID)
	if err != nil {
		return AuthOut{}, ErrInternal
	}
	if !ok {
		return AuthOut{}, ErrRateLimited{RetryAfterSec: retry}
	}

	if time.Now().After(rt.ExpiresAt) || rt.RevokedAt != nil {
		_ = u.writeAudit("auth.refresh.fail", int64Pointer(rt.UserID), input.IP, input.UA, nil)
		return AuthOut{}, ErrUnauthorized
	}

	if rt.UsedAt != nil {
		return AuthOut{}, u.handleRefreshReuse(rt, input)
	}

	user, err := u.user.GetByID(rt.UserID)
	if err != nil {
		return AuthOut{}, mapRepoErr(err)
	}

	rawRefresh, err := u.tk.NewOpaque()
	if err != nil {
		return AuthOut{}, ErrInternal
	}

	newRT, err := u.rt.Create(entity.RefreshToken{
		UserID:    rt.UserID,
		FamilyID:  rt.FamilyID,
		TokenHash: sha256Hex(rawRefresh),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		return AuthOut{}, mapRepoErr(err)
	}

	if err := u.rt.MarkUsed(rt.ID); err != nil {
		_ = u.rt.Revoke(newRT.ID)
		if errors.Is(err, repository.ErrConflict) {
			return AuthOut{}, u.handleRefreshReuse(rt, input)
		}
		return AuthOut{}, mapRepoErr(err)
	}

	if err := u.rt.Revoke(rt.ID); err != nil {
		_ = u.rt.Revoke(newRT.ID)
		return AuthOut{}, mapRepoErr(err)
	}

	if err := u.rt.SetReplacedBy(rt.ID, newRT.ID); err != nil {
		_ = u.rt.Revoke(newRT.ID)
		return AuthOut{}, mapRepoErr(err)
	}

	access, err := u.tk.NewAccess(user.ID, user.Role, user.TokenVer)
	if err != nil {
		_ = u.rt.Revoke(newRT.ID)
		return AuthOut{}, ErrInternal
	}

	csrf, err := u.tk.NewCSRF()
	if err != nil {
		_ = u.rt.Revoke(newRT.ID)
		return AuthOut{}, ErrInternal
	}

	if err := u.writeAudit(
		"auth.refresh.success",
		int64Pointer(user.ID),
		input.IP,
		input.UA,
		refreshOKMeta{
			UserID:   user.ID,
			FamilyID: rt.FamilyID,
		},
	); err != nil {
		_ = u.rt.Revoke(newRT.ID)
		return AuthOut{}, err
	}

	return AuthOut{
		AccessToken:  access,
		RefreshToken: rawRefresh,
		CsrfToken:    csrf,
		User:         user,
	}, nil
}

// Logoutはtoken_verを上げて、refresh familyを失効する
func (u *AuthUC) Logout(in LogoutIn) error {
	_, err := u.user.BumpTokenVer(in.UserID)
	if err != nil {
		return mapRepoErr(err)
	}

	if in.RefreshToken != "" {
		rt, err := u.rt.GetByTokenHash(sha256Hex(in.RefreshToken))
		if err == nil {
			if err := u.rt.RevokeByFamilyID(rt.FamilyID); err != nil {
				return mapRepoErr(err)
			}
		}
	}

	if err := u.writeAudit(
		"auth.logout",
		int64Pointer(in.UserID),
		in.IP,
		in.UA,
		logoutMeta{UserID: in.UserID},
	); err != nil {
		return err
	}

	return nil
}

func (u *AuthUC) writeLoginUnauthorized(userID *int64, input LoginIn) error {
	_ = u.writeAudit(
		"auth.login.fail",
		userID,
		input.IP,
		input.UA,
		loginFailMeta{Reason: "unauthorized"},
	)
	return ErrUnauthorized
}

func (u *AuthUC) handleRefreshReuse(rt entity.RefreshToken, input RefreshIn) error {
	_ = u.rt.RevokeByFamilyID(rt.FamilyID)
	_, _ = u.user.BumpTokenVer(rt.UserID)

	_ = u.writeAudit(
		"auth.refresh.reuse_detected",
		int64Pointer(rt.UserID),
		input.IP,
		input.UA,
		refreshReuseMeta{
			UserID:   rt.UserID,
			FamilyID: rt.FamilyID,
			RtID:     rt.ID,
		},
	)

	return ErrUnauthorized
}
