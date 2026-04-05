package repository

import (
	"errors"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

type sourceRepository struct {
	db *gorm.DB
}

func NewSourceRepository(db *gorm.DB) SourceRepository {
	return &sourceRepository{db: db}
}

// sourcesに1件保存する。

func (r *sourceRepository) Create(src *entity.Source) error {
	if src == nil {
		return ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(src).Error
	if err != nil {
		if isDup(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

// sourcesを1件取得する。
func (r *sourceRepository) GetByID(id uint) (*entity.Source, error) {
	// 0は不正ID。
	if id == 0 {
		return nil, ErrNotFound
	}
	var src entity.Source

	// 主キー検索を行う。
	err := r.db.First(&src, id).Error
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &src, nil
}

// sourcesの一覧を返す。
func (r *sourceRepository) List(q SourceListQ) ([]entity.Source, error) {

	var sources []entity.Source
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

	// 一覧取得を行う。
	err := r.db.Model(&entity.Source{}).Order("id ASC").Limit(limit).Offset(offset).Find(&sources).Error
	if err != nil {
		return nil, ErrInternal
	}

	return sources, nil
}
