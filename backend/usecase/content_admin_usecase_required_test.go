package usecase_test

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
	"coffee-spa/testutil/repomock"
	"coffee-spa/testutil/valmock"
	"coffee-spa/usecase"
)

func adminActor() entity.Actor { return entity.Actor{UserID: 1, Role: entity.RoleAdmin, TokenVer: 1} }
func userActor() entity.Actor  { return entity.Actor{UserID: 2, Role: entity.RoleUser, TokenVer: 1} }

func TestContentAdminUsecases_RequiredCoverage(t *testing.T) {
	sentinel := errors.New("sentinel")

	t.Run("source create requires admin", func(t *testing.T) {
		uc := usecase.NewSourceUsecase(&repomock.Source{}, &repomock.Audit{}, valmock.Source{})
		_, err := uc.Create(userActor(), usecase.CreateSourceIn{Name: "src", SiteURL: "https://example.com"})
		if !errors.Is(err, usecase.ErrForbidden) {
			t.Fatalf("err=%v, want forbidden", err)
		}
	})

	t.Run("source create propagates validator error", func(t *testing.T) {
		uc := usecase.NewSourceUsecase(&repomock.Source{}, &repomock.Audit{}, valmock.Source{Err: sentinel})
		_, err := uc.Create(adminActor(), usecase.CreateSourceIn{Name: "", SiteURL: "bad"})
		if !errors.Is(err, sentinel) {
			t.Fatalf("err=%v, want sentinel", err)
		}
	})

	t.Run("source create writes audit", func(t *testing.T) {
		called := false
		uc := usecase.NewSourceUsecase(&repomock.Source{}, &repomock.Audit{CreateFn: func(log *entity.AuditLog) error {
			called = true
			if log.Type != "admin.sources.create" {
				t.Fatalf("audit type=%s", log.Type)
			}
			return nil
		}}, valmock.Source{})
		_, err := uc.Create(adminActor(), usecase.CreateSourceIn{Name: "src", SiteURL: "https://example.com"})
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		if !called {
			t.Fatal("audit was not written")
		}
	})

	t.Run("item create checks source before insert", func(t *testing.T) {
		itemCreateCalled := false
		uc := usecase.NewItemUsecase(&repomock.Item{CreateFn: func(item *entity.Item) error { itemCreateCalled = true; return nil }}, &repomock.Source{GetByIDFn: func(id uint) (*entity.Source, error) { return nil, usecase.ErrNotFound }}, &repomock.Audit{}, valmock.Item{})
		_, err := uc.Create(adminActor(), usecase.CreateItemIn{Title: "item", Summary: "summary", URL: "https://example.com", Kind: entity.ItemKindNews, SourceID: 99, PublishedAt: time.Now()})
		if !errors.Is(err, usecase.ErrNotFound) {
			t.Fatalf("err=%v, want not_found", err)
		}
		if itemCreateCalled {
			t.Fatal("item was created before source existence was confirmed")
		}
	})

	t.Run("bean update loads existing bean before saving", func(t *testing.T) {
		updated := false
		uc := usecase.NewBeanUsecase(&repomock.Bean{GetByIDFn: func(id uint) (*entity.Bean, error) { return &entity.Bean{ID: id, Name: "old"}, nil }, UpdateFn: func(bean *entity.Bean) error {
			updated = true
			if bean.Name != "new" {
				t.Fatalf("name=%s", bean.Name)
			}
			return nil
		}}, &repomock.Audit{}, valmock.Bean{})
		_, err := uc.Update(adminActor(), usecase.UpdateBeanIn{ID: 1, Name: "new", Roast: entity.RoastMedium, Origin: "BR", Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3, Active: true})
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		if !updated {
			t.Fatal("bean update was not called")
		}
	})

	t.Run("recipe create checks bean before insert", func(t *testing.T) {
		recipeCreateCalled := false
		uc := usecase.NewRecipeUsecase(&repomock.Recipe{CreateFn: func(recipe *entity.Recipe) error { recipeCreateCalled = true; return nil }}, &repomock.Bean{GetByIDFn: func(id uint) (*entity.Bean, error) { return nil, usecase.ErrNotFound }}, &repomock.Audit{}, valmock.Recipe{})
		_, err := uc.Create(adminActor(), usecase.CreateRecipeIn{BeanID: 99, Name: "recipe", Method: entity.MethodDrip, TempPref: entity.TempHot, Grind: "medium", Ratio: "1:15", Temp: 90, TimeSec: 180, Steps: []string{"brew"}, Active: true})
		if !errors.Is(err, usecase.ErrNotFound) {
			t.Fatalf("err=%v, want not_found", err)
		}
		if recipeCreateCalled {
			t.Fatal("recipe was created before bean existence was confirmed")
		}
	})

	t.Run("audit list requires admin", func(t *testing.T) {
		uc := usecase.NewAuditUsecase(&repomock.Audit{}, valmock.Audit{})
		_, err := uc.List(userActor(), usecase.AuditListIn{Limit: 20})
		if !errors.Is(err, usecase.ErrForbidden) {
			t.Fatalf("err=%v, want forbidden", err)
		}
	})

	t.Run("audit list passes query to repository", func(t *testing.T) {
		uid := uint(7)
		uc := usecase.NewAuditUsecase(&repomock.Audit{ListFn: func(q repository.AuditListQ) ([]entity.AuditLog, error) {
			if q.Type != "ai.failed" || q.UserID == nil || *q.UserID != uid || q.Limit != 10 || q.Offset != 5 {
				t.Fatalf("unexpected query: %#v", q)
			}
			return []entity.AuditLog{{ID: 1, Type: q.Type}}, nil
		}}, valmock.Audit{})
		logs, err := uc.List(adminActor(), usecase.AuditListIn{Type: "ai.failed", UserID: &uid, Limit: 10, Offset: 5})
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		if len(logs) != 1 {
			t.Fatalf("len=%d", len(logs))
		}
	})
}
