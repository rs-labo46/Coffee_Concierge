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
	return &userRepository{db: db}
}

// usersに新規ユーザーを保存する。
func (r *userRepository) Create(user *entity.User) error {
	if user == nil {
		return ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(user).Error
	if err != nil {
		if isDup(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

// 主キーでusersを1件取得する。
func (r *userRepository) GetByID(id uint) (*entity.User, error) {
	// 0は不正ID。
	if id == 0 {
		return nil, ErrNotFound
	}
	var user entity.User

	// 主キーで検索を行う。
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &user, nil
}

// emailでusersを1件取得する。
func (r *userRepository) GetByEmail(email string) (*entity.User, error) {
	var user entity.User

	// emailの完全一致で1件取得する。
	err := r.db.
		Where("email = ?", email).
		First(&user).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &user, nil
}

// usersの可変項目を更新する(SaveではなくUpdatesを使い、主キー0の誤まりでのinsertを防ぐ)。
func (r *userRepository) Update(user *entity.User) error {
	if user == nil {
		return ErrInvalidState
	}

	// 主キーなしの更新を防ぐ。
	if user.ID == 0 {
		return ErrInvalidState
	}

	// 更新対象を明示し、created_atは更新しない。
	res := r.db.
		Model(&entity.User{}).
		Where("id = ?", user.ID).
		Select(
			"email",
			"pass_hash",
			"role",
			"token_ver",
			"email_verified",
			"updated_at",
		).
		Updates(user)

	if res.Error != nil {
		//emailの重複はconflict。
		if isDup(res.Error) {
			return ErrConflict
		}
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// token_verを指定した値へと更新する。
func (r *userRepository) UpdateTokenVer(userID uint, tokenVer int) error {
	// 0は不正ID。
	if userID == 0 {
		return ErrNotFound
	}

	// 指定されたuserのtoken_verを更新する。
	res := r.db.
		Model(&entity.User{}).
		Where("id = ?", userID).
		Update("token_ver", tokenVer)

	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
