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

type pwRepository struct {
	db *gorm.DB
}

func NewEvRepository(db *gorm.DB) EvRepository {
	return &evRepository{db}
}

func (r *evRepository) Create(ev entity.EmailVerify) error {
	err := r.db.Create(&ev).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *evRepository) GetByTokenHash(hash string) (entity.EmailVerify, error) {
	var ev entity.EmailVerify

	err := r.db.
		Where("token_hash = ?", hash).
		First(&ev).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.EmailVerify{}, ErrNotFound
		}
		return entity.EmailVerify{}, ErrInternal
	}

	return ev, nil
}

func (r *evRepository) Use(id int64) error {
	now := time.Now()

	res := r.db.
		Model(&entity.EmailVerify{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now)

	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected > 0 {
		return nil
	}

	var ev entity.EmailVerify
	err := r.db.
		Select("id", "used_at").
		First(&ev, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}

	if ev.UsedAt != nil {
		return ErrConflict
	}

	return ErrInternal
}

func (r *evRepository) RevokeUnusedByUser(userID int64) error {
	now := time.Now()

	res := r.db.
		Model(&entity.EmailVerify{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now)

	if res.Error != nil {
		return ErrInternal
	}

	return nil
}

func NewPwRepository(db *gorm.DB) PwRepository {
	return &pwRepository{db}
}

func (r *pwRepository) Create(pw entity.PwReset) error {
	err := r.db.Create(&pw).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *pwRepository) GetByTokenHash(hash string) (entity.PwReset, error) {
	var pw entity.PwReset

	err := r.db.
		Where("token_hash = ?", hash).
		First(&pw).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.PwReset{}, ErrNotFound
		}
		return entity.PwReset{}, ErrInternal
	}

	return pw, nil
}

func (r *pwRepository) Use(id int64) error {
	now := time.Now()

	res := r.db.
		Model(&entity.PwReset{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now)

	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected > 0 {
		return nil
	}

	var pw entity.PwReset
	err := r.db.
		Select("id", "used_at").
		First(&pw, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return ErrInternal
	}

	if pw.UsedAt != nil {
		return ErrConflict
	}

	return ErrInternal
}

func (r *pwRepository) RevokeUnusedByUser(userID int64) error {
	now := time.Now()

	res := r.db.
		Model(&entity.PwReset{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now)

	if res.Error != nil {
		return ErrInternal
	}

	return nil
}
