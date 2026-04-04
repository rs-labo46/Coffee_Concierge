package controller

import (
	"net/http"
	"strconv"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type ItemCtl struct {
	uc usecase.ItemUsecase
}

func NewItemCtl(uc usecase.ItemUsecase) ItemCtl {
	return ItemCtl{
		uc: uc,
	}
}

// item作成入力。
type AddItemReq struct {
	Title       string  `json:"title"`
	Summary     *string `json:"summary"`
	Body        *string `json:"body"`
	URL         *string `json:"url"`
	ImageURL    *string `json:"image_url"`
	Kind        string  `json:"kind"`
	SourceID    int64   `json:"source_id"`
	PublishedAt string  `json:"published_at"`
}

// item単体のレスポンス。
type ItemRes struct {
	Item entity.Item `json:"item"`
}

// item詳細レスポンス。
type ItemDetailRes struct {
	Item   entity.Item   `json:"item"`
	Source entity.Source `json:"source"`
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

// GET /items/topを処理。
func (ctl ItemCtl) Top(c echo.Context) error {
	limit, err := qInt(c, "limit", 3)
	if err != nil {
		return writeErr(c, err)
	}

	topItems, err := ctl.uc.Top(limit)
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, TopItemsRes{
		News:   topItems.News,
		Recipe: topItems.Recipe,
		Deal:   topItems.Deal,
		Shop:   topItems.Shop,
	})
}

// GET /items/:idを処理。
func (ctl ItemCtl) Detail(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return writeErr(c, usecase.ErrInvalidRequest)
	}

	item, err := ctl.uc.Get(id)
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, ItemDetailRes{
		Item:   item,
		Source: item.Source,
	})
}

// GET /itemsを処理。
func (ctl ItemCtl) List(c echo.Context) error {
	limit, err := qInt(c, "limit", 20)
	if err != nil {
		return writeErr(c, err)
	}
	offset, err := qInt(c, "offset", 0)
	if err != nil {
		return writeErr(c, err)
	}

	items, err := ctl.uc.Search(usecase.ItemQ{
		Q:      c.QueryParam("q"),
		Kind:   c.QueryParam("kind"),
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

// POST /itemsを処理。
// 認可はmiddleware側で行う。
func (ctl ItemCtl) Create(c echo.Context) error {
	var req AddItemReq

	if err := bindJSON(c, &req); err != nil {
		return writeErr(c, err)
	}

	actor := actorFromCtx(c)

	item, err := ctl.uc.Add(actor, usecase.AddItemIn{
		Title:       req.Title,
		Summary:     req.Summary,
		Body:        req.Body,
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
