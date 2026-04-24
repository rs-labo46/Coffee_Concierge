package controller

import (
	"net/http"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

func TestRecipeCtlListBadBool(t *testing.T) {
	uc := &recipeUCMock{listFn: func(usecase.RecipeListIn) ([]entity.Recipe, error) { t.Fatal("should not be called"); return nil, nil }}
	ctl := NewRecipeCtl(uc)
	c, rec, err := newJSONContext(http.MethodGet, "/recipes?active=oops", nil)
	if err != nil { t.Fatal(err) }
	if err := ctl.List(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestRecipeCtlUpdateOK(t *testing.T) {
	uc := &recipeUCMock{updateFn: func(actor entity.Actor, in usecase.UpdateRecipeIn) (entity.Recipe, error) {
		if in.ID != 5 { t.Fatalf("id=%d", in.ID) }
		return entity.Recipe{ID: in.ID, Name: in.Name}, nil
	}}
	ctl := NewRecipeCtl(uc)
	body := map[string]interface{}{"bean_id": 1, "name": "V60", "method": "drip", "temp_pref": "hot", "grind": "mid", "ratio": "1:15", "temp": 92, "time_sec": 180, "steps": []string{"a"}, "desc": "d", "active": true}
	c, rec, err := newJSONContext(http.MethodPatch, "/recipes/5", body)
	if err != nil { t.Fatal(err) }
	c.SetParamNames("id")
	c.SetParamValues("5")
	setActor(c, &entity.Actor{UserID: 1, Role: entity.RoleAdmin})
	if err := ctl.Update(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}
