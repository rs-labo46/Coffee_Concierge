package repository

import (
	"coffee-spa/apperr"
	"coffee-spa/entity"

	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *entity.User) error
	GetByID(id uint) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	Update(user *entity.User) error
	UpdateTokenVer(userID uint, tokenVer int) error
}

type EmailVerifyRepository interface {
	Create(v *entity.EmailVerify) error
	GetByTokenHash(tokenHash string) (*entity.EmailVerify, error)
	MarkUsed(id uint, usedAt time.Time) error
	DeleteExpired(now time.Time) error
}

type PwResetRepository interface {
	Create(r *entity.PwReset) error
	GetByTokenHash(tokenHash string) (*entity.PwReset, error)
	MarkUsed(id uint, usedAt time.Time) error
	DeleteExpired(now time.Time) error
}

type RtRepository interface {
	Create(rt *entity.Rt) error
	GetByTokenHash(tokenHash string) (*entity.Rt, error)
	Update(rt *entity.Rt) error
	RevokeFamily(familyID string, revokedAt time.Time) error
	DeleteExpired(now time.Time) error
}

type userRepository struct {
	db *gorm.DB
}
type rtRepository struct {
	db *gorm.DB
}
type evRepository struct {
	db *gorm.DB
}

// パスワード再設定tokenの発行・取得・使用済み
type pwRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func NewEmailVerifyRepository(db *gorm.DB) EmailVerifyRepository {
	return &evRepository{
		db: db,
	}
}

func NewPwResetRepository(db *gorm.DB) PwResetRepository {
	return &pwRepository{
		db: db,
	}
}

func NewRtRepository(db *gorm.DB) RtRepository {
	return &rtRepository{
		db: db,
	}
}

func isDup(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}

	return strings.Contains(err.Error(), "duplicate key")
}

func isFK(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return true
	}

	msg := err.Error()
	return strings.Contains(msg, "foreign key") ||
		strings.Contains(msg, "violates foreign key constraint")
}

