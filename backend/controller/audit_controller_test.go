package controller

import (
	"net/http"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

func TestAuditCtlListOK(t *testing.T) {
	uc := &auditUCMock{listFn: func(actor entity.Actor, in usecase.AuditListIn) ([]entity.AuditLog, error) {
		if actor.Role != entity.RoleAdmin || in.Type != "x" || in.Limit != 5 || in.Offset != 2 { t.Fatalf("unexpected input: %+v %+v", actor, in) }
		return []entity.AuditLog{{ID: 1, Type: "x"}}, nil
	}}
	ctl := NewAuditCtl(uc)
	c, rec, err := newJSONContext(http.MethodGet, "/audit-logs?type=x&limit=5&offset=2", nil)
	if err != nil { t.Fatal(err) }
	setActor(c, &entity.Actor{UserID: 1, Role: entity.RoleAdmin})
	if err := ctl.List(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}
