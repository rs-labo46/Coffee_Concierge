package repository

import (
	"errors"
	"fmt"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

type recipeRepository struct {
	db *gorm.DB
}

func NewRecipeRepository(db *gorm.DB) RecipeRepository {
	return &recipeRepository{
		db: db,
	}
}

func (r *recipeRepository) Create(recipe *entity.Recipe) error {
	if recipe == nil {
		return ErrInvalidState
	}

	err := r.db.Create(recipe).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *recipeRepository) Update(recipe *entity.Recipe) error {
	if recipe == nil || recipe.ID == 0 {
		return ErrInvalidState
	}

	err := r.db.
		Model(&entity.Recipe{}).
		Where("id = ?", recipe.ID).
		Select(
			"bean_id",
			"name",
			"method",
			"temp_pref",
			"grind",
			"ratio",
			"temp",
			"time_sec",
			"steps",
			"desc",
			"active",
			"updated_at",
		).
		Updates(recipe).
		Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

func (r *recipeRepository) GetByID(id uint) (*entity.Recipe, error) {
	if id == 0 {
		return nil, ErrNotFound
	}

	var recipe entity.Recipe

	err := r.db.
		Preload("Bean").
		First(&recipe, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &recipe, nil
}

func (r *recipeRepository) List(q RecipeListQ) ([]entity.Recipe, error) {
	var recipes []entity.Recipe

	tx := r.db.
		Model(&entity.Recipe{}).
		Preload("Bean")

	if q.BeanID != nil {
		tx = tx.Where("bean_id = ?", *q.BeanID)
	}

	if q.Method != "" {
		tx = tx.Where("method = ?", q.Method)
	}

	if q.TempPref != "" {
		tx = tx.Where("temp_pref = ?", q.TempPref)
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
		Order("bean_id ASC").
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&recipes).
		Error
	if err != nil {
		return nil, ErrInternal
	}

	return recipes, nil
}

func (r *recipeRepository) FindPrimaryByBean(
	beanID uint,
	method entity.Method,
	tempPref entity.TempPref,
) (*entity.Recipe, error) {
	if beanID == 0 {
		return nil, ErrNotFound
	}

	var recipe entity.Recipe

	orderExpr := fmt.Sprintf(
		"CASE "+
			"WHEN method = '%s' AND temp_pref = '%s' THEN 0 "+
			"WHEN method = '%s' THEN 1 "+
			"WHEN temp_pref = '%s' THEN 2 "+
			"ELSE 3 END",
		string(method),
		string(tempPref),
		string(method),
		string(tempPref),
	)

	err := r.db.
		Model(&entity.Recipe{}).
		Preload("Bean").
		Where("bean_id = ?", beanID).
		Where("active = ?", true).
		Order(orderExpr).
		Order("id ASC").
		First(&recipe).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	return &recipe, nil
}
