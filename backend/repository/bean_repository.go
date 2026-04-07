package repository

import (
	"errors"
	"fmt"
	"strings"

	"coffee-spa/entity"

	"gorm.io/gorm"
)

// 豆データの保存・更新・取得・一覧・条件検索。
type beanRepository struct {
	db *gorm.DB
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
		return ErrInvalidState
	}

	// レコードをそのままINSERT 。
	err := r.db.Create(bean).Error
	if err != nil {
		// unique制約違反はconflict。
		if isDup(err) {
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

// beanを更新。
func (r *beanRepository) Update(bean *entity.Bean) error {
	// nilまたはIDの未設定は、更新対象を特定できないため不正。
	if bean == nil || bean.ID == 0 {
		return ErrInvalidState
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
			return ErrConflict
		}
		return ErrInternal
	}

	return nil
}

// IDでbeanを1件取得。
func (r *beanRepository) GetByID(id uint) (*entity.Bean, error) {
	// 0 は有効な ID ではないため not found として扱います。
	if id == 0 {
		return nil, ErrNotFound
	}

	var bean entity.Bean

	// 主キー検索で1件取得。
	err := r.db.First(&bean, id).Error
	if err != nil {
		// レコード未存在はnot found。
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
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
		return nil, ErrInternal
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
		return nil, ErrInternal
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
			// 酸味を避けたいので acidity が高すぎるもの
			tx = tx.Where("acidity <= ?", 3)
		case "bitter":
			// 苦味を避けたいので bitterness が高すぎるもの
			tx = tx.Where("bitterness <= ?", 3)
		case "dark_roast":
			// 深煎りを避けたいので dark
			tx = tx.Where("roast <> ?", entity.RoastDark)
		}
	}

	return tx
}