// usersに新規ユーザーを保存する。
func (r *userRepository) Create(user *entity.User) error {
	if user == nil {
		return apperr.ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(user).Error
	if err != nil {
		if isDup(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// 主キーでusersを1件取得する。
func (r *userRepository) GetByID(id uint) (*entity.User, error) {
	// 0は不正ID。
	if id == 0 {
		return nil, apperr.ErrNotFound
	}
	var user entity.User

	// 主キーで検索を行う。
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &user, nil
}

// emailでusersを1件取得する。
func (r *userRepository) GetByEmail(email string) (*entity.User, error) {
	var user entity.User

	// emailの完全一致で1件取得する。
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &user, nil
}

// usersの可変項目を更新する(SaveではなくUpdatesを使い、主キー0の誤まりでのinsertを防ぐ)。
func (r *userRepository) Update(user *entity.User) error {
	if user == nil {
		return apperr.ErrInvalidState
	}

	// 主キーなしの更新を防ぐ。
	if user.ID == 0 {
		return apperr.ErrInvalidState
	}

	// 更新対象を明示し、created_atは更新しない。
	res := r.db.Model(&entity.User{}).Where("id = ?", user.ID).Select(
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
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

// token_verを指定した値へと更新する。
func (r *userRepository) UpdateTokenVer(userID uint, tokenVer int) error {
	// 0は不正ID。
	if userID == 0 {
		return apperr.ErrNotFound
	}

	// 指定されたuserのtoken_verを更新する。
	res := r.db.Model(&entity.User{}).Where("id = ?", userID).Update("token_ver", tokenVer)

	if res.Error != nil {
		return apperr.ErrInternal
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound
	}

	return nil
}

// email_verifiesに1件保存する。
func (r *evRepository) Create(v *entity.EmailVerify) error {
	if v == nil {
		return apperr.ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(v).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
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
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &v, nil
}

// used_atが未設定のtokenを使用済みにする。
func (r *evRepository) MarkUsed(id uint, usedAt time.Time) error {
	// 0は不正ID。
	if id == 0 {
		return apperr.ErrNotFound
	}

	// used_atがNULLのものだけ更新する。
	res := r.db.
		Model(&entity.EmailVerify{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", usedAt)

	if res.Error != nil {
		return apperr.ErrInternal
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
			return apperr.ErrNotFound
		}
		return apperr.ErrInternal
	}

	// 取得できてused_at が埋まっているなら二重使用。
	if v.UsedAt != nil {
		return apperr.ErrConflict
	}

	return apperr.ErrInternal
}

// expires_atを過ぎたEmailVerifyTokenを削除する。
func (r *evRepository) DeleteExpired(now time.Time) error {
	// 期限切れtokenを削除する。
	res := r.db.
		Where("expires_at < ?", now).
		Delete(&entity.EmailVerify{})

	if res.Error != nil {
		return apperr.ErrInternal
	}

	return nil
}

// pw_resetsに1件保存する。
func (r *pwRepository) Create(p *entity.PwReset) error {
	// nil は不正状態。
	if p == nil {
		return apperr.ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(p).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
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
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &p, nil
}

// used_atが未設定のPasswordResetTokenを使用済みにする。
func (r *pwRepository) MarkUsed(id uint, usedAt time.Time) error {
	// 0は不正ID。
	if id == 0 {
		return apperr.ErrNotFound
	}

	// 未使用tokenのみ更新する。
	res := r.db.Model(&entity.PwReset{}).Where("id = ? AND used_at IS NULL", id).Update("used_at", usedAt)

	if res.Error != nil {
		return apperr.ErrInternal
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
			return apperr.ErrNotFound
		}
		return apperr.ErrInternal
	}
	if p.UsedAt != nil {
		return apperr.ErrConflict
	}
	return apperr.ErrInternal
}

// expires_atを過ぎた PasswordResetTokenを削除する。
func (r *pwRepository) DeleteExpired(now time.Time) error {
	// 期限切れtokenを削除する。
	res := r.db.
		Where("expires_at < ?", now).
		Delete(&entity.PwReset{})
	if res.Error != nil {
		return apperr.ErrInternal
	}

	return nil
}

// refresh_tokensに1件保存する。
func (r *rtRepository) Create(rt *entity.Rt) error {
	if rt == nil {
		return apperr.ErrInvalidState
	}

	// INSERTを実行。
	err := r.db.Create(rt).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// token_hashでrefresh tokenを1件取得する。
func (r *rtRepository) GetByTokenHash(tokenHash string) (*entity.Rt, error) {
	var rt entity.Rt

	//token_hashで検索する。
	err := r.db.Where("token_hash = ?", tokenHash).First(&rt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &rt, nil
}

// revoked_at/used_at/replaced_by_id更新やrotation。
func (r *rtRepository) Update(rt *entity.Rt) error {
	if rt == nil {
		return apperr.ErrInvalidState
	}

	// 主キーがない状態での更新を防ぐ。
	if rt.ID == 0 {
		return apperr.ErrInvalidState
	}

	// 更新対象を明示して更新し、created_atは更新しない
	res := r.db.Model(&entity.Rt{}).Where("id = ?", rt.ID).Select(
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
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	// 対象の行がない。
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound
	}

	return nil
}

// 同一family_idだった場合、revoke tokenをまとめて失効させる。
func (r *rtRepository) RevokeFamily(familyID string, revokedAt time.Time) error {
	// familyIDが空でもDBは動くが、不正状態。
	if familyID == "" {
		return apperr.ErrInvalidState
	}

	// 同じfamilyの失効tokenをまとめて更新する。
	res := r.db.Model(&entity.Rt{}).Where("family_id = ? AND revoked_at IS NULL", familyID).Update("revoked_at", revokedAt)

	if res.Error != nil {
		return apperr.ErrInternal
	}

	return nil
}

// 削除件数0でも正常。
func (r *rtRepository) DeleteExpired(now time.Time) error {
	// 期限切れ token を削除する。
	res := r.db.Where("expires_at < ?", now).Delete(&entity.Rt{})

	if res.Error != nil {
		return apperr.ErrInternal
	}
	return nil
}
