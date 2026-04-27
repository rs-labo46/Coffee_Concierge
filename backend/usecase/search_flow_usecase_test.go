package usecase_test

import (
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/testutil/repomock"
	"coffee-spa/testutil/servicemock"
	"coffee-spa/testutil/valmock"
	"coffee-spa/usecase"
)

func TestSearchFlowUsecase_StartSession_GuestCreatesSessionKey(t *testing.T) {
	created := false
	sessions := &repomock.Session{
		CreateSessionFn: func(s *entity.Session) error {
			created = true
			if s.UserID != nil {
				t.Fatalf("guest session UserID = %v, want nil", s.UserID)
			}
			if s.SessionKeyHash == "" {
				t.Fatal("guest session key hash is empty")
			}
			if s.GuestExpiresAt == nil {
				t.Fatal("guest expiration is nil")
			}
			s.ID = 10
			return nil
		},
	}

	uc := usecase.NewSearchFlowUsecase(
		sessions,
		nil,
		nil,
		nil,
		&repomock.Audit{},
		valmock.Search{},
		servicemock.Ranker{},
		servicemock.Gemini{},
		servicemock.Clock{NowFn: func() time.Time { return time.Unix(1000, 0).UTC() }},
		servicemock.IDGen{NewFn: func() string { return "guest-secret" }},
		24*time.Hour,
	)

	out, err := uc.StartSession(usecase.StartSessionIn{Title: "guest search"})
	if err != nil {
		t.Fatalf("StartSession() error = %v", err)
	}
	if !created {
		t.Fatal("CreateSession was not called")
	}
	if out.Session.ID != 10 {
		t.Fatalf("session id = %d, want 10", out.Session.ID)
	}
	if out.SessionKey != "guest-secret" {
		t.Fatalf("session key = %q", out.SessionKey)
	}
}
