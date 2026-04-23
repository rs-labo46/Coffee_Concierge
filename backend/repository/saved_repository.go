package repository

import (
	"errors"

	"coffee-spa/apperr"
	"coffee-spa/entity"
	"coffee-spa/usecase/port"

	"gorm.io/gorm"
)

// 保存済み提案の保存・一覧・取得・削除。
type savedRepository struct {
	db *gorm.DB
}

func NewSavedRepository(db *gorm.DB) port.SavedRepository {
	return &savedRepository{
		db: db,
	}
}

// saved suggestionを新規作成。
func (r *savedRepository) Create(saved *entity.SavedSuggestion) error {
	if saved == nil {
		return apperr.ErrInvalidState
	}

	// レコードをINSERT 。
	err := r.db.Create(saved).Error
	if err != nil {
		// 二重保存やFKの不整合はconflict。
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// ユーザーの保存済み提案一覧を返す。
func (r *savedRepository) List(q port.SavedListQ) ([]entity.SavedSuggestion, error) {
	if q.UserID == 0 {
		return nil, apperr.ErrUnauthorized
	}

	// limitは未指定なら20、上限は100に。
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// offsetはマイナスを許可しない。
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	var saveds []entity.SavedSuggestion

	// 一覧表示に必要なsuggestionと関連データ先読み取り 。
	err := r.db.
		Model(&entity.SavedSuggestion{}).
		Preload("Suggestion").
		Preload("Suggestion.Bean").
		Preload("Suggestion.Recipe").
		Preload("Suggestion.Item").
		Preload("Suggestion.Item.Source").
		Where("user_id = ?", q.UserID).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&saveds).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	return saveds, nil
}

// user_idとsuggestion_idの組で、保存済み提案を削除。
// 本人の保存データだけを削除するためのメソッドです。
func (r *savedRepository) DeleteByUserAndSuggestionID(userID uint, suggestionID uint) error {
	if userID == 0 || suggestionID == 0 {
		return apperr.ErrInvalidState
	}

	// 本人の保存データだけを対象に削除。
	res := r.db.
		Where("user_id = ?", userID).
		Where("suggestion_id = ?", suggestionID).
		Delete(&entity.SavedSuggestion{})
	if res.Error != nil {
		return apperr.ErrInternal
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound
	}

	return nil
}

// user_idとsuggestion_idの組で保存済み提案を1件取得。
func (r *savedRepository) GetByUserAndSuggestionID(
	userID uint,
	suggestionID uint,
) (*entity.SavedSuggestion, error) {
	if userID == 0 || suggestionID == 0 {
		return nil, apperr.ErrNotFound
	}

	var saved entity.SavedSuggestion

	// suggestionと関連データ先読み取りpreloadする。
	err := r.db.
		Preload("Suggestion").
		Preload("Suggestion.Bean").
		Preload("Suggestion.Recipe").
		Preload("Suggestion.Item").
		Preload("Suggestion.Item.Source").
		Where("user_id = ?", userID).
		Where("suggestion_id = ?", suggestionID).
		First(&saved).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &saved, nil
}
