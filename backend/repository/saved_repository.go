package repository

import (
	"errors"

	"coffee-spa/apperr"
	"coffee-spa/entity"

	"gorm.io/gorm"
)

type SavedRepository interface {
	Create(saved *entity.SavedSuggestion) error
	List(q SavedListQ) ([]entity.SavedSuggestion, error)
	DeleteByUserAndSuggestionID(userID uint, suggestionID uint) error
	GetByUserAndSuggestionID(userID uint, suggestionID uint) (*entity.SavedSuggestion, error)
}

// SavedListQは、保存済み提案一覧の検索条件を表す。
type SavedListQ struct {
	UserID uint
	Limit  int
	Offset int
}

type savedRepository struct {
	db *gorm.DB
}

func NewSavedRepository(db *gorm.DB) SavedRepository {
	return &savedRepository{
		db: db,
	}
}

// saved suggestionを新規作成。
func (r *savedRepository) Create(saved *entity.SavedSuggestion) error {
	if saved == nil {
		return apperr.ErrInvalidState
	}

	err := r.db.Create(saved).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// ユーザーの保存済み提案一覧を返す。
func (r *savedRepository) List(q SavedListQ) ([]entity.SavedSuggestion, error) {
	if q.UserID == 0 {
		return nil, apperr.ErrUnauthorized
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	var saveds []entity.SavedSuggestion

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
func (r *savedRepository) DeleteByUserAndSuggestionID(userID uint, suggestionID uint) error {
	if userID == 0 || suggestionID == 0 {
		return apperr.ErrInvalidState
	}

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
