package controller

import (
	"net/http"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

// 依存先はusecase.SourceUCのみ
type SourceCtl struct {
	uc usecase.SourceUC
}

func NewSourceCtl(uc usecase.SourceUC) *SourceCtl {
	return &SourceCtl{
		uc: uc,
	}
}

// Source作成用のリクエストボディ
type CreateSourceReq struct {
	Name    string `json:"name"`
	SiteURL string `json:"site_url"`
}

// source単体レスポンス。
type SourceRes struct {
	Source entity.Source `json:"source"`
}

// source一覧レスポンス。
type SourceListRes struct {
	Sources []entity.Source `json:"sources"`
}

// GET /sources/:idを処理(path param idを読んで、usecaseを呼びsourceを返す)
func (ctl *SourceCtl) Get(c echo.Context) error {
	id, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}
	src, err := ctl.uc.Get(id)
	if err != nil {
		return writeErr(c, err)
	}
	return c.JSON(http.StatusOK, SourceRes{
		Source: src,
	})
}

// GET /sourcesを処理(queryのlimitとoffset)。
func (ctl *SourceCtl) List(c echo.Context) error {
	limit, err := qInt(c, "limit", 20)
	if err != nil {
		return writeErr(c, err)
	}
	offset, err := qInt(c, "offset", 0)
	if err != nil {
		return writeErr(c, err)
	}

	sources, err := ctl.uc.List(limit, offset)
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, SourceListRes{
		Sources: sources,
	})
}

// POST /sourcesを処理(管理者のみ。actor取得->requestbodyをbind->usecaseをDTO->usecaseを呼ぶ->errまたは作成結果を返す)
func (ctl *SourceCtl) Create(c echo.Context) error {
	// 管理系endpoint = 認証済みactorを要求。
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}

	var req CreateSourceReq
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	src, err := ctl.uc.Create(*actor, usecase.CreateSourceIn{
		Name:    req.Name,
		SiteURL: req.SiteURL,
	})
	if err != nil {
		return writeErr(c, err)
	}
	return c.JSON(http.StatusCreated, SourceRes{
		Source: src,
	})

}
