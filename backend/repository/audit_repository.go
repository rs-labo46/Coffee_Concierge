package repository

import (
	"coffee-spa/entity"

	"gorm.io/gorm"
)

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

// audit_logsに1件保存する。
func (r *auditRepository) Create(log *entity.AuditLog) error {
	if log == nil {
		return ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(log).Error
	if err != nil {
		if isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

// 監査ログ一覧
func (r *auditRepository) List(q AuditListQ) ([]entity.AuditLog, error) {
	var logs []entity.AuditLog

	// ベースクエリを作る。
	tx := r.db.
		Model(&entity.AuditLog{})

	// type指定がある場合は絞り込む。
	if q.Type != "" {
		tx = tx.Where("type = ?", q.Type)
	}

	// user_id 指定がある場合は絞り込む。
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

	// 一覧取得を行う。
	err := tx.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).
		Error
	if err != nil {
		return nil, ErrInternal
	}

	return logs, nil
}
