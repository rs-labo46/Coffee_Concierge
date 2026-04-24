package usecase

import (
	"database/sql"
	"testing"
)

func TestNewHealthUC(t *testing.T) {
	var db *sql.DB
	uc := NewHealthUC(db)
	if uc == nil {
		t.Fatal("expected usecase")
	}
}
