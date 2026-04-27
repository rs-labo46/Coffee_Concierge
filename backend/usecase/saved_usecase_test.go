package usecase_test

import (
	"testing"

	"coffee-spa/entity"
	"coffee-spa/testutil/repomock"
	"coffee-spa/testutil/valmock"
	"coffee-spa/usecase"
)

func TestSavedUsecase_Save_RejectsGuest(t *testing.T) {
	uc := usecase.NewSavedUsecase(&repomock.Saved{}, &repomock.Session{}, &repomock.Audit{}, valmock.Saved{})

	_, err := uc.Save(usecase.SaveSuggestionIn{SessionID: 1, SuggestionID: 1})
	if err == nil {
		t.Fatal("Save() error = nil, want unauthorized")
	}
	if err != usecase.ErrUnauthorized {
		t.Fatalf("Save() error = %v, want %v", err, usecase.ErrUnauthorized)
	}
}

func TestSavedUsecase_Save_CreatesSavedSuggestionForOwner(t *testing.T) {
	uid := uint(2)
	sessions := &repomock.Session{
		GetSessionByIDFn: func(id uint) (*entity.Session, error) {
			return &entity.Session{ID: id, UserID: &uid}, nil
		},
		GetSuggestionByIDFn: func(id uint) (*entity.Suggestion, error) {
			return &entity.Suggestion{ID: id, SessionID: 9, BeanID: 1}, nil
		},
	}
	uc := usecase.NewSavedUsecase(&repomock.Saved{}, sessions, &repomock.Audit{}, valmock.Saved{})

	got, err := uc.Save(usecase.SaveSuggestionIn{Actor: entity.Actor{UserID: uid}, SessionID: 9, SuggestionID: 4})
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if got.UserID != uid || got.SessionID != 9 || got.SuggestionID != 4 {
		t.Fatalf("Save() = %+v", got)
	}
}
