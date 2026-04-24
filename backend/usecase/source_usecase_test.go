package usecase

import (
	"testing"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

func TestSourceUsecase_Create(t *testing.T) {
	t.Run("adminだけ作成できる", func(t *testing.T) {
		repo := sourceRepoMock{createFn: func(src *entity.Source) error {
			src.ID = 10
			return nil
		}}
		audit := &auditRepoMock{}
		uc := NewSourceUsecase(repo, audit, sourceValMock{})
		out, err := uc.Create(entity.Actor{UserID: 1, Role: entity.RoleAdmin}, CreateSourceIn{Name: "S", SiteURL: "https://x"})
		if err != nil {
			t.Fatalf("unexpected error:%v", err)
		}
		if out.ID != 10 || out.Name != "S" {
			t.Fatalf("unexpected output:%+v", out)
		}
		if len(audit.logs) != 1 {
			t.Fatalf("audit not written")
		}
	})

	t.Run("admin以外はforbidden", func(t *testing.T) {
		uc := NewSourceUsecase(sourceRepoMock{}, &auditRepoMock{}, sourceValMock{})
		_, err := uc.Create(entity.Actor{UserID: 2, Role: entity.RoleUser}, CreateSourceIn{})
		if err != ErrForbidden {
			t.Fatalf("expected ErrForbidden, got:%v", err)
		}
	})
}

func TestSourceUsecase_List(t *testing.T) {
	uc := NewSourceUsecase(sourceRepoMock{listFn: func(q repository.SourceListQ) ([]entity.Source, error) {
		if q.Limit != 3 || q.Offset != 1 {
			t.Fatalf("unexpected query:%+v", q)
		}
		return []entity.Source{{ID: 1, Name: "A"}}, nil
	}}, &auditRepoMock{}, sourceValMock{})
	out, err := uc.List(3, 1)
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if len(out) != 1 || out[0].ID != 1 {
		t.Fatalf("unexpected output:%+v", out)
	}
}
