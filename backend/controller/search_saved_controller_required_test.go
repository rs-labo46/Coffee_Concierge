package controller_test

import (
	"net/http"
	"testing"

	"coffee-spa/controller"
	"coffee-spa/entity"
	"coffee-spa/testutil/controllertest"
	"coffee-spa/testutil/usecasemock"
	"coffee-spa/usecase"
)

func TestSearchAndSavedControllers_RequiredCoverage(t *testing.T) {
	t.Run("set pref guest requires session key", func(t *testing.T) {
		_, c, rec := controllertest.JSONContext(http.MethodPost, "/search/sessions/1/pref", `{"flavor":3}`)
		controllertest.WithParam(c, "id", "1")
		ctl := controller.NewSearchCtl(&usecasemock.SearchFlow{}, &usecasemock.Session{})
		if err := ctl.SetPref(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("set pref passes guest session key", func(t *testing.T) {
		body := `{"flavor":3,"acidity":3,"bitterness":3,"body":3,"aroma":3,"mood":"relax","method":"drip","scene":"relax","temp_pref":"hot"}`
		_, c, rec := controllertest.JSONContext(http.MethodPost, "/search/sessions/1/pref", body)
		controllertest.WithParam(c, "id", "1")
		c.Request().Header.Set(controller.HeaderSessionKey, "guest-key")
		ctl := controller.NewSearchCtl(&usecasemock.SearchFlow{SetPrefFn: func(in usecase.SetPrefIn) (usecase.SetPrefOut, error) {
			if in.SessionID != 1 || in.SessionKey != "guest-key" {
				t.Fatalf("in=%#v", in)
			}
			return usecase.SetPrefOut{Pref: entity.Pref{SessionID: in.SessionID}}, nil
		}}, &usecasemock.Session{})
		if err := ctl.SetPref(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("get session requires actor", func(t *testing.T) {
		_, c, rec := controllertest.EmptyContext(http.MethodGet, "/search/sessions/1")
		controllertest.WithParam(c, "id", "1")
		ctl := controller.NewSearchCtl(&usecasemock.SearchFlow{}, &usecasemock.Session{})
		if err := ctl.GetSession(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("saved list passes paging", func(t *testing.T) {
		_, c, rec := controllertest.EmptyContext(http.MethodGet, "/saved-suggestions?limit=5&offset=2")
		setActorForController(c)
		ctl := controller.NewSavedCtl(&usecasemock.Saved{ListFn: func(in usecase.ListSavedIn) ([]entity.SavedSuggestion, error) {
			if in.Limit != 5 || in.Offset != 2 || in.Actor.UserID != 1 {
				t.Fatalf("in=%#v", in)
			}
			return []entity.SavedSuggestion{{ID: 1}}, nil
		}})
		if err := ctl.List(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d", rec.Code)
		}
	})

	t.Run("saved delete rejects bad suggestion id", func(t *testing.T) {
		_, c, rec := controllertest.EmptyContext(http.MethodDelete, "/saved-suggestions/abc")
		setActorForController(c)
		controllertest.WithParam(c, "suggestionId", "abc")
		ctl := controller.NewSavedCtl(&usecasemock.Saved{})
		if err := ctl.Delete(c); err != nil {
			t.Fatalf("err=%v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", rec.Code)
		}
	})
}
