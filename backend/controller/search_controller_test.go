package controller

import (
	"net/http"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

func TestSearchCtlStartSessionGuestCreated(t *testing.T) {
	flow := &searchFlowUCMock{startSessionFn: func(in usecase.StartSessionIn) (usecase.StartSessionOut, error) {
		if in.Actor != nil || in.Title != "hello" { t.Fatalf("unexpected input: %+v", in) }
		return usecase.StartSessionOut{Session: entity.Session{ID: 3, Title: in.Title}, SessionKey: "guest-key"}, nil
	}}
	ctl := NewSearchCtl(flow, &sessionUCMock{})
	c, rec, err := newJSONContext(http.MethodPost, "/search/sessions", map[string]string{"title": "hello"})
	if err != nil { t.Fatal(err) }
	c.Set("request_id", "req-1")
	if err := ctl.StartSession(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestSearchCtlSetPrefGuestMissingSessionKey(t *testing.T) {
	flow := &searchFlowUCMock{setPrefFn: func(usecase.SetPrefIn) (usecase.SetPrefOut, error) { t.Fatal("should not be called"); return usecase.SetPrefOut{}, nil }}
	ctl := NewSearchCtl(flow, &sessionUCMock{})
	c, rec, err := newJSONContext(http.MethodPost, "/search/sessions/9/pref", map[string]int{"flavor": 3})
	if err != nil { t.Fatal(err) }
	c.SetParamNames("id")
	c.SetParamValues("9")
	if err := ctl.SetPref(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestSearchCtlGetGuestSessionOK(t *testing.T) {
	sessionUC := &sessionUCMock{getSessionFn: func(in usecase.GetSessionIn) (usecase.GetSessionOut, error) {
		if in.Actor != nil || in.SessionID != 4 || in.SessionKey != "k" { t.Fatalf("unexpected input: %+v", in) }
		return usecase.GetSessionOut{Session: entity.Session{ID: 4}}, nil
	}}
	ctl := NewSearchCtl(&searchFlowUCMock{}, sessionUC)
	c, rec, err := newJSONContext(http.MethodGet, "/search/guest/sessions/4", nil)
	if err != nil { t.Fatal(err) }
	c.SetParamNames("id")
	c.SetParamValues("4")
	c.Request().Header.Set(HeaderSessionKey, "k")
	if err := ctl.GetGuestSession(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}
