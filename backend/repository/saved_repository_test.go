package repository

import (
	"errors"
	"testing"

	"coffee-spa/apperr"
)

func TestSavedRepository_GuardClauses(t *testing.T) {
	repo := NewSavedRepository(nil)
	if err := repo.Create(nil); !errors.Is(err, apperr.ErrInvalidState) {
		t.Fatalf("Create(nil) error = %v, want %v", err, apperr.ErrInvalidState)
	}
	if _, err := repo.List(SavedListQ{}); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("List(empty) error = %v, want %v", err, apperr.ErrUnauthorized)
	}
	if err := repo.DeleteByUserAndSuggestionID(0, 1); !errors.Is(err, apperr.ErrInvalidState) {
		t.Fatalf("Delete invalid error = %v, want %v", err, apperr.ErrInvalidState)
	}
	if _, err := repo.GetByUserAndSuggestionID(1, 0); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("Get invalid error = %v, want %v", err, apperr.ErrNotFound)
	}
}
