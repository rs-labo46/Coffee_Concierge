package usecase

import (
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase/port"
)

func TestBeanUsecase_Create(t *testing.T) {
	uc := NewBeanUsecase(beanRepoMock{createFn: func(bean *entity.Bean) error {
		bean.ID = 11
		return nil
	}}, &auditRepoMock{}, beanValMock{})
	out, err := uc.Create(entity.Actor{UserID: 1, Role: entity.RoleAdmin}, CreateBeanIn{Name: "Kenya", Roast: entity.RoastLight, Origin: "Kenya", Flavor: 4, Acidity: 4, Bitterness: 2, Body: 3, Aroma: 4, Desc: "d", BuyURL: "u", Active: true})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if out.ID != 11 || out.Name != "Kenya" {
		t.Fatalf("unexpected output:%+v", out)
	}
}

func TestBeanUsecase_Update(t *testing.T) {
	uc := NewBeanUsecase(beanRepoMock{
		getByIDFn: func(id uint) (*entity.Bean, error) { return &entity.Bean{ID: id, Name: "old"}, nil },
		updateFn: func(bean *entity.Bean) error {
			if bean.Name != "new" {
				t.Fatalf("bean not updated:%+v", bean)
			}
			return nil
		},
	}, &auditRepoMock{}, beanValMock{})
	out, err := uc.Update(entity.Actor{UserID: 1, Role: entity.RoleAdmin}, UpdateBeanIn{ID: 1, Name: "new", Roast: entity.RoastDark, Origin: "B", Flavor: 1, Acidity: 1, Bitterness: 1, Body: 1, Aroma: 1, Desc: "d", BuyURL: "u", Active: false})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if out.Name != "new" {
		t.Fatalf("unexpected output:%+v", out)
	}
}

func TestBeanUsecase_List(t *testing.T) {
	uc := NewBeanUsecase(beanRepoMock{listFn: func(q port.BeanListQ) ([]entity.Bean, error) {
		return []entity.Bean{{ID: 1, Name: q.Q}}, nil
	}}, &auditRepoMock{}, beanValMock{})
	out, err := uc.List(BeanListIn{Q: "abc", Limit: 5})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if len(out) != 1 || out[0].Name != "abc" {
		t.Fatalf("unexpected output:%+v", out)
	}
}
