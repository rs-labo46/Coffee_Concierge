package controller

import (
	"encoding/json"
	"net/http"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/testutil/controllertest"
	"coffee-spa/testutil/usecasemock"
)

func TestSavedCtl_Save_ReturnsCreated(t *testing.T) {
	_, c, rec := controllertest.JSONContext(http.MethodPost, "/saved-suggestions", `{"session_id":9,"suggestion_id":4}`)
	c.Set(ContextActorKey, &entity.Actor{UserID: 2, Role: entity.RoleUser})
	ctl := NewSavedCtl(&usecasemock.Saved{})

	if err := ctl.Save(c); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var body SavedSuggestionRes
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Saved.SessionID != 9 || body.Saved.SuggestionID != 4 {
		t.Fatalf("body = %+v", body)
	}
}
