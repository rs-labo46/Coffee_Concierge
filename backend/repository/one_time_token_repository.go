package repository

import (
	"errors"
	"time"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

type evRepository struct {
	db *gorm.DB
}

// パスワード再設定tokenの発行・取得・使用済み
type pwRepository struct {
	db *gorm.DB
}

func NewEvRepository(db *gorm.DB) EmailVerifyRepository {
	return &evRepository{db}
}

func NewPwRepository(db *gorm.DB) PwResetRepository {
	return &pwRepository{db: db}
}

// email_verifiesに1件保存する。
func (r *evRepository) Create(v *entity.EmailVerify) error {
	if v == nil {
		return ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(v).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

// token_hashでEmailVerifyTokenを取得する。
func (r *evRepository) GetByTokenHash(tokenHash string) (*entity.EmailVerify, error) {
	var v entity.EmailVerify

	// token_hashで検索する。
	err := r.db.Where("token_hash = ?", tokenHash).First(&v).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &v, nil
}

// used_atが未設定のtokenを使用済みにする。
func (r *evRepository) MarkUsed(id uint, usedAt time.Time) error {
	// 0は不正ID。
	if id == 0 {
		return ErrNotFound
	}

	// used_atがNULLのものだけ更新する。
	res := r.db.
		Model(&entity.EmailVerify{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", usedAt)

	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected > 0 {
		return nil
	}

	// 更新できなかった場合は、存在しないのか、すでに使用済みなのかを切り分ける。
	var v entity.EmailVerify
	err := r.db.
		Select("id", "used_at").
		First(&v, id).
		Error
	if err != nil {
		// 対象そのものが存在しない。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}

	// 取得できてused_at が埋まっているなら二重使用。
	if v.UsedAt != nil {
		return ErrConflict
	}

	return ErrInternal
}

// expires_atを過ぎたEmailVerifyTokenを削除する。
func (r *evRepository) DeleteExpired(now time.Time) error {
	// 期限切れtokenを削除する。
	res := r.db.
		Where("expires_at < ?", now).
		Delete(&entity.EmailVerify{})

	if res.Error != nil {
		return ErrInternal
	}

	return nil
}

// pw_resetsに1件保存する。
func (r *pwRepository) Create(p *entity.PwReset) error {
	// nil は不正状態。
	if p == nil {
		return ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(p).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

// token_hashでPasswordResetTokenを取得する。
func (r *pwRepository) GetByTokenHash(tokenHash string) (*entity.PwReset, error) {
	var p entity.PwReset

	// token_hashで検索する。
	err := r.db.
		Where("token_hash = ?", tokenHash).
		First(&p).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &p, nil
}

// used_atが未設定のPasswordResetTokenを使用済みにする。
func (r *pwRepository) MarkUsed(id uint, usedAt time.Time) error {
	// 0は不正ID。
	if id == 0 {
		return ErrNotFound
	}

	// 未使用tokenのみ更新する。
	res := r.db.Model(&entity.PwReset{}).Where("id = ? AND used_at IS NULL", id).Update("used_at", usedAt)

	if res.Error != nil {
		return ErrInternal
	}

	// 更新成功。
	if res.RowsAffected > 0 {
		return nil
	}

	// 存在しないのか、二重使用かを切り分ける。
	var p entity.PwReset
	err := r.db.
		Select("id", "used_at").
		First(&p, id).
		Error
	if err != nil {
		// 対象そのものが存在しない。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}
	if p.UsedAt != nil {
		return ErrConflict
	}
	return ErrInternal
}

// expires_atを過ぎた PasswordResetTokenを削除する。
func (r *pwRepository) DeleteExpired(now time.Time) error {
	// 期限切れtokenを削除する。
	res := r.db.
		Where("expires_at < ?", now).
		Delete(&entity.PwReset{})
	if res.Error != nil {
		return ErrInternal
	}

	return nil
}
