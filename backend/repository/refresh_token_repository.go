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
	return &rtRepository{db}
}

func (r *rtRepository) Create(rt entity.RefreshToken) (entity.RefreshToken, error) {
	err := r.db.Create(&rt).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return entity.RefreshToken{}, ErrConflict
		}
		return entity.RefreshToken{}, ErrInternal
	}

	return rt, nil
}

func (r *rtRepository) GetByTokenHash(hash string) (entity.RefreshToken, error) {
	var rt entity.RefreshToken

	err := r.db.
		Where("token_hash = ?", hash).
		First(&rt).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.RefreshToken{}, ErrNotFound
		}
		return entity.RefreshToken{}, ErrInternal
	}

	return rt, nil
}

func (r *rtRepository) Revoke(id int64) error {
	now := time.Now()

	res := r.db.
		Model(&entity.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now)

	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *rtRepository) MarkUsed(id int64) error {
	now := time.Now()

	res := r.db.
		Model(&entity.RefreshToken{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now)

	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrConflict
	}

	return nil
}

func (r *rtRepository) SetReplacedBy(id int64, newID int64) error {
	res := r.db.
		Model(&entity.RefreshToken{}).
		Where("id = ?", id).
		Update("replaced_by_id", newID)

	if res.Error != nil {
		if isFK(res.Error) {
			return ErrConflict
		}
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *rtRepository) RevokeByFamilyID(familyID string) error {
	now := time.Now()

	res := r.db.
		Model(&entity.RefreshToken{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Update("revoked_at", now)

	if res.Error != nil {
		return ErrInternal
	}

	return nil
}

func (r *rtRepository) RevokeAllByUser(userID int64) error {
	now := time.Now()

	res := r.db.
		Model(&entity.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now)

	if res.Error != nil {
		return ErrInternal
	}

	return nil
}
