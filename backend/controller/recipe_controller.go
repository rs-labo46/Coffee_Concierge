package controller

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

// 依存先はusecase.RecipeUCのみ
type RecipeCtl struct {
	uc usecase.RecipeUC
}

// 生成
func NewRecipeCtl(uc usecase.RecipeUC) *RecipeCtl {
	return &RecipeCtl{
		uc: uc,
	}
}

// Recipe作成用のrequest body
type CreateRecipeReq struct {
	BeanID   uint            `json:"bean_id"`
	Name     string          `json:"name"`
	Method   entity.Method   `json:"method"`
	TempPref entity.TempPref `json:"temp_pref"`
	Grind    string          `json:"grind"`
	Ratio    string          `json:"ratio"`
	Temp     int             `json:"temp"`
	TimeSec  int             `json:"time_sec"`
	Steps    []string        `json:"steps"`
	Desc     string          `json:"desc"`
	Active   bool            `json:"active"`
}

// Recipe更新用のrequest body
type UpdateRecipeReq struct {
	BeanID   uint            `json:"bean_id"`
	Name     string          `json:"name"`
	Method   entity.Method   `json:"method"`
	TempPref entity.TempPref `json:"temp_pref"`
	Grind    string          `json:"grind"`
	Ratio    string          `json:"ratio"`
	Temp     int             `json:"temp"`
	TimeSec  int             `json:"time_sec"`
	Steps    []string        `json:"steps"`
	Desc     string          `json:"desc"`
	Active   bool            `json:"active"`
}

// Recipeの単体レスポンス
type RecipeRes struct {
	Recipe entity.Recipe `json:"recipe"`
}

// Recipe一覧レスポンス
type RecipeListRes struct {
	Recipes []entity.Recipe `json:"recipes"`
}

// POST /recipesを処理
func (ctl *RecipeCtl) Create(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}

	var req CreateRecipeReq
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	recipe, err := ctl.uc.Create(*actor, usecase.CreateRecipeIn{
		BeanID:   req.BeanID,
		Name:     req.Name,
		Method:   req.Method,
		TempPref: req.TempPref,
		Grind:    req.Grind,
		Ratio:    req.Ratio,
		Temp:     req.Temp,
		TimeSec:  req.TimeSec,
		Steps:    req.Steps,
		Desc:     req.Desc,
		Active:   req.Active,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusCreated, RecipeRes{
		Recipe: recipe,
	})
}

// PATCH /recipes/:idを処理
func (ctl *RecipeCtl) Update(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}

	id, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}

	var req UpdateRecipeReq
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	recipe, err := ctl.uc.Update(*actor, usecase.UpdateRecipeIn{
		ID:       id,
		BeanID:   req.BeanID,
		Name:     req.Name,
		Method:   req.Method,
		TempPref: req.TempPref,
		Grind:    req.Grind,
		Ratio:    req.Ratio,
		Temp:     req.Temp,
		TimeSec:  req.TimeSec,
		Steps:    req.Steps,
		Desc:     req.Desc,
		Active:   req.Active,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, RecipeRes{
		Recipe: recipe,
	})
}

// GET /recipes/:idを処理
func (ctl *RecipeCtl) Get(c echo.Context) error {
	id, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}

	recipe, err := ctl.uc.Get(id)
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, RecipeRes{
		Recipe: recipe,
	})
}

// GET /recipesを処理(query:bean_id,method,temp_pref,active,limit,offset)

func (ctl *RecipeCtl) List(c echo.Context) error {
	beanID, err := qUint(c, "bean_id")
	if err != nil {
		return writeErr(c, err)
	}

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

	recipes, err := ctl.uc.List(usecase.RecipeListIn{
		BeanID:   beanID,
		Method:   entity.Method(qStr(c, "method")),
		TempPref: entity.TempPref(qStr(c, "temp_pref")),
		Active:   active,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, RecipeListRes{
		Recipes: recipes,
	})
}
