package controller_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"coffee-spa/controller"
	"coffee-spa/entity"
	"coffee-spa/testutil/controllertest"
	"coffee-spa/testutil/usecasemock"

	"coffee-spa/usecase"
)

func setActorForController(c interface{ Set(string, interface{}) }) {
	c.Set(controller.ContextActorKey, &entity.Actor{UserID: 1, Role: entity.RoleAdmin, TokenVer: 1})
}

func TestContentAdminControllers_RequiredCoverage(t *testing.T) {
	t.Run("source create returns 201 and passes actor/body", func(t *testing.T) {
		_, c, rec := controllertest.JSONContext(http.MethodPost, "/admin/sources", `{"name":"src","site_url":"https://example.com"}`)
		setActorForController(c)
		ctl := controller.NewSourceCtl(&usecasemock.Source{CreateFn: func(a entity.Actor, in usecase.CreateSourceIn) (entity.Source, error) {
			if a.UserID != 1 || in.Name != "src" || in.SiteURL != "https://example.com" {
				t.Fatalf("unexpected input: %#v %#v", a, in)
			}
			return entity.Source{ID: 1, Name: in.Name, SiteURL: in.SiteURL}, nil
		}})
		if err := ctl.Create(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("source get rejects invalid path id", func(t *testing.T) {
		_, c, rec := controllertest.EmptyContext(http.MethodGet, "/sources/0")
		controllertest.WithParam(c, "id", "0")
		ctl := controller.NewSourceCtl(&usecasemock.Source{})
		if err := ctl.Get(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("item create maps source not found", func(t *testing.T) {
		body := `{"title":"item","summary":"summary","url":"https://example.com","kind":"news","source_id":1,"published_at":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
		_, c, rec := controllertest.JSONContext(http.MethodPost, "/admin/items", body)
		setActorForController(c)
		ctl := controller.NewItemCtl(&usecasemock.Item{CreateFn: func(a entity.Actor, in usecase.CreateItemIn) (entity.Item, error) {
			return entity.Item{}, usecase.ErrNotFound
		}})
		if err := ctl.Create(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("bean update rejects bad path id", func(t *testing.T) {
		_, c, rec := controllertest.JSONContext(http.MethodPatch, "/admin/beans/abc", `{}`)
		setActorForController(c)
		controllertest.WithParam(c, "id", "abc")
		ctl := controller.NewBeanCtl(&usecasemock.Bean{})
		if err := ctl.Update(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("recipe create returns 400 on bind error", func(t *testing.T) {
		_, c, rec := controllertest.JSONContext(http.MethodPost, "/admin/recipes", `{bad json`)
		setActorForController(c)
		ctl := controller.NewRecipeCtl(&usecasemock.Recipe{})
		if err := ctl.Create(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("audit list maps usecase forbidden", func(t *testing.T) {
		_, c, rec := controllertest.EmptyContext(http.MethodGet, "/admin/audit-logs?limit=10")
		setActorForController(c)
		ctl := controller.NewAuditCtl(&usecasemock.Audit{ListFn: func(a entity.Actor, in usecase.AuditListIn) ([]entity.AuditLog, error) {
			return nil, usecase.ErrForbidden
		}})
		if err := ctl.List(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("audit list rejects bad user_id query", func(t *testing.T) {
		_, c, rec := controllertest.EmptyContext(http.MethodGet, "/admin/audit-logs?user_id=abc")
		setActorForController(c)
		ctl := controller.NewAuditCtl(&usecasemock.Audit{})
		if err := ctl.List(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("list query parse error returns 400", func(t *testing.T) {
		_, c, rec := controllertest.EmptyContext(http.MethodGet, "/items?limit=bad")
		ctl := controller.NewItemCtl(&usecasemock.Item{})
		if err := ctl.List(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("unknown error maps to 500", func(t *testing.T) {
		_, c, rec := controllertest.EmptyContext(http.MethodGet, "/beans/1")
		controllertest.WithParam(c, "id", "1")
		ctl := controller.NewBeanCtl(&usecasemock.Bean{GetFn: func(id uint) (entity.Bean, error) { return entity.Bean{}, errors.New("boom") }})
		if err := ctl.Get(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d", rec.Code)
		}
	})
}
