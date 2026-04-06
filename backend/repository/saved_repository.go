package repository

import (
	"errors"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

type savedRepository struct {
	db *gorm.DB
}

func NewSavedRepository(db *gorm.DB) SavedRepository {
	return &savedRepository{
		db: db,
	}
}

func (r *savedRepository) Create(saved *entity.SavedSuggestion) error {
	if saved == nil {
		return ErrInvalidState
	}

	err := r.db.Create(saved).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *savedRepository) List(q SavedListQ) ([]entity.SavedSuggestion, error) {
	if q.UserID == 0 {
		return nil, ErrUnauthorized
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
		return nil, ErrInternal
	}

	return saveds, nil
}

func (r *savedRepository) DeleteByUserAndSuggestionID(userID uint, suggestionID uint) error {
	if userID == 0 || suggestionID == 0 {
		return ErrInvalidState
	}

	res := r.db.
		Where("user_id = ?", userID).
		Where("suggestion_id = ?", suggestionID).
		Delete(&entity.SavedSuggestion{})
	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *savedRepository) GetByUserAndSuggestionID(
	userID uint,
	suggestionID uint,
) (*entity.SavedSuggestion, error) {
	if userID == 0 || suggestionID == 0 {
		return nil, ErrNotFound
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
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &saved, nil
}
