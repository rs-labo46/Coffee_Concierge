package repository

import (
	"errors"
	"time"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{
		db: db,
	}
}

func (r *sessionRepository) CreateSession(session *entity.Session) error {
	if session == nil {
		return ErrInvalidState
	}

	err := r.db.Create(session).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *sessionRepository) GetSessionByID(id uint) (*entity.Session, error) {
	if id == 0 {
		return nil, ErrNotFound
	}

	var session entity.Session

	err := r.db.First(&session, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &session, nil
}

func (r *sessionRepository) GetGuestSessionByID(
	id uint,
	sessionKeyHash string,
	now time.Time,
) (*entity.Session, error) {
	if id == 0 || sessionKeyHash == "" {
		return nil, ErrNotFound
	}

	var session entity.Session

	err := r.db.
		Where("id = ?", id).
		Where("user_id IS NULL").
		Where("session_key_hash = ?", sessionKeyHash).
		Where("guest_expires_at IS NOT NULL").
		Where("guest_expires_at > ?", now).
		First(&session).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &session, nil
}

func (r *sessionRepository) ListHistory(q HistoryQ) ([]entity.Session, error) {
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

	var sessions []entity.Session

	err := r.db.
		Model(&entity.Session{}).
		Where("user_id = ?", q.UserID).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).
		Error
	if err != nil {
		return nil, ErrInternal
	}

	return sessions, nil
}

func (r *sessionRepository) CloseSession(id uint) error {
	if id == 0 {
		return ErrNotFound
	}

	now := time.Now()
	session := entity.Session{
		Status:    entity.SessionClosed,
		UpdatedAt: now,
	}

	res := r.db.
		Model(&entity.Session{}).
		Where("id = ?", id).
		Select("status", "updated_at").
		Updates(&session)
	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *sessionRepository) CreateTurn(turn *entity.Turn) error {
	if turn == nil {
		return ErrInvalidState
	}

	err := r.db.Create(turn).Error
	if err != nil {
		if isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *sessionRepository) ListTurns(sessionID uint) ([]entity.Turn, error) {
	if sessionID == 0 {
		return nil, ErrNotFound
	}

	var turns []entity.Turn

	err := r.db.
		Model(&entity.Turn{}).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Order("id ASC").
		Find(&turns).
		Error
	if err != nil {
		return nil, ErrInternal
	}

	return turns, nil
}

func (r *sessionRepository) CreatePref(pref *entity.Pref) error {
	if pref == nil {
		return ErrInvalidState
	}

	err := r.db.Create(pref).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *sessionRepository) UpdatePref(pref *entity.Pref) error {
	if pref == nil || pref.ID == 0 {
		return ErrInvalidState
	}

	err := r.db.
		Model(&entity.Pref{}).
		Where("id = ?", pref.ID).
		Select(
			"session_id",
			"flavor",
			"acidity",
			"bitterness",
			"body",
			"aroma",
			"mood",
			"method",
			"scene",
			"temp_pref",
			"excludes",
			"note",
			"updated_at",
		).
		Updates(pref).
		Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *sessionRepository) GetPrefBySessionID(sessionID uint) (*entity.Pref, error) {
	if sessionID == 0 {
		return nil, ErrNotFound
	}

	var pref entity.Pref

	err := r.db.
		Where("session_id = ?", sessionID).
		First(&pref).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &pref, nil
}

func (r *sessionRepository) ReplaceSuggestions(
	sessionID uint,
	suggestions []entity.Suggestion,
) error {
	if sessionID == 0 {
		return ErrInvalidState
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("session_id = ?", sessionID).
			Delete(&entity.Suggestion{}).
			Error; err != nil {
			return ErrInternal
		}

		if len(suggestions) == 0 {
			return nil
		}

		for i := range suggestions {
			suggestions[i].ID = 0
			suggestions[i].SessionID = sessionID
		}

		if err := tx.Create(&suggestions).Error; err != nil {
			if isDup(err) || isFK(err) {
				return ErrConflict
			}
			return ErrInternal
		}

		return nil
	})
}

func (r *sessionRepository) ListSuggestions(sessionID uint) ([]entity.Suggestion, error) {
	if sessionID == 0 {
		return nil, ErrNotFound
	}

	var suggestions []entity.Suggestion

	err := r.db.
		Preload("Bean").
		Preload("Recipe").
		Preload("Item").
		Preload("Item.Source").
		Where("session_id = ?", sessionID).
		Order("rank ASC").
		Order("id ASC").
		Find(&suggestions).
		Error
	if err != nil {
		return nil, ErrInternal
	}

	return suggestions, nil
}

func (r *sessionRepository) GetSuggestionByID(id uint) (*entity.Suggestion, error) {
	if id == 0 {
		return nil, ErrNotFound
	}

	var suggestion entity.Suggestion

	err := r.db.
		Preload("Bean").
		Preload("Recipe").
		Preload("Item").
		Preload("Item.Source").
		First(&suggestion, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &suggestion, nil
}
