package usecase_test

import (
	"testing"

	"coffee-spa/entity"
	"coffee-spa/repository"
	"coffee-spa/testutil/repomock"
	"coffee-spa/testutil/servicemock"
	"coffee-spa/testutil/valmock"
	"coffee-spa/usecase"
)

func TestSessionUsecase_ListHistory_UsesActorUserID(t *testing.T) {
	sessions := &repomock.Session{
		ListHistoryFn: func(q repository.HistoryQ) ([]entity.Session, error) {
			if q.UserID != 3 {
				t.Fatalf("UserID = %d, want 3", q.UserID)
			}
			if q.Limit != 20 || q.Offset != 5 {
				t.Fatalf("paging = %d/%d", q.Limit, q.Offset)
			}
			uid := q.UserID
			return []entity.Session{{ID: 1, UserID: &uid}}, nil
		},
	}
	uc := usecase.NewSessionUsecase(sessions, &repomock.Audit{}, valmock.Search{}, servicemock.Clock{})

	got, err := uc.ListHistory(usecase.ListHistoryIn{Actor: entity.Actor{UserID: 3}, Limit: 20, Offset: 5})
	if err != nil {
		t.Fatalf("ListHistory() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("ListHistory() = %+v", got)
	}
}
