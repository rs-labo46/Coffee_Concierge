package controller

import (
	"net/http"
	"time"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type ItemCtl struct {
	uc usecase.ItemUC
}

func NewItemCtl(uc usecase.ItemUC) *ItemCtl {
	return &ItemCtl{
		uc: uc,
	}
}

// item作成入力。
type CreateItemReq struct {
	Title       string          `json:"title"`
	Summary     string          `json:"summary"`
	URL         string          `json:"url"`
	ImageURL    string          `json:"image_url"`
	Kind        entity.ItemKind `json:"kind"`
	SourceID    uint            `json:"source_id"`
	PublishedAt time.Time       `json:"published_at"`
}

// item単体のレスポンス。
type ItemRes struct {
	Item entity.Item `json:"item"`
}

// item一覧のレスポンス。
type ItemListRes struct {
	Items []entity.Item `json:"items"`
}

// top itemsのレスポンス。
type TopItemsRes struct {
	News   []entity.Item `json:"news"`
	Recipe []entity.Item `json:"recipe"`
	Deal   []entity.Item `json:"deal"`
	Shop   []entity.Item `json:"shop"`
}

// // item詳細レスポンス。
// type ItemDetailRes struct {
// 	Item   entity.Item   `json:"item"`
// 	Source entity.Source `json:"source"`
// }

// GET /items/topを処理(query:limit)。
func (ctl *ItemCtl) Top(c echo.Context) error {
	limit, err := qInt(c, "limit", 3)
	if err != nil {
		return writeErr(c, err)
	}

	top, err := ctl.uc.Top(limit)
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, TopItemsRes{
		News:   top.News,
		Recipe: top.Recipe,
		Deal:   top.Deal,
		Shop:   top.Shop,
	})
}

// GET /itemsを処理(query: q,kind,limit,offset)。
func (ctl *ItemCtl) List(c echo.Context) error {
	limit, err := qInt(c, "limit", 20)
	if err != nil {
		return writeErr(c, err)
	}
	offset, err := qInt(c, "offset", 0)
	if err != nil {
		return writeErr(c, err)
	}

	items, err := ctl.uc.List(entity.ItemQ{
		Q:      qStr(c, "q"),
		Kind:   entity.ItemKind(qStr(c, "kind")),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, ItemListRes{
		Items: items,
	})
}

// POST /itemsを処理。管理者のみ
func (ctl ItemCtl) Create(c echo.Context) error {
	var req CreateItemReq

	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	item, err := ctl.uc.Create(*actor, usecase.CreateItemIn{
		Title:       req.Title,
		Summary:     req.Summary,
		URL:         req.URL,
		ImageURL:    req.ImageURL,
		Kind:        req.Kind,
		SourceID:    req.SourceID,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusCreated, ItemRes{
		Item: item,
	})
}

// GET /items/:idを処理
func (ctl *ItemCtl) Get(c echo.Context) error {
	id, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}
	item, err := ctl.uc.Get(id)
	if err != nil {
		return writeErr(c, err)
	}
	return c.JSON(http.StatusOK, ItemRes{
		Item: item,
	})
}
