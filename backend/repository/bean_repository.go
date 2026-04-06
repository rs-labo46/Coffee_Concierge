package repository

import (
	"errors"
	"fmt"
	"strings"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

type beanRepository struct {
	db *gorm.DB
}

func NewBeanRepository(db *gorm.DB) BeanRepository {
	return &beanRepository{
		db: db,
	}
}

func (r *beanRepository) Create(bean *entity.Bean) error {
	if bean == nil {
		return ErrInvalidState
	}

	err := r.db.Create(bean).Error
	if err != nil {
		if isDup(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *beanRepository) Update(bean *entity.Bean) error {
	if bean == nil || bean.ID == 0 {
		return ErrInvalidState
	}

	err := r.db.
		Model(&entity.Bean{}).
		Where("id = ?", bean.ID).
		Select(
			"name",
			"roast",
			"origin",
			"flavor",
			"acidity",
			"bitterness",
			"body",
			"aroma",
			"desc",
			"buy_url",
			"active",
			"updated_at",
		).
		Updates(bean).
		Error
	if err != nil {
		if isDup(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *beanRepository) GetByID(id uint) (*entity.Bean, error) {
	if id == 0 {
		return nil, ErrNotFound
	}

	var bean entity.Bean

	err := r.db.First(&bean, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &bean, nil
}

func (r *beanRepository) List(q BeanListQ) ([]entity.Bean, error) {
	var beans []entity.Bean

	tx := r.db.Model(&entity.Bean{})

	if q.Q != "" {
		like := "%" + q.Q + "%"
		tx = tx.Where(
			"name ILIKE ? OR origin ILIKE ? OR desc ILIKE ?",
			like,
			like,
			like,
		)
	}

	if q.Roast != "" {
		tx = tx.Where("roast = ?", q.Roast)
	}

	if q.Active != nil {
		tx = tx.Where("active = ?", *q.Active)
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

	err := tx.
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&beans).
		Error
	if err != nil {
		return nil, ErrInternal
	}

	return beans, nil
}

func (r *beanRepository) SearchByPref(pref entity.Pref, limit int) ([]entity.Bean, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var beans []entity.Bean

	tx := r.db.
		Model(&entity.Bean{}).
		Where("active = ?", true)

	tx = applyBeanExcludes(tx, pref.Excludes)

	orderExpr := fmt.Sprintf(
		"(ABS(flavor - %d) + ABS(acidity - %d) + ABS(bitterness - %d) + ABS(body - %d) + ABS(aroma - %d)) ASC",
		pref.Flavor,
		pref.Acidity,
		pref.Bitterness,
		pref.Body,
		pref.Aroma,
	)

	err := tx.
		Order(orderExpr).
		Order("id ASC").
		Limit(limit).
		Find(&beans).
		Error
	if err != nil {
		return nil, ErrInternal
	}

	return beans, nil
}

func applyBeanExcludes(tx *gorm.DB, excludes []string) *gorm.DB {
	if len(excludes) == 0 {
		return tx
	}

	seen := make(map[string]struct{}, len(excludes))
	for _, raw := range excludes {
		v := strings.TrimSpace(strings.ToLower(raw))
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}

		switch v {
		case "acidic":
			tx = tx.Where("acidity <= ?", 3)
		case "bitter":
			tx = tx.Where("bitterness <= ?", 3)
		case "dark_roast":
			tx = tx.Where("roast <> ?", entity.RoastDark)
		}
	}

	return tx
}
