package usecase

import (
	"coffee-spa/entity"
	"coffee-spa/usecase/port"
	"encoding/json"
)

// Recipeの作成・更新・取得・一覧を扱う。
type RecipeUC interface {
	Create(actor entity.Actor, in CreateRecipeIn) (entity.Recipe, error)
	Update(actor entity.Actor, in UpdateRecipeIn) (entity.Recipe, error)
	Get(id uint) (entity.Recipe, error)
	List(in RecipeListIn) ([]entity.Recipe, error)
}

type RecipeListIn struct {
	BeanID   *uint
	Method   entity.Method
	TempPref entity.TempPref
	Active   *bool
	Limit    int
	Offset   int
}

type CreateRecipeIn struct {
	BeanID   uint
	Name     string
	Method   entity.Method
	TempPref entity.TempPref
	Grind    string
	Ratio    string
	Temp     int
	TimeSec  int
	Steps    []string
	Desc     string
	Active   bool
}

type UpdateRecipeIn struct {
	ID       uint
	BeanID   uint
	Name     string
	Method   entity.Method
	TempPref entity.TempPref
	Grind    string
	Ratio    string
	Temp     int
	TimeSec  int
	Steps    []string
	Desc     string
	Active   bool
}

type recipeUsecase struct {
	recipes port.RecipeRepository
	beans   port.BeanRepository
	audits  port.AuditRepository
	val     RecipeVal
}

func NewRecipeUsecase(
	recipes port.RecipeRepository,
	beans port.BeanRepository,
	audits port.AuditRepository,
	val RecipeVal,
) RecipeUC {
	return &recipeUsecase{
		recipes: recipes,
		beans:   beans,
		audits:  audits,
		val:     val,
	}
}

// Recipeを新規作成する。adminのみ許可する。bean_idの存在確認を先に行う。
func (u *recipeUsecase) Create(actor entity.Actor, in CreateRecipeIn) (entity.Recipe, error) {
	if actor.Role != entity.RoleAdmin {
		return entity.Recipe{}, ErrForbidden
	}

	if err := u.val.Create(in); err != nil {
		return entity.Recipe{}, err
	}

	if _, err := u.beans.GetByID(in.BeanID); err != nil {
		return entity.Recipe{}, err
	}

	recipe := &entity.Recipe{
		BeanID:   in.BeanID,
		Name:     in.Name,
		Method:   in.Method,
		TempPref: in.TempPref,
		Grind:    in.Grind,
		Ratio:    in.Ratio,
		Temp:     in.Temp,
		TimeSec:  in.TimeSec,
		Steps:    in.Steps,
		Desc:     in.Desc,
		Active:   in.Active,
	}

	if err := u.recipes.Create(recipe); err != nil {
		return entity.Recipe{}, err
	}

	u.writeAudit(
		"admin.recipes.create",
		&actor.UserID,
		map[string]string{
			"recipe_id": uintToStr(recipe.ID),
			"bean_id":   uintToStr(recipe.BeanID),
			"name":      recipe.Name,
			"method":    string(recipe.Method),
			"active":    boolToStr(recipe.Active),
		},
	)

	return *recipe, nil
}

// Recipeを更新する + adminのみ許可する + 更新前にbeanの存在確認を行う。
func (u *recipeUsecase) Update(actor entity.Actor, in UpdateRecipeIn) (entity.Recipe, error) {
	if actor.Role != entity.RoleAdmin {
		return entity.Recipe{}, ErrForbidden
	}

	if err := u.val.Update(in); err != nil {
		return entity.Recipe{}, err
	}

	if _, err := u.beans.GetByID(in.BeanID); err != nil {
		return entity.Recipe{}, err
	}

	recipe, err := u.recipes.GetByID(in.ID)
	if err != nil {
		return entity.Recipe{}, err
	}

	recipe.BeanID = in.BeanID
	recipe.Name = in.Name
	recipe.Method = in.Method
	recipe.TempPref = in.TempPref
	recipe.Grind = in.Grind
	recipe.Ratio = in.Ratio
	recipe.Temp = in.Temp
	recipe.TimeSec = in.TimeSec
	recipe.Steps = in.Steps
	recipe.Desc = in.Desc
	recipe.Active = in.Active

	if err := u.recipes.Update(recipe); err != nil {
		return entity.Recipe{}, err
	}

	u.writeAudit(
		"admin.recipes.update",
		&actor.UserID,
		map[string]string{
			"recipe_id": uintToStr(recipe.ID),
			"bean_id":   uintToStr(recipe.BeanID),
			"name":      recipe.Name,
			"method":    string(recipe.Method),
			"active":    boolToStr(recipe.Active),
		},
	)

	return *recipe, nil
}

// Recipeを1件取得する。
func (u *recipeUsecase) Get(id uint) (entity.Recipe, error) {
	if err := u.val.Get(id); err != nil {
		return entity.Recipe{}, err
	}

	recipe, err := u.recipes.GetByID(id)
	if err != nil {
		return entity.Recipe{}, err
	}

	return *recipe, nil
}

// Recipe一覧を返す。
func (u *recipeUsecase) List(in RecipeListIn) ([]entity.Recipe, error) {
	if err := u.val.List(in); err != nil {
		return nil, err
	}

	out, err := u.recipes.List(port.RecipeListQ{
		BeanID:   in.BeanID,
		Method:   in.Method,
		TempPref: in.TempPref,
		Active:   in.Active,
		Limit:    in.Limit,
		Offset:   in.Offset,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (u *recipeUsecase) writeAudit(
	typ string,
	userID *uint,
	meta map[string]string,
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
