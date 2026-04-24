package controller

import (
	"net/http"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

func TestBeanCtlGetBadID(t *testing.T) {
	uc := &beanUCMock{getFn: func(uint) (entity.Bean, error) { t.Fatal("should not be called"); return entity.Bean{}, nil }}
	ctl := NewBeanCtl(uc)
	c, rec, err := newJSONContext(http.MethodGet, "/beans/abc", nil)
	if err != nil { t.Fatal(err) }
	c.SetParamNames("id")
	c.SetParamValues("abc")
	if err := ctl.Get(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestBeanCtlCreateCreated(t *testing.T) {
	uc := &beanUCMock{createFn: func(actor entity.Actor, in usecase.CreateBeanIn) (entity.Bean, error) {
		return entity.Bean{ID: 3, Name: in.Name, Roast: in.Roast}, nil
	}}
	ctl := NewBeanCtl(uc)
	body := map[string]interface{}{"name": "Kenya", "roast": "light", "origin": "KE", "flavor": 4, "acidity": 4, "bitterness": 2, "body": 3, "aroma": 5, "desc": "d", "buy_url": "https://ex.com", "active": true}
	c, rec, err := newJSONContext(http.MethodPost, "/beans", body)
	if err != nil { t.Fatal(err) }
	setActor(c, &entity.Actor{UserID: 1, Role: entity.RoleAdmin})
	if err := ctl.Create(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}
