package controller

import (
	"net/http"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

func TestSourceCtlCreateCreated(t *testing.T) {
	uc := &sourceUCMock{createFn: func(actor entity.Actor, in usecase.CreateSourceIn) (entity.Source, error) {
		if actor.UserID != 1 || in.Name != "Official" { t.Fatalf("unexpected input: %+v %+v", actor, in) }
		return entity.Source{ID: 10, Name: in.Name, SiteURL: in.SiteURL}, nil
	}}
	ctl := NewSourceCtl(uc)
	c, rec, err := newJSONContext(http.MethodPost, "/sources", map[string]string{"name": "Official", "site_url": "https://ex.com"})
	if err != nil { t.Fatal(err) }
	setActor(c, &entity.Actor{UserID: 1, Role: entity.RoleAdmin})
	if err := ctl.Create(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestSourceCtlListBadLimit(t *testing.T) {
	uc := &sourceUCMock{listFn: func(int, int) ([]entity.Source, error) { t.Fatal("should not be called"); return nil, nil }}
	ctl := NewSourceCtl(uc)
	c, rec, err := newJSONContext(http.MethodGet, "/sources?limit=x", nil)
	if err != nil { t.Fatal(err) }
	if err := ctl.List(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}
