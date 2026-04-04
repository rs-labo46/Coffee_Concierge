package usecase

import (
	"coffee-spa/entity"
	"coffee-spa/repository"
	"errors"
	"strings"
	"time"
)

// reset tokenを発行してメール送信する
func (u *AuthUC) ForgotPw(ForgotPasswordInput ForgotPwIn) error {
	email := normEmail(ForgotPasswordInput.Email)
	if err := u.val.EmailOnly(email); err != nil {
		return ErrInvalidRequest
	}

	emailHash := sha256Hex(email)

	okIP, _, err := u.rl.AllowForgotIP(ForgotPasswordInput.IP)
	if err != nil {
		return ErrInternal
	}

	okMail, _, err := u.rl.AllowForgotMail(emailHash)
	if err != nil {
		return ErrInternal
	}

	if !okIP || !okMail {
		_ = u.writeAudit(
			"auth.password.forgot.rate_limited",
			nil,
			ForgotPasswordInput.IP,
			ForgotPasswordInput.UA,
			forgotMeta{Email: maskEmail(email)},
		)
		return nil
	}

	user, err := u.user.GetByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			_ = u.writeAudit(
				"auth.password.forgot",
				nil,
				ForgotPasswordInput.IP,
				ForgotPasswordInput.UA,
				forgotMeta{Email: maskEmail(email)},
			)
			return nil
		}
		return ErrInternal
	}

	raw, err := u.tk.NewOpaque()
	if err != nil {
		return ErrInternal
	}

	if err := u.pw.RevokeUnusedByUser(user.ID); err != nil {
		return mapRepoErr(err)
	}

	err = u.pw.Create(entity.PwReset{
		UserID:    user.ID,
		TokenHash: sha256Hex(raw),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	})
	if err != nil {
		return mapRepoErr(err)
	}

	if err := u.mail.SendReset(user.Email, raw); err != nil {
		_ = u.writeAudit(
			"auth.password.forgot.mail_failed",
			int64Pointer(user.ID),
			ForgotPasswordInput.IP,
			ForgotPasswordInput.UA,
			verifyMeta{UserID: user.ID},
		)
	}

	if err := u.writeAudit(
		"auth.password.forgot",
		int64Pointer(user.ID),
		ForgotPasswordInput.IP,
		ForgotPasswordInput.UA,
		forgotMeta{Email: maskEmail(user.Email)},
	); err != nil {
		return err
	}

	return nil
}

// password更新 + reset token消費 + 全refresh失効
func (u *AuthUC) ResetPw(ResetPasswordInput ResetPwIn) error {
	if strings.TrimSpace(ResetPasswordInput.Token) == "" {
		return ErrInvalidRequest
	}

	if err := u.val.NewPw(ResetPasswordInput.NewPw); err != nil {
		return ErrInvalidRequest
	}

	pw, err := u.pw.GetByTokenHash(sha256Hex(ResetPasswordInput.Token))
	if err != nil {
		return ErrUnauthorized
	}

	if pw.UsedAt != nil || time.Now().After(pw.ExpiresAt) {
		return ErrUnauthorized
	}

	newHash, err := u.ph.Hash(ResetPasswordInput.NewPw)
	if err != nil {
		return ErrInternal
	}

	if err := u.user.UpdatePassHash(pw.UserID, newHash); err != nil {
		return mapRepoErr(err)
	}

	if err := u.pw.Use(pw.ID); err != nil {
		return mapRepoErr(err)
	}

	if _, err := u.user.BumpTokenVer(pw.UserID); err != nil {
		return mapRepoErr(err)
	}

	if err := u.rt.RevokeAllByUser(pw.UserID); err != nil {
		return mapRepoErr(err)
	}

	if err := u.writeAudit(
		"auth.password.reset",
		int64Pointer(pw.UserID),
		ResetPasswordInput.IP,
		ResetPasswordInput.UA,
		resetMeta{UserID: pw.UserID},
	); err != nil {
		return err
	}

	return nil
}

// userを返す
func (u *AuthUC) Me(userID int64) (entity.User, error) {
	user, err := u.user.GetByID(userID)
	if err != nil {
		return entity.User{}, mapRepoErr(err)
	}

	return user, nil
}
