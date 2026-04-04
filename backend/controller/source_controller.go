package controller

import (
	"net/http"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type SrcCtl struct {
	uc usecase.SourceUsecase
}

func NewSrcCtl(uc usecase.SourceUsecase) SrcCtl {
	return SrcCtl{
		uc: uc,
	}
}

// source 作成入力。
type AddSourceReq struct {
	Name    string  `json:"name"`
	SiteURL *string `json:"site_url"`
}

// source 単体レスポンス。
type SourceRes struct {
	Source entity.Source `json:"source"`
}

// source 一覧レスポンス。
type SourceListRes struct {
	Sources []entity.Source `json:"sources"`
}

// GET /sources を処理。
func (ctl SrcCtl) List(c echo.Context) error {
	xs, err := ctl.uc.List()
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, SourceListRes{
		Sources: xs,
	})
}

func (ctl SrcCtl) Create(c echo.Context) error {
	var req AddSourceReq

	if err := bindJSON(c, &req); err != nil {
		return writeErr(c, err)
	}

	src, err := ctl.uc.Add(actorFromCtx(c), usecase.AddSourceIn{
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
