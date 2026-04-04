package repository

import (
	"errors"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Create(u entity.User) (entity.User, error) {
	err := r.db.Create(&u).Error
	if err != nil {
		if isDup(err) {
			return entity.User{}, ErrConflict
		}
		return entity.User{}, ErrInternal
	}

	return u, nil
}

func (r *userRepository) GetByEmail(email string) (entity.User, error) {
	var u entity.User

	err := r.db.
		Where("email = ?", email).
		First(&u).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.User{}, ErrNotFound
		}
		return entity.User{}, ErrInternal
	}

	return u, nil
}

func (r *userRepository) GetByID(id int64) (entity.User, error) {
	var u entity.User

	err := r.db.
		First(&u, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity.User{}, ErrNotFound
		}
		return entity.User{}, ErrInternal
	}

	return u, nil
}

func (r *userRepository) SetEmailVerified(userID int64) error {
	res := r.db.
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("email_verified", true)

	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *userRepository) UpdatePassHash(userID int64, newHash string) error {
	res := r.db.
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("pass_hash", newHash)

	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *userRepository) BumpTokenVer(userID int64) (int, error) {
	res := r.db.
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("token_ver", gorm.Expr("token_ver + 1"))

	if res.Error != nil {
		return 0, ErrInternal
	}
	if res.RowsAffected == 0 {
		return 0, ErrNotFound
	}

	var u entity.User
	err := r.db.
		Select("token_ver").
		First(&u, userID).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrNotFound
		}
		return 0, ErrInternal
	}

	return u.TokenVer, nil
}
