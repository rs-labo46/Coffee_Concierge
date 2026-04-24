package usecase

import (
	"testing"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

func TestAuditUsecase_List(t *testing.T) {
	t.Run("adminのみ成功", func(t *testing.T) {
		repo := &auditRepoMock{listFn: func(q repository.AuditListQ) ([]entity.AuditLog, error) {
			if q.Type != "auth.login" || q.Limit != 10 || q.Offset != 5 {
				t.Fatalf("unexpected query:%+v", q)
			}
			return []entity.AuditLog{{Type: "auth.login"}}, nil
		}}
		uc := NewAuditUsecase(repo, auditValMock{})
		out, err := uc.List(entity.Actor{UserID: 1, Role: entity.RoleAdmin}, AuditListIn{Type: "auth.login", Limit: 10, Offset: 5})
		if err != nil {
			t.Fatalf("unexpected error:%v", err)
		}
		if len(out) != 1 || out[0].Type != "auth.login" {
			t.Fatalf("unexpected output:%+v", out)
		}
	})

	t.Run("admin以外はforbidden", func(t *testing.T) {
		uc := NewAuditUsecase(&auditRepoMock{}, auditValMock{})
		_, err := uc.List(entity.Actor{UserID: 2, Role: entity.RoleUser}, AuditListIn{})
		if err != ErrForbidden {
			t.Fatalf("expected ErrForbidden, got:%v", err)
		}
	})
}
