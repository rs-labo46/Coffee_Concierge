package repository

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/apperr"
)

func TestSessionRepository_GuardClauses(t *testing.T) {
	repo := NewSessionRepository(nil)
	if err := repo.CreateSession(nil); !errors.Is(err, apperr.ErrInvalidState) {
		t.Fatalf("CreateSession(nil) error = %v, want %v", err, apperr.ErrInvalidState)
	}
	if _, err := repo.GetSessionByID(0); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("GetSessionByID(0) error = %v, want %v", err, apperr.ErrNotFound)
	}
	if _, err := repo.GetGuestSessionByID(0, "key", time.Now()); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("GetGuestSessionByID(0) error = %v, want %v", err, apperr.ErrNotFound)
	}
	if _, err := repo.ListHistory(HistoryQ{}); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("ListHistory(empty) error = %v, want %v", err, apperr.ErrUnauthorized)
	}
}
