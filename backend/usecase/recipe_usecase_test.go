package usecase

import (
	"testing"

	"coffee-spa/entity"
)

func TestRecipeUsecase_Create(t *testing.T) {
	uc := NewRecipeUsecase(
		recipeRepoMock{createFn: func(recipe *entity.Recipe) error { recipe.ID = 31; return nil }},
		beanRepoMock{getByIDFn: func(id uint) (*entity.Bean, error) { return &entity.Bean{ID: id}, nil }},
		&auditRepoMock{},
		recipeValMock{},
	)
	out, err := uc.Create(entity.Actor{UserID: 1, Role: entity.RoleAdmin}, CreateRecipeIn{BeanID: 2, Name: "R", Method: entity.MethodDrip, TempPref: entity.TempHot, Grind: "m", Ratio: "1:15", Temp: 90, TimeSec: 180, Steps: []string{"a"}, Desc: "d", Active: true})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if out.ID != 31 || out.BeanID != 2 {
		t.Fatalf("unexpected output:%+v", out)
	}
}

func TestRecipeUsecase_Update(t *testing.T) {
	uc := NewRecipeUsecase(
		recipeRepoMock{
			getByIDFn: func(id uint) (*entity.Recipe, error) { return &entity.Recipe{ID: id, Name: "old"}, nil },
			updateFn: func(recipe *entity.Recipe) error {
				if recipe.Name != "new" {
					t.Fatalf("unexpected recipe:%+v", recipe)
				}
				return nil
			},
		},
		beanRepoMock{getByIDFn: func(id uint) (*entity.Bean, error) { return &entity.Bean{ID: id}, nil }},
		&auditRepoMock{},
		recipeValMock{},
	)
	out, err := uc.Update(entity.Actor{UserID: 1, Role: entity.RoleAdmin}, UpdateRecipeIn{ID: 1, BeanID: 2, Name: "new", Method: entity.MethodMilk, TempPref: entity.TempHot, Grind: "g", Ratio: "1", Temp: 80, TimeSec: 60, Steps: []string{"b"}, Desc: "d", Active: false})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if out.Name != "new" {
		t.Fatalf("unexpected output:%+v", out)
	}
}
