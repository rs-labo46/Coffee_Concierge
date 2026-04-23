package repository

import (
	"errors"
	"fmt"
	"strings"

	"coffee-spa/apperr"
	"coffee-spa/entity"
	"coffee-spa/usecase/port"

	"gorm.io/gorm"
)

// 豆データの保存・更新・取得・一覧・条件検索。
type beanRepository struct {
	db *gorm.DB
}

// レシピの保存・更新・取得・一覧・レシピ選定。
type recipeRepository struct {
	db *gorm.DB
}

func NewBeanRepository(db *gorm.DB) port.BeanRepository {
	return &beanRepository{
		db: db,
	}
}

// beanを新規作成。
func (r *beanRepository) Create(bean *entity.Bean) error {
	// nilは保存対象が存在しないため不正状態。
	if bean == nil {
		return apperr.ErrInvalidState
	}

	// レコードをそのままINSERT 。
	err := r.db.Create(bean).Error
	if err != nil {
		// unique制約違反はconflict。
		if isDup(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// beanを更新。
func (r *beanRepository) Update(bean *entity.Bean) error {
	// nilまたはIDの未設定は、更新対象を特定できないため不正。
	if bean == nil || bean.ID == 0 {
		return apperr.ErrInvalidState
	}

	// 更新していいカラムだけを明示して更新。
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
		// unique制約違反はconflictに。
		if isDup(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// IDでbeanを1件取得。
func (r *beanRepository) GetByID(id uint) (*entity.Bean, error) {
	// 0 は有効な ID ではないため not found として扱います。
	if id == 0 {
		return nil, apperr.ErrNotFound
	}

	var bean entity.Bean

	// 主キー検索で1件取得。
	err := r.db.First(&bean, id).Error
	if err != nil {
		// レコード未存在はnot found。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &bean, nil
}

// Bean一覧を取得。
// q / roast / active / pagination条件に対応する。
func (r *beanRepository) List(q port.BeanListQ) ([]entity.Bean, error) {
	var beans []entity.Bean

	// Beanテーブルをもとに検索クエリ。
	tx := r.db.Model(&entity.Bean{})

	// 検索語がある場合は name / origin / desc を部分一致で絞る。
	if q.Q != "" {
		like := "%" + q.Q + "%"
		tx = tx.Where(
			"name ILIKE ? OR origin ILIKE ? OR desc ILIKE ?",
			like,
			like,
			like,
		)
	}

	// roastで指定がある場合は焙煎度で絞る。
	if q.Roast != "" {
		tx = tx.Where("roast = ?", q.Roast)
	}

	// activeで指定がある場合は公開状態で絞る。
	if q.Active != nil {
		tx = tx.Where("active = ?", *q.Active)
	}

	// limitが未指定なら20、上限は100に固定。
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

	// 一覧は安定した順序になるようIDで昇順に。
	err := tx.
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&beans).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	return beans, nil
}

// PrefをもとにBean候補を検索。
// 5軸の差分が小さい順に並べ、必要に応じてexcludesを適用。
func (r *beanRepository) SearchByPref(pref entity.Pref, limit int) ([]entity.Bean, error) {
	// limitが未指定なら10、上限は50に。
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var beans []entity.Bean

	// 検索対象はactive = trueのBeanに限定。
	tx := r.db.
		Model(&entity.Bean{}).
		Where("active = ?", true)

	// excludesに応じて検索条件を追加。
	tx = applyBeanExcludes(tx, pref.Excludes)

	// 5軸の差分合計が小さいほど上位になるようにする。
	orderExpr := fmt.Sprintf(
		"(ABS(flavor - %d) + ABS(acidity - %d) + ABS(bitterness - %d) + ABS(body - %d) + ABS(aroma - %d)) ASC",
		pref.Flavor,
		pref.Acidity,
		pref.Bitterness,
		pref.Body,
		pref.Aroma,
	)

	// 類似度順、その次にID昇順でソート。
	err := tx.
		Order(orderExpr).
		Order("id ASC").
		Limit(limit).
		Find(&beans).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	return beans, nil
}

// excludesをBeanの検索条件へ反映。
func applyBeanExcludes(tx *gorm.DB, excludes []string) *gorm.DB {
	// 空なら何も足さない。
	if len(excludes) == 0 {
		return tx
	}

	// 同じexcludesが複数入っても二重適用しないようにする。
	seen := make(map[string]struct{}, len(excludes))

	for _, raw := range excludes {
		// 大文字小文字や前後空白が空。
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
			// 酸味を避けたいのでacidityが高すぎるもの
			tx = tx.Where("acidity <= ?", 3)
		case "bitter":
			// 苦味を避けたいのでbitternessが高すぎるもの
			tx = tx.Where("bitterness <= ?", 3)
		case "dark_roast":
			// 深煎りを避けたいのでdark
			tx = tx.Where("roast <> ?", entity.RoastDark)
		}
	}

	return tx
}

func NewRecipeRepository(db *gorm.DB) port.RecipeRepository {
	return &recipeRepository{
		db: db,
	}
}

// recipeを新規作成。
func (r *recipeRepository) Create(recipe *entity.Recipe) error {
	// nilは保存対象がないため不正状態。
	if recipe == nil {
		return apperr.ErrInvalidState
	}

	// レコードをINSERT。
	err := r.db.Create(recipe).Error
	if err != nil {
		// unique / FK制約違反はconflict。
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// recipeを更新。
func (r *recipeRepository) Update(recipe *entity.Recipe) error {
	// nilまたはID未設定は更新対象を特定できないため不正。
	if recipe == nil || recipe.ID == 0 {
		return apperr.ErrInvalidState
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
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// IDでrecipeを1件取得。
func (r *recipeRepository) GetByID(id uint) (*entity.Recipe, error) {
	if id == 0 {
		return nil, apperr.ErrNotFound
	}

	var recipe entity.Recipe

	// 参照側でBean情報も使いやすいようにpreload。
	err := r.db.
		Preload("Bean").
		First(&recipe, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &recipe, nil
}

// recipe一覧を取得。
func (r *recipeRepository) List(q port.RecipeListQ) ([]entity.Recipe, error) {
	var recipes []entity.Recipe

	// Recipeテーブルを起点にして、必要なBean先読み取り。
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
		return nil, apperr.ErrInternal
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
		return nil, apperr.ErrNotFound
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
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &recipe, nil
}
