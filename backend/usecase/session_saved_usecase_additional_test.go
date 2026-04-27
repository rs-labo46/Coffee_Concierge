package usecase_test

import (
	"errors"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/testutil/repomock"
	"coffee-spa/testutil/servicemock"
	"coffee-spa/testutil/valmock"
	"coffee-spa/usecase"
)

func TestSessionUsecase_AdditionalCoverage(t *testing.T) {
	// U-SES-ADD-01
	t.Run("ListHistory rejects empty actor", func(t *testing.T) {
		uc := usecase.NewSessionUsecase(&repomock.Session{}, &repomock.Audit{}, valmock.Search{}, servicemock.Clock{})
		_, err := uc.ListHistory(usecase.ListHistoryIn{Actor: entity.Actor{}, Limit: 10})
		if !errors.Is(err, usecase.ErrUnauthorized) {
			t.Fatalf("ListHistory() error = %v, want unauthorized", err)
		}
	})

	// U-SES-ADD-02
	t.Run("GetSession rejects other owner", func(t *testing.T) {
		owner := uint(2)
		sessions := &repomock.Session{GetSessionByIDFn: func(id uint) (*entity.Session, error) {
			return &entity.Session{ID: id, UserID: &owner, Status: entity.SessionActive}, nil
		}}
		uc := usecase.NewSessionUsecase(sessions, &repomock.Audit{}, valmock.Search{}, servicemock.Clock{})
		_, err := uc.GetSession(usecase.GetSessionIn{Actor: &entity.Actor{UserID: 1}, SessionID: 10})
		if !errors.Is(err, usecase.ErrForbidden) {
			t.Fatalf("GetSession() error = %v, want forbidden", err)
		}
	})

	// U-SES-ADD-03
	t.Run("CloseSession rejects already closed", func(t *testing.T) {
		owner := uint(1)
		sessions := &repomock.Session{GetSessionByIDFn: func(id uint) (*entity.Session, error) {
			return &entity.Session{ID: id, UserID: &owner, Status: entity.SessionClosed}, nil
		}}
		uc := usecase.NewSessionUsecase(sessions, &repomock.Audit{}, valmock.Search{}, servicemock.Clock{})
		err := uc.CloseSession(usecase.CloseSessionIn{Actor: &entity.Actor{UserID: 1}, SessionID: 10})
		if !errors.Is(err, usecase.ErrConflict) {
			t.Fatalf("CloseSession() error = %v, want conflict", err)
		}
	})
}

func TestSavedUsecase_AdditionalCoverage(t *testing.T) {
	// U-SAV-ADD-01
	t.Run("Save rejects guest", func(t *testing.T) {
		uc := usecase.NewSavedUsecase(&repomock.Saved{}, &repomock.Session{}, &repomock.Audit{}, valmock.Saved{})
		_, err := uc.Save(usecase.SaveSuggestionIn{Actor: entity.Actor{}, SessionID: 1, SuggestionID: 1})
		if !errors.Is(err, usecase.ErrUnauthorized) {
			t.Fatalf("Save() error = %v, want unauthorized", err)
		}
	})

	// U-SAV-ADD-02
	t.Run("Save rejects suggestion from another session", func(t *testing.T) {
		owner := uint(1)
		sessions := &repomock.Session{
			GetSessionByIDFn:    func(id uint) (*entity.Session, error) { return &entity.Session{ID: id, UserID: &owner}, nil },
			GetSuggestionByIDFn: func(id uint) (*entity.Suggestion, error) { return &entity.Suggestion{ID: id, SessionID: 999}, nil },
		}
		uc := usecase.NewSavedUsecase(&repomock.Saved{}, sessions, &repomock.Audit{}, valmock.Saved{})
		_, err := uc.Save(usecase.SaveSuggestionIn{Actor: entity.Actor{UserID: 1}, SessionID: 1, SuggestionID: 2})
		if !errors.Is(err, usecase.ErrConflict) {
			t.Fatalf("Save() error = %v, want conflict", err)
		}
	})

	// U-SAV-ADD-03
	t.Run("Delete returns lookup error", func(t *testing.T) {
		want := errors.New("lookup failed")
		saveds := &repomock.Saved{GetByUserAndSuggestionIDFn: func(u uint, s uint) (*entity.SavedSuggestion, error) {
			return nil, want
		}}
		uc := usecase.NewSavedUsecase(saveds, &repomock.Session{}, &repomock.Audit{}, valmock.Saved{})
		err := uc.Delete(usecase.DeleteSavedIn{Actor: entity.Actor{UserID: 1}, SuggestionID: 2})
		if !errors.Is(err, want) {
			t.Fatalf("Delete() error = %v, want %v", err, want)
		}
	})
}
