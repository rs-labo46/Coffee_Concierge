package repository

import (
	"errors"
	"time"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

type rtRepository struct {
	db *gorm.DB
}

func NewRtRepository(db *gorm.DB) RtRepository {
	return &rtRepository{db: db}
}

// refresh_tokensに1件保存する。
func (r *rtRepository) Create(rt *entity.Rt) error {
	if rt == nil {
		return ErrInvalidState
	}

	// INSERTを実行。
	err := r.db.Create(rt).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

// token_hashでrefresh tokenを1件取得する。
func (r *rtRepository) GetByTokenHash(tokenHash string) (*entity.Rt, error) {
	var rt entity.Rt

	//token_hashで検索する。
	err := r.db.
		Where("token_hash = ?", tokenHash).
		First(&rt).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &rt, nil
}

// revoked_at/used_at/replaced_by_id更新やrotation。
func (r *rtRepository) Update(rt *entity.Rt) error {
	if rt == nil {
		return ErrInvalidState
	}

	// 主キーがない状態での更新を防ぐ。
	if rt.ID == 0 {
		return ErrInvalidState
	}

	// 更新対象を明示して更新し、created_atは更新しない
	res := r.db.
		Model(&entity.Rt{}).
		Where("id = ?", rt.ID).
		Select(
			"user_id",
			"family_id",
			"token_hash",
			"expires_at",
			"revoked_at",
			"used_at",
			"replaced_by_id",
		).
		Updates(rt)

	if res.Error != nil {
		if isDup(res.Error) || isFK(res.Error) {
			return ErrConflict
		}
		return ErrInternal
	}

	// 対象の行がない。
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// 同一family_idだった場合、revoke tokenをまとめて失効させる。
func (r *rtRepository) RevokeFamily(familyID string, revokedAt time.Time) error {
	// familyIDが空でもDBは動くが、不正状態。
	if familyID == "" {
		return ErrInvalidState
	}

	// 同じfamilyの失効tokenをまとめて更新する。
	res := r.db.
		Model(&entity.Rt{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Update("revoked_at", revokedAt)

	if res.Error != nil {
		return ErrInternal
	}

	return nil
}

// 削除件数0でも正常。
func (r *rtRepository) DeleteExpired(now time.Time) error {
	// 期限切れ token を削除する。
	res := r.db.
		Where("expires_at < ?", now).
		Delete(&entity.Rt{})

	if res.Error != nil {
		return ErrInternal
	}
	return nil
}
