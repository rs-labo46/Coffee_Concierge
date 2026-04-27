package controller

import (
	"encoding/json"
	"net/http"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/testutil/controllertest"
	"coffee-spa/testutil/usecasemock"
)

func TestAuthCtl_Me_ReturnsCurrentUser(t *testing.T) {
	_, c, rec := controllertest.EmptyContext(http.MethodGet, "/me")
	c.Set(ContextActorKey, &entity.Actor{UserID: 5, Role: entity.RoleUser, TokenVer: 2})
	ctl := NewAuthCtl(&usecasemock.Auth{}, nil)

	if err := ctl.Me(c); err != nil {
		t.Fatalf("Me() error = %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var body MeRes
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.User.ID != 5 {
		t.Fatalf("user id = %d, want 5", body.User.ID)
	}
}
