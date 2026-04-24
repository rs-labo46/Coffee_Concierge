package usecase

import (
	"testing"
	"time"

	"coffee-spa/entity"
)

func TestSearchFlowUsecase_StartSession_Guest(t *testing.T) {
	now := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	var created *entity.Session
	uc := NewSearchFlowUsecase(
		sessionRepoMock{createSessionFn: func(session *entity.Session) error { created = session; session.ID = 1; return nil }},
		beanRepoMock{},
		recipeRepoMock{},
		itemRepoMock{},
		&auditRepoMock{},
		searchValMock{},
		NewCoffeeRanker(),
		nil,
		fixedClock{now: now},
		fixedIDGen{id: "guest-key"},
		24*time.Hour,
	)
	out, err := uc.StartSession(StartSessionIn{Title: "guest-session"})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if out.Session.ID != 1 || out.SessionKey != "guest-key" {
		t.Fatalf("unexpected output:%+v", out)
	}
	if created == nil || created.UserID != nil || created.GuestExpiresAt == nil {
		t.Fatalf("guest session was not created correctly:%+v", created)
	}
}

func TestSearchFlowUsecase_StartSession_User(t *testing.T) {
	now := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	var created *entity.Session
	uc := NewSearchFlowUsecase(
		sessionRepoMock{createSessionFn: func(session *entity.Session) error { created = session; session.ID = 2; return nil }},
		beanRepoMock{}, recipeRepoMock{}, itemRepoMock{}, &auditRepoMock{}, searchValMock{}, NewCoffeeRanker(), nil, fixedClock{now: now}, fixedIDGen{id: "ignored"}, 24*time.Hour,
	)
	out, err := uc.StartSession(StartSessionIn{Actor: &entity.Actor{UserID: 9, Role: entity.RoleUser}, Title: "user-session"})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if out.Session.ID != 2 || out.SessionKey != "" {
		t.Fatalf("unexpected output:%+v", out)
	}
	if created == nil || created.UserID == nil || *created.UserID != 9 {
		t.Fatalf("user session was not created correctly:%+v", created)
	}
}
