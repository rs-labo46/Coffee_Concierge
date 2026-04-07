package repository

import (
	"errors"
	"fmt"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

// レシピの保存・更新・取得・一覧・レシピ選定。
type recipeRepository struct {
	db *gorm.DB
}

func NewRecipeRepository(db *gorm.DB) RecipeRepository {
	return &recipeRepository{
		db: db,
	}
}

// recipeを新規作成。
func (r *recipeRepository) Create(recipe *entity.Recipe) error {
	// nilは保存対象がないため不正状態。
	if recipe == nil {
		return ErrInvalidState
	}

	// レコードをINSERT。
	err := r.db.Create(recipe).Error
	if err != nil {
		// unique / FK制約違反はconflict。
		if isDup(err) || isFK(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

// recipeを更新。
func (r *recipeRepository) Update(recipe *entity.Recipe) error {
	// nilまたはID未設定は更新対象を特定できないため不正。
	if recipe == nil || recipe.ID == 0 {
		return ErrInvalidState
	}

	// 更新対象カラムを明示して更新。
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

// IDでrecipeを1件取得。
func (r *recipeRepository) GetByID(id uint) (*entity.Recipe, error) {
	if id == 0 {
		return nil, ErrNotFound
	}

	var recipe entity.Recipe

	// 参照側でBean情報も使いやすいようにpreload 。
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

// recipe一覧を取得。
func (r *recipeRepository) List(q RecipeListQ) ([]entity.Recipe, error) {
	var recipes []entity.Recipe

	// Recipeテーブルを起点にして、必要なBean先読み取り 。
	tx := r.db.
		Model(&entity.Recipe{}).
		Preload("Bean")

	// beanIDが指定されていれば、そのBeanに紐づくrecipeのみに絞る。
	if q.BeanID != nil {
		tx = tx.Where("bean_id = ?", *q.BeanID)
	}

	// method指定があれば抽出方法で絞る。
	if q.Method != "" {
		tx = tx.Where("method = ?", q.Method)
	}

	// tempPref指定があれば温度で絞る。
	if q.TempPref != "" {
		tx = tx.Where("temp_pref = ?", q.TempPref)
	}

	// active指定があれば公開状態で絞る。
	if q.Active != nil {
		tx = tx.Where("active = ?", *q.Active)
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

	// 一覧の順序が安定するようbean_id → idの順で。
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

// 1つのBeanに対して最も適した主レシピを1件返す。
// 優先順位は、
// 1. method と temp_pref の両方一致
// 2. method 一致
// 3. temp_pref 一致
// 4. それ以外
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

	// activeなrecipeから優先順位順で1件選ぶ。
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
