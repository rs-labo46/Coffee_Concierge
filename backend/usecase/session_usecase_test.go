package usecase

import (
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

func TestSessionUsecase_GetSession_UserOwned(t *testing.T) {
	uid := uint(1)
	uc := NewSessionUsecase(sessionRepoMock{
		getSessionByIDFn: func(id uint) (*entity.Session, error) {
			return &entity.Session{ID: id, UserID: &uid, Status: entity.SessionActive}, nil
		},
		listTurnsFn: func(sessionID uint) ([]entity.Turn, error) { return []entity.Turn{{SessionID: sessionID}}, nil },
		getPrefFn:   func(sessionID uint) (*entity.Pref, error) { return &entity.Pref{SessionID: sessionID}, nil },
		listSuggestionsFn: func(sessionID uint) ([]entity.Suggestion, error) {
			return []entity.Suggestion{{SessionID: sessionID}}, nil
		},
	}, &auditRepoMock{}, searchValMock{}, fixedClock{now: time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)})
	out, err := uc.GetSession(GetSessionIn{Actor: &entity.Actor{UserID: 1}, SessionID: 9})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if out.Session.ID != 9 || len(out.Turns) != 1 || out.Pref.SessionID != 9 {
		t.Fatalf("unexpected output:%+v", out)
	}
}

func TestSessionUsecase_ListHistory_RequiresUser(t *testing.T) {
	uc := NewSessionUsecase(sessionRepoMock{}, &auditRepoMock{}, searchValMock{}, fixedClock{})
	_, err := uc.ListHistory(ListHistoryIn{Actor: entity.Actor{}})
	if err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got:%v", err)
	}
}

func TestSessionUsecase_CloseSession_GuestConflict(t *testing.T) {
	uc := NewSessionUsecase(sessionRepoMock{getGuestSessionFn: func(id uint, hash string, now time.Time) (*entity.Session, error) {
		return &entity.Session{ID: id, Status: entity.SessionClosed}, nil
	}}, &auditRepoMock{}, searchValMock{}, fixedClock{now: time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)})
	err := uc.CloseSession(CloseSessionIn{SessionID: 5, SessionKey: "plain"})
	if err != ErrConflict {
		t.Fatalf("expected ErrConflict, got:%v", err)
	}
}

func TestSessionUsecase_ListHistory_User(t *testing.T) {
	uc := NewSessionUsecase(sessionRepoMock{listHistoryFn: func(q repository.HistoryQ) ([]entity.Session, error) {
		return []entity.Session{{ID: 1}}, nil
	}}, &auditRepoMock{}, searchValMock{}, fixedClock{})
	out, err := uc.ListHistory(ListHistoryIn{Actor: entity.Actor{UserID: 3}, Limit: 10, Offset: 2})
	if err != nil || len(out) != 1 || out[0].ID != 1 {
		t.Fatalf("unexpected output:%+v err=%v", out, err)
	}
}
