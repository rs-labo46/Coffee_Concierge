package repository

import (
	"errors"
	"testing"

	"coffee-spa/apperr"
)

func TestUserRepository_GuardClauses(t *testing.T) {
	repo := NewUserRepository(nil)
	if err := repo.Create(nil); !errors.Is(err, apperr.ErrInvalidState) {
		t.Fatalf("Create(nil) error = %v, want %v", err, apperr.ErrInvalidState)
	}
	if _, err := repo.GetByID(0); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("GetByID(0) error = %v, want %v", err, apperr.ErrNotFound)
	}
}
