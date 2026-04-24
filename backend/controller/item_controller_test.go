package controller

import (
	"net/http"
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

func TestItemCtlTopOK(t *testing.T) {
	uc := &itemUCMock{topFn: func(limit int) (entity.TopItems, error) {
		if limit != 5 { t.Fatalf("limit=%d", limit) }
		return entity.TopItems{News: []entity.Item{{ID: 1, Title: "n"}}}, nil
	}}
	ctl := NewItemCtl(uc)
	c, rec, err := newJSONContext(http.MethodGet, "/items/top?limit=5", nil)
	if err != nil { t.Fatal(err) }
	if err := ctl.Top(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestItemCtlCreateCreated(t *testing.T) {
	publishedAt := time.Now().UTC().Truncate(time.Second)
	uc := &itemUCMock{createFn: func(actor entity.Actor, in usecase.CreateItemIn) (entity.Item, error) {
		if actor.UserID != 9 || in.SourceID != 2 { t.Fatalf("unexpected input: %+v %+v", actor, in) }
		return entity.Item{ID: 1, Title: in.Title, SourceID: in.SourceID, PublishedAt: publishedAt}, nil
	}}
	ctl := NewItemCtl(uc)
	body := map[string]interface{}{"title": "Bean", "summary": "s", "url": "https://ex.com", "image_url": "https://img", "kind": "news", "source_id": 2, "published_at": publishedAt.Format(time.RFC3339)}
	c, rec, err := newJSONContext(http.MethodPost, "/items", body)
	if err != nil { t.Fatal(err) }
	setActor(c, &entity.Actor{UserID: 9, Role: entity.RoleAdmin})
	if err := ctl.Create(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}
