package usecase

import (
	"encoding/json"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

// Itemの作成・取得・一覧・Topを扱う。
type ItemUC interface {
	Create(actor entity.Actor, in CreateItemIn) (entity.Item, error)
	Get(id uint) (entity.Item, error)
	List(q string, kind entity.ItemKind, limit int, offset int) ([]entity.Item, error)
	Top(limit int) (entity.TopItems, error)
}

// Item作成入力。
type CreateItemIn struct {
	Title       string
	Summary     string
	URL         string
	ImageURL    string
	Kind        entity.ItemKind
	SourceID    uint
	PublishedAt time.Time
}

// Item 系のvalidator。
type ItemVal interface {
	NewItem(in CreateItemIn) error
	Get(id uint) error
	List(q string, kind entity.ItemKind, limit int, offset int) error
	Top(limit int) error
}

type itemUsecase struct {
	items   repository.ItemRepository
	sources repository.SourceRepository
	audits  repository.AuditRepository
	val     ItemVal
}

func NewItemUsecase(
	items repository.ItemRepository,
	sources repository.SourceRepository,
	audits repository.AuditRepository,
	val ItemVal,
) ItemUC {
	return &itemUsecase{
		items:   items,
		sources: sources,
		audits:  audits,
		val:     val,
	}
}

// Itemを新規作成する。
// adminのみ許可。
// source_idの存在確認を先に行う。
func (u *itemUsecase) Create(actor entity.Actor, in CreateItemIn) (entity.Item, error) {
	if actor.Role != entity.RoleAdmin {
		return entity.Item{}, repository.ErrForbidden
	}

	if err := u.val.NewItem(in); err != nil {
		return entity.Item{}, err
	}

	// sourceの存在確認を先に行う。
	if _, err := u.sources.GetByID(in.SourceID); err != nil {
		return entity.Item{}, err
	}

	item := &entity.Item{
		Title:       in.Title,
		Summary:     in.Summary,
		URL:         in.URL,
		ImageURL:    in.ImageURL,
		Kind:        in.Kind,
		SourceID:    in.SourceID,
		PublishedAt: in.PublishedAt,
	}

	if err := u.items.Create(item); err != nil {
		return entity.Item{}, err
	}

	u.writeAudit(
		"admin.items.create",
		&actor.UserID,
		map[string]any{
			"item_id":   item.ID,
			"source_id": item.SourceID,
			"kind":      item.Kind,
			"title":     item.Title,
		},
	)

	return *item, nil
}

// Itemを1件取得する。
func (u *itemUsecase) Get(id uint) (entity.Item, error) {
	if err := u.val.Get(id); err != nil {
		return entity.Item{}, err
	}

	item, err := u.items.GetByID(id)
	if err != nil {
		return entity.Item{}, err
	}

	return *item, nil
}

// Item一覧を返す。
func (u *itemUsecase) List(
	q string,
	kind entity.ItemKind,
	limit int,
	offset int,
) ([]entity.Item, error) {
	if err := u.val.List(q, kind, limit, offset); err != nil {
		return nil, err
	}

	out, err := u.items.List(repository.ItemListQ{
		Q:      q,
		Kind:   kind,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Top表示用のカテゴリ別アイテム
func (u *itemUsecase) Top(limit int) (entity.TopItems, error) {
	if err := u.val.Top(limit); err != nil {
		return entity.TopItems{}, err
	}

	top, err := u.items.Top(limit)
	if err != nil {
		return entity.TopItems{}, err
	}

	return *top, nil
}

func (u *itemUsecase) writeAudit(
	typ string,
	userID *uint,
	meta map[string]any,
) {
	if u.audits == nil {
		return
	}

	raw, err := json.Marshal(meta)
	if err != nil {
		raw = []byte(`{}`)
	}

	_ = u.audits.Create(&entity.AuditLog{
		Type:   typ,
		UserID: userID,
		IP:     "",
		UA:     "",
		Meta:   raw,
	})
}
