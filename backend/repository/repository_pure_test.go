package repository

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/apperr"
	"coffee-spa/entity"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsDup_DetectsPostgresUniqueViolation(t *testing.T) {
	err := &pgconn.PgError{Code: "23505"}
	if !isDup(err) {
		t.Fatalf("expected duplicate error")
	}
}

func TestIsFK_DetectsPostgresForeignKeyViolation(t *testing.T) {
	err := &pgconn.PgError{Code: "23503"}
	if !isFK(err) {
		t.Fatalf("expected foreign key error")
	}
}

func TestUniqueNonEmpty_TrimsAndDeduplicates(t *testing.T) {
	got := uniqueNonEmpty(" Kenya ", "", "drip", "Kenya", "drip")
	want := []string{"Kenya", "drip"}
	if len(got) != len(want) {
		t.Fatalf("length mismatch: want=%d got=%d %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("value mismatch at %d: want=%s got=%s", i, want[i], got[i])
		}
	}
}

func TestToInt64_RejectsUnexpectedType(t *testing.T) {
	_, err := toInt64(struct{}{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestToInt_ParsesSupportedValues(t *testing.T) {
	cases := []interface{}{int(3), int64(3), float64(3), "3"}
	for _, tc := range cases {
		got, err := toInt(tc)
		if err != nil {
			t.Fatalf("toInt(%#v) returned error: %v", tc, err)
		}
		if got != 3 {
			t.Fatalf("toInt(%#v)=%d", tc, got)
		}
	}
}

func TestRateLimitStoreAllow_InvalidInputs(t *testing.T) {
	store := NewRateLimitRepository(nil)
	_, _, err := store.Allow("key", 1, 1, 1, time.Now())
	if err == nil {
		t.Fatalf("expected nil redis client error")
	}
}

func TestRepositoryNilInputs_ReturnInvalidState(t *testing.T) {
	db := openTestDB(t)
	if !errors.Is(NewUserRepository(db).Create(nil), apperr.ErrInvalidState) {
		t.Fatalf("user nil should be invalid state")
	}
	if !errors.Is(NewBeanRepository(db).Create(nil), apperr.ErrInvalidState) {
		t.Fatalf("bean nil should be invalid state")
	}
	if !errors.Is(NewRecipeRepository(db).Create(nil), apperr.ErrInvalidState) {
		t.Fatalf("recipe nil should be invalid state")
	}
	if !errors.Is(NewSourceRepository(db).Create(nil), apperr.ErrInvalidState) {
		t.Fatalf("source nil should be invalid state")
	}
	if !errors.Is(NewItemRepository(db).Create(nil), apperr.ErrInvalidState) {
		t.Fatalf("item nil should be invalid state")
	}
	if !errors.Is(NewSessionRepository(db).CreateSession(nil), apperr.ErrInvalidState) {
		t.Fatalf("session nil should be invalid state")
	}
	if !errors.Is(NewSavedRepository(db).Create(nil), apperr.ErrInvalidState) {
		t.Fatalf("saved nil should be invalid state")
	}
	if !errors.Is(NewAuditRepository(db).Create(nil), apperr.ErrInvalidState) {
		t.Fatalf("audit nil should be invalid state")
	}
}

func TestRepositoryInvalidID_ReturnsExpectedErrors(t *testing.T) {
	db := openTestDB(t)
	if _, err := NewUserRepository(db).GetByID(0); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("user id 0 should be not_found: %v", err)
	}
	if _, err := NewBeanRepository(db).GetByID(0); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("bean id 0 should be not_found: %v", err)
	}
	if _, err := NewSourceRepository(db).GetByID(0); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("source id 0 should be not_found: %v", err)
	}
	if _, err := NewSessionRepository(db).GetSessionByID(0); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("session id 0 should be not_found: %v", err)
	}
	if _, err := NewSavedRepository(db).List(SavedListQ{}); !errors.Is(err, apperr.ErrUnauthorized) {
		t.Fatalf("saved list without user should be unauthorized: %v", err)
	}
}

func TestConstructors_ReturnInterfaces(t *testing.T) {
	db := openTestDB(t)
	constructors := []interface{}{
		NewUserRepository(db), NewEmailVerifyRepository(db), NewPwResetRepository(db), NewRtRepository(db),
		NewSourceRepository(db), NewItemRepository(db), NewBeanRepository(db), NewRecipeRepository(db),
		NewSessionRepository(db), NewSavedRepository(db), NewAuditRepository(db),
	}
	for i, v := range constructors {
		if v == nil {
			t.Fatalf("constructor %d returned nil", i)
		}
	}
}

func TestApplyBeanExcludes_DoesNotPanic(t *testing.T) {
	db := openTestDB(t)
	tx := db.Model(&entity.Bean{})
	got := applyBeanExcludes(tx, []string{"acidic", " bitter ", "dark_roast", "milk_recipe", "unknown"})
	if got == nil {
		t.Fatalf("expected non nil query")
	}
}
