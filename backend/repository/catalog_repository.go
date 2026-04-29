package repository

import (
	"coffee-spa/apperr"
	"coffee-spa/entity"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type BeanRepository interface {
	Create(bean *entity.Bean) error
	Update(bean *entity.Bean) error
	GetByID(id uint) (*entity.Bean, error)
	List(q BeanListQ) ([]entity.Bean, error)
	SearchByPref(pref entity.Pref, limit int) ([]entity.Bean, error)
}

type RecipeRepository interface {
	Create(recipe *entity.Recipe) error
	Update(recipe *entity.Recipe) error
	GetByID(id uint) (*entity.Recipe, error)
	List(q RecipeListQ) ([]entity.Recipe, error)
	FindPrimaryByBean(beanID uint, method entity.Method, tempPref entity.TempPref) (*entity.Recipe, error)
}

// 豆データの保存・更新・取得・一覧・条件検索。
type beanRepository struct {
	db *gorm.DB
}

type BeanListQ struct {
	Q      string
	Roast  entity.Roast
	Active *bool
	Limit  int
	Offset int
}

type RecipeListQ struct {
	BeanID   *uint
	Method   entity.Method
	TempPref entity.TempPref
	Active   *bool
	Limit    int
	Offset   int
}

// レシピの保存・更新・取得・一覧・レシピ選定。
type recipeRepository struct {
	db *gorm.DB
}

type recipeRow struct {
	ID        uint           `gorm:"column:id"`
	BeanID    uint           `gorm:"column:bean_id"`
	Name      string         `gorm:"column:name"`
	Method    string         `gorm:"column:method"`
	TempPref  string         `gorm:"column:temp_pref"`
	Grind     string         `gorm:"column:grind"`
	Ratio     string         `gorm:"column:ratio"`
	Temp      int            `gorm:"column:temp"`
	TimeSec   int            `gorm:"column:time_sec"`
	Steps     pq.StringArray `gorm:"column:steps;type:text[]"`
	Desc      string         `gorm:"column:desc"`
	Active    bool           `gorm:"column:active"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
}

func recipeRowFromEntity(recipe *entity.Recipe) recipeRow {
	return recipeRow{
		ID:        recipe.ID,
		BeanID:    recipe.BeanID,
		Name:      recipe.Name,
		Method:    string(recipe.Method),
		TempPref:  string(recipe.TempPref),
		Grind:     recipe.Grind,
		Ratio:     recipe.Ratio,
		Temp:      recipe.Temp,
		TimeSec:   recipe.TimeSec,
		Steps:     pq.StringArray(recipe.Steps),
		Desc:      recipe.Desc,
		Active:    recipe.Active,
		CreatedAt: recipe.CreatedAt,
		UpdatedAt: recipe.UpdatedAt,
	}
}

func recipeRowToEntity(row recipeRow) entity.Recipe {
	return entity.Recipe{
		ID:        row.ID,
		BeanID:    row.BeanID,
		Name:      row.Name,
		Method:    entity.Method(row.Method),
		TempPref:  entity.TempPref(row.TempPref),
		Grind:     row.Grind,
		Ratio:     row.Ratio,
		Temp:      row.Temp,
		TimeSec:   row.TimeSec,
		Steps:     []string(row.Steps),
		Desc:      row.Desc,
		Active:    row.Active,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func NewBeanRepository(db *gorm.DB) BeanRepository {
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
func (r *beanRepository) List(q BeanListQ) ([]entity.Bean, error) {
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

	// limitが未指定なら20、上限は1000に固定。
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 1000 {
		limit = 1000
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

func NewRecipeRepository(db *gorm.DB) RecipeRepository {
	return &recipeRepository{
		db: db,
	}
}

// recipeを新規作成。
func (r *recipeRepository) Create(recipe *entity.Recipe) error {
	if recipe == nil {
		return apperr.ErrInvalidState
	}

	row := recipeRowFromEntity(recipe)

	err := r.db.Table("recipes").Create(&row).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	recipe.ID = row.ID
	recipe.CreatedAt = row.CreatedAt
	recipe.UpdatedAt = row.UpdatedAt

	return nil
}

// recipeを更新。
func (r *recipeRepository) Update(recipe *entity.Recipe) error {
	if recipe == nil || recipe.ID == 0 {
		return apperr.ErrInvalidState
	}

	row := recipeRowFromEntity(recipe)

	err := r.db.
		Table("recipes").
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
		Updates(&row).
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

	var row recipeRow

	err := r.db.
		Table("recipes").
		Where("id = ?", id).
		First(&row).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	recipe := recipeRowToEntity(row)

	var bean entity.Bean
	err = r.db.First(&bean, recipe.BeanID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrInternal
		}
		return nil, apperr.ErrInternal
	}

	recipe.Bean = bean

	return &recipe, nil
}

// recipe一覧を取得。
func (r *recipeRepository) List(q RecipeListQ) ([]entity.Recipe, error) {
	var rows []recipeRow

	tx := r.db.Table("recipes")

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
		Find(&rows).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	recipes := make([]entity.Recipe, 0, len(rows))
	beanIDs := make([]uint, 0, len(rows))
	seen := map[uint]struct{}{}

	for _, row := range rows {
		recipe := recipeRowToEntity(row)
		recipes = append(recipes, recipe)

		if _, ok := seen[recipe.BeanID]; ok {
			continue
		}

		seen[recipe.BeanID] = struct{}{}
		beanIDs = append(beanIDs, recipe.BeanID)
	}

	if len(beanIDs) == 0 {
		return recipes, nil
	}

	var beans []entity.Bean
	if err := r.db.Where("id IN ?", beanIDs).Find(&beans).Error; err != nil {
		return nil, apperr.ErrInternal
	}

	beanByID := make(map[uint]entity.Bean, len(beans))
	for _, bean := range beans {
		beanByID[bean.ID] = bean
	}

	for i := range recipes {
		if bean, ok := beanByID[recipes[i].BeanID]; ok {
			recipes[i].Bean = bean
		}
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

	var row recipeRow

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
		Table("recipes").
		Where("bean_id = ?", beanID).
		Where("active = ?", true).
		Order(orderExpr).
		Order("id ASC").
		First(&row).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	recipe := recipeRowToEntity(row)

	var bean entity.Bean
	err = r.db.First(&bean, recipe.BeanID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrInternal
		}
		return nil, apperr.ErrInternal
	}

	recipe.Bean = bean

	return &recipe, nil
}
