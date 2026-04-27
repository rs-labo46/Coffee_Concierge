package controller

import (
	"encoding/json"
	"net/http"
	"testing"

	"coffee-spa/testutil/controllertest"
	"coffee-spa/testutil/usecasemock"
)

func TestSearchCtl_StartSession_ReturnsSessionKeyForGuest(t *testing.T) {
	_, c, rec := controllertest.JSONContext(http.MethodPost, "/search/sessions", `{"title":"morning"}`)
	ctl := NewSearchCtl(&usecasemock.SearchFlow{}, &usecasemock.Session{})

	if err := ctl.StartSession(c); err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var body StartSessionRes
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.SessionKey != "guest-key" {
		t.Fatalf("session key = %q", body.SessionKey)
	}
}
