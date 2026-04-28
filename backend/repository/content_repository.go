package repository

import (
	"coffee-spa/apperr"
	"coffee-spa/entity"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ItemRepository interface {
	Create(item *entity.Item) error
	GetByID(id uint) (*entity.Item, error)
	List(q ItemListQ) ([]entity.Item, error)
	Top(limit int) (*entity.TopItems, error)
	SearchRelated(
		beanName string,
		roast entity.Roast,
		origin string,
		mood entity.Mood,
		method entity.Method,
		limit int,
		now time.Time,
	) ([]entity.Item, error)
}

type SourceRepository interface {
	Create(src *entity.Source) error
	GetByID(id uint) (*entity.Source, error)
	List(q SourceListQ) ([]entity.Source, error)
}

type SourceListQ struct {
	Limit  int
	Offset int
}

type ItemListQ struct {
	Q      string
	Kind   entity.ItemKind
	Limit  int
	Offset int
}

type sourceRepository struct {
	db *gorm.DB
}
type itemRepository struct {
	db *gorm.DB
}

func NewSourceRepository(db *gorm.DB) SourceRepository {
	return &sourceRepository{
		db: db,
	}
}

func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{
		db: db,
	}
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
		return nil, apperr.ErrInternal
	}

	return sources, nil
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

func (r *itemRepository) List(q ItemListQ) ([]entity.Item, error) {
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

	// Bean名や英語enumだけでは日本語seedのitemsに当たりにくいため、
	// roast / mood / method から日本語の関連語も追加する。
	rawTerms := []string{
		beanName,
		string(roast),
		origin,
		string(mood),
		string(method),
	}
	rawTerms = append(rawTerms, roastRelatedTerms(roast)...)
	rawTerms = append(rawTerms, moodRelatedTerms(mood)...)
	rawTerms = append(rawTerms, methodRelatedTerms(method)...)

	terms := uniqueNonEmpty(rawTerms...)

	var items []entity.Item

	// ベースクエリ。
	tx := r.db.
		Model(&entity.Item{}).
		Preload("Source").
		Where("published_at <= ?", now)

	// 検索語がある場合だけ、title / summary の部分一致条件を追加する。
	// OR条件は1つのWHERE句にまとめ、published_at条件との優先順位崩れを避ける。
	if len(terms) > 0 {
		conds := make([]string, 0, len(terms))
		args := make([]interface{}, 0, len(terms)*2)

		for _, term := range terms {
			like := "%" + term + "%"
			conds = append(conds, "(title ILIKE ? OR summary ILIKE ?)")
			args = append(args, like, like)
		}

		tx = tx.Where("("+strings.Join(conds, " OR ")+")", args...)
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

	// 関連語で1件も当たらない場合でも、公開中Itemを返して既存SPAへの回遊導線を残す。
	if len(items) == 0 {
		return fallbackRelatedItems(r.db, limit, now)
	}

	return items, nil
}

// 焙煎度を日本語の関連語へ変換する。
// items.title / items.summary が日本語seed中心のため、英語enumだけでは検索に当たりにくい。
func roastRelatedTerms(roast entity.Roast) []string {
	switch roast {
	case entity.RoastLight:
		return []string{
			"浅煎り",
			"フローラル",
			"果実感",
			"透明感",
			"エチオピア",
			"柑橘",
		}
	case entity.RoastMedium:
		return []string{
			"中煎り",
			"バランス",
			"家庭向け",
			"甘さ",
			"基本比率",
			"飲みやすい",
		}
	case entity.RoastDark:
		return []string{
			"深煎り",
			"苦味",
			"濃い",
			"牛乳",
			"ラテ",
			"厚み",
		}
	default:
		return nil
	}
}

// 気分を日本語の関連語へ変換する。
// mood enumだけでは日本語記事に当たりにくいため、記事タイトルに出やすい語へ広げる。
func moodRelatedTerms(mood entity.Mood) []string {
	switch mood {
	case entity.MoodMorning:
		return []string{
			"朝",
			"朝の一杯",
			"時短",
			"朝食",
			"出勤前",
		}
	case entity.MoodWork:
		return []string{
			"作業",
			"集中",
			"デスクワーク",
			"長時間",
			"仕事",
		}
	case entity.MoodRelax:
		return []string{
			"休日",
			"週末",
			"ゆっくり",
			"香り",
			"リラックス",
		}
	case entity.MoodNight:
		return []string{
			"夜",
			"深夜",
			"食後",
			"落ち着き",
		}
	default:
		return nil
	}
}

// 抽出方法を日本語の関連語へ変換する。
// method enumだけでは日本語記事に当たりにくいため、記事タイトルに出やすい語へ広げる。
func methodRelatedTerms(method entity.Method) []string {
	switch method {
	case entity.MethodDrip:
		return []string{
			"ドリップ",
			"ハンドドリップ",
			"ペーパー",
			"抽出",
		}
	case entity.MethodEspresso:
		return []string{
			"エスプレッソ",
			"濃縮",
			"ラテ",
		}
	case entity.MethodMilk:
		return []string{
			"ミルク",
			"牛乳",
			"ラテ",
			"オレ",
		}
	case entity.MethodIced:
		return []string{
			"アイス",
			"冷たい",
			"氷",
			"夏",
		}
	default:
		return nil
	}
}

// 関連語検索で0件だった場合の保険として、公開中Itemをkind優先順で返す。
// これにより、コンシェルジュ画面で関連情報導線が完全に空になることを避ける。
func fallbackRelatedItems(db *gorm.DB, limit int, now time.Time) ([]entity.Item, error) {
	if limit <= 0 {
		limit = 3
	}
	if limit > 3 {
		limit = 3
	}

	var items []entity.Item

	kindOrder := `
		CASE kind
			WHEN 'recipe' THEN 1
			WHEN 'shop' THEN 2
			WHEN 'news' THEN 3
			WHEN 'deal' THEN 4
			ELSE 5
		END
	`

	err := db.
		Model(&entity.Item{}).
		Preload("Source").
		Where("published_at <= ?", now).
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
