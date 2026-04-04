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

func (r *auditRepository) Create(a entity.AuditLog) error {
	err := r.db.Create(&a).Error
	if err != nil {
		if isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}
