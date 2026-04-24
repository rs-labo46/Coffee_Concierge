package repository

import (
	"coffee-spa/apperr"
	"coffee-spa/entity"

	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(log *entity.AuditLog) error
	List(q AuditListQ) ([]entity.AuditLog, error)
}

type AuditListQ struct {
	Type   string
	UserID *uint
	Limit  int
	Offset int
}
type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

// audit_logsに1件保存する。
func (r *auditRepository) Create(log *entity.AuditLog) error {
	if log == nil {
		return apperr.ErrInvalidState
	}

	err := r.db.Create(log).Error
	if err != nil {
		if isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// 監査ログ一覧
func (r *auditRepository) List(q AuditListQ) ([]entity.AuditLog, error) {
	var logs []entity.AuditLog

	tx := r.db.Model(&entity.AuditLog{})

	if q.Type != "" {
		tx = tx.Where("type = ?", q.Type)
	}

	if q.UserID != nil {
		tx = tx.Where("user_id = ?", *q.UserID)
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	err := tx.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	return logs, nil
}
