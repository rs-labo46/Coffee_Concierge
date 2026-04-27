package usecase_test

import (
	"testing"

	"coffee-spa/entity"
	"coffee-spa/testutil/repomock"
	"coffee-spa/usecase"
)

func TestAuthUsecase_Me_ReturnsCurrentUser(t *testing.T) {
	users := &repomock.User{
		GetByIDFn: func(id uint) (*entity.User, error) {
			if id != 7 {
				t.Fatalf("user id = %d, want 7", id)
			}
			return &entity.User{ID: id, Email: "me@example.com", Role: entity.RoleUser, TokenVer: 3}, nil
		},
	}
	uc := usecase.NewAuthUsecase(users, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, 0)

	got, err := uc.Me(entity.Actor{UserID: 7, Role: entity.RoleUser, TokenVer: 3})
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}
	if got.ID != 7 || got.Email != "me@example.com" {
		t.Fatalf("Me() = %+v", got)
	}
}
