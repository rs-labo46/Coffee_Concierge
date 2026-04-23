package repository

import (
	"errors"
	"strings"
	"time"

	"coffee-spa/apperr"
	"coffee-spa/entity"
	"coffee-spa/usecase/port"

	"gorm.io/gorm"
)

type sourceRepository struct {
	db *gorm.DB
}
type itemRepository struct {
	db *gorm.DB
}

func NewSourceRepository(db *gorm.DB) port.SourceRepository {
	return &sourceRepository{db: db}
}

// sourcesに1件保存する。
func (r *sourceRepository) Create(src *entity.Source) error {
	if src == nil {
		return apperr.ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(src).Error
	if err != nil {
		if isDup(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// sourcesを1件取得する。
func (r *sourceRepository) GetByID(id uint) (*entity.Source, error) {
	// 0は不正ID。
	if id == 0 {
		return nil, apperr.ErrNotFound
	}
	var src entity.Source

	// 主キー検索を行う。
	err := r.db.First(&src, id).Error
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &src, nil
}

// sourcesの一覧を返す。
func (r *sourceRepository) List(q port.SourceListQ) ([]entity.Source, error) {

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
		return nil, apperr.ErrInternal
	}

	return sources, nil
}

func NewItemRepository(db *gorm.DB) port.ItemRepository {
	return &itemRepository{db}
}

func (r *itemRepository) Create(item *entity.Item) error {
	if item == nil {
		return apperr.ErrInvalidState
	}

	// INSERTを実行する。
	err := r.db.Create(item).Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// 公開/非公開を問わずIDで1件取得
func (r *itemRepository) GetByID(id uint) (*entity.Item, error) {
	if id == 0 {
		return nil, apperr.ErrNotFound
	}
	var item entity.Item

	err := r.db.Preload("Source").First(&item, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound

		}

		return nil, apperr.ErrInternal
	}
	return &item, nil
}

func (r *itemRepository) List(q port.ItemListQ) ([]entity.Item, error) {
	var items []entity.Item

	//公開中のアイテムクエリ
	tx := r.db.Model(&entity.Item{}).Preload("Source").Where("published_at <= ?", time.Now())
	//kindの指定がある場合
	if q.Kind != "" {
		tx = tx.Where("kind = ?", q.Kind)
	}
	// qの指定がある場合の部分検索
	if q.Q != "" {
		like := "%" + q.Q + "%"
		tx = tx.Where(
			"title ILIKE ? OR summary ILIKE ?",
			like,
			like,
		)
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	// 公開日時が新しい順で返す。
	err := tx.
		Order("published_at DESC").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	return items, nil

}

func (r *itemRepository) Top(limit int) (*entity.TopItems, error) {
	if limit <= 0 {
		limit = 3
	}
	if limit > 50 {
		limit = 50
	}

	// 戻り値を初期化する。
	top := &entity.TopItems{
		News:   []entity.Item{},
		Recipe: []entity.Item{},
		Deal:   []entity.Item{},
		Shop:   []entity.Item{},
	}

	// kindごとに同じ条件で取得する。
	queries := []struct {
		kind entity.ItemKind
		dst  *[]entity.Item
	}{
		{kind: entity.ItemKindNews, dst: &top.News},
		{kind: entity.ItemKindRecipe, dst: &top.Recipe},
		{kind: entity.ItemKindDeal, dst: &top.Deal},
		{kind: entity.ItemKindShop, dst: &top.Shop},
	}

	// 各カテゴリを順番に取得する。
	for _, q := range queries {
		err := r.db.Model(&entity.Item{}).
			Preload("Source").
			Where("kind = ?", q.kind).
			Where("published_at <= ?", time.Now()).
			Order("published_at DESC").
			Order("created_at DESC").
			Limit(limit).
			Find(q.dst).
			Error
		if err != nil {
			return nil, apperr.ErrInternal
		}
	}

	return top, nil
}

func (r *itemRepository) SearchRelated(
	beanName string,
	roast entity.Roast,
	origin string,
	mood entity.Mood,
	method entity.Method,
	limit int,
	now time.Time,
) ([]entity.Item, error) {

	if limit <= 0 {
		limit = 3
	}
	if limit > 3 {
		limit = 3
	}

	// 候補語を作る。
	// 空文字は除外し、重複も避ける。
	terms := uniqueNonEmpty(
		beanName,
		string(roast),
		origin,
		string(mood),
		string(method),
	)

	var items []entity.Item

	// ベースクエリ。
	tx := r.db.
		Model(&entity.Item{}).
		Preload("Source").
		Where("published_at <= ?", now)

	// 検索語が1つもない場合は、公開中Itemをkindの優先順で返す。
	if len(terms) > 0 {
		// 最初の検索語でベース条件を作る。
		firstLike := "%" + terms[0] + "%"
		tx = tx.Where(
			"(title ILIKE ? OR summary ILIKE ?)",
			firstLike,
			firstLike,
		)

		// 2語目以降は OR 条件を順に追加する。
		for _, term := range terms[1:] {
			like := "%" + term + "%"
			tx = tx.Or(
				"(title ILIKE ? OR summary ILIKE ?)",
				like,
				like,
			)
		}
	}

	// kindの優先順をCASE式で表現する。
	kindOrder := `
		CASE kind
			WHEN 'recipe' THEN 1
			WHEN 'shop' THEN 2
			WHEN 'news' THEN 3
			WHEN 'deal' THEN 4
			ELSE 5
		END
	`
	err := tx.
		Order(kindOrder).
		Order("published_at DESC").
		Order("created_at DESC").
		Limit(limit).
		Find(&items).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	return items, nil
}

// 空文字を除きつつ、重複しない文字列配列を返す。
func uniqueNonEmpty(xs ...string) []string {
	// 重複判定map。
	seen := make(map[string]struct{})
	// 結果配列。
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		// 前後空白を落とす。
		v := strings.TrimSpace(x)

		// 空文字は除外。
		if v == "" {
			continue
		}

		// すでに入っていれば飛ばす。
		if _, ok := seen[v]; ok {
			continue
		}

		// 初登場なら導入。
		seen[v] = struct{}{}
		out = append(out, v)
	}

	return out
}
