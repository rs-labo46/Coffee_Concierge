package controller

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

type BeanCtl struct {
	uc usecase.BeanUC
}

// 生成
func NewBeanCtl(uc usecase.BeanUC) *BeanCtl {
	return &BeanCtl{
		uc: uc,
	}
}

// Bean作成用のrequest body
type CreateBeanReq struct {
	Name       string       `json:"name"`
	Roast      entity.Roast `json:"roast"`
	Origin     string       `json:"origin"`
	Flavor     int          `json:"flavor"`
	Acidity    int          `json:"acidity"`
	Bitterness int          `json:"bitterness"`
	Body       int          `json:"body"`
	Aroma      int          `json:"aroma"`
	Desc       string       `json:"desc"`
	BuyURL     string       `json:"buy_url"`
	Active     bool         `json:"active"`
}

// Bean更新用のrequest body
type UpdateBeanReq struct {
	Name       string       `json:"name"`
	Roast      entity.Roast `json:"roast"`
	Origin     string       `json:"origin"`
	Flavor     int          `json:"flavor"`
	Acidity    int          `json:"acidity"`
	Bitterness int          `json:"bitterness"`
	Body       int          `json:"body"`
	Aroma      int          `json:"aroma"`
	Desc       string       `json:"desc"`
	BuyURL     string       `json:"buy_url"`
	Active     bool         `json:"active"`
}

// beanの単体レスポンス
type BeanRes struct {
	Bean entity.Bean `json:"bean"`
}

// Bean一覧レスポンス
type BeanListRes struct {
	Beans []entity.Bean `json:"beans"`
}

// POST /beansを処理
// 管理者のみが作成
func (ctl *BeanCtl) Create(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}

	var req CreateBeanReq
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	bean, err := ctl.uc.Create(*actor, usecase.CreateBeanIn{
		Name:       req.Name,
		Roast:      req.Roast,
		Origin:     req.Origin,
		Flavor:     req.Flavor,
		Acidity:    req.Acidity,
		Bitterness: req.Bitterness,
		Body:       req.Body,
		Aroma:      req.Aroma,
		Desc:       req.Desc,
		BuyURL:     req.BuyURL,
		Active:     req.Active,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusCreated, BeanRes{
		Bean: bean,
	})
}

// PATCH /beans/:id を処理
// 管理者のみが更新できる
func (ctl *BeanCtl) Update(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}

	id, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}

	var req UpdateBeanReq
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	bean, err := ctl.uc.Update(*actor, usecase.UpdateBeanIn{
		ID:         id,
		Name:       req.Name,
		Roast:      req.Roast,
		Origin:     req.Origin,
		Flavor:     req.Flavor,
		Acidity:    req.Acidity,
		Bitterness: req.Bitterness,
		Body:       req.Body,
		Aroma:      req.Aroma,
		Desc:       req.Desc,
		BuyURL:     req.BuyURL,
		Active:     req.Active,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, BeanRes{
		Bean: bean,
	})
}

// GET /beans/:id を処理
func (ctl *BeanCtl) Get(c echo.Context) error {
	id, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}

	bean, err := ctl.uc.Get(id)
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, BeanRes{
		Bean: bean,
	})
}

// GET /beansを処理(quert:roaset,active,limit,offset)

func (ctl *BeanCtl) List(c echo.Context) error {
	active, err := qBool(c, "active")
	if err != nil {
		return writeErr(c, err)
	}

	limit, err := qInt(c, "limit", 20)
	if err != nil {
		return writeErr(c, err)
	}

	offset, err := qInt(c, "offset", 0)
	if err != nil {
		return writeErr(c, err)
	}

	beans, err := ctl.uc.List(usecase.BeanListIn{
		Q:      qStr(c, "q"),
		Roast:  entity.Roast(qStr(c, "roast")),
		Active: active,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, BeanListRes{
		Beans: beans,
	})
}
