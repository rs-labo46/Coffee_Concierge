package db_test

import (
	"testing"

	"coffee-spa/db"
	"coffee-spa/entity"
	"coffee-spa/testutil/dbtest"
)

func TestMigrationAndSeed_RequiredCoverage(t *testing.T) {
	g := dbtest.OpenPostgres(t)
	if err := db.Migrate(db.DB{G: g}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := g.Exec("TRUNCATE saved_suggestions, suggestions, prefs, turns, sessions, recipes, beans, items, sources, audit_logs, refresh_tokens, pw_resets, email_verifies, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}

	if err := db.SeedDev(g, "admin@example.com", "password123"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var admin entity.User
	if err := g.Where("email = ?", "admin@example.com").First(&admin).Error; err != nil {
		t.Fatalf("admin missing: %v", err)
	}
	if admin.Role != entity.RoleAdmin || !admin.EmailVerified {
		t.Fatalf("admin=%#v", admin)
	}

	var sourceCount int64
	if err := g.Model(&entity.Source{}).Count(&sourceCount).Error; err != nil {
		t.Fatalf("count sources: %v", err)
	}
	if sourceCount == 0 {
		t.Fatal("expected seeded sources")
	}

	var beanCount int64
	if err := g.Model(&entity.Bean{}).Count(&beanCount).Error; err != nil {
		t.Fatalf("count beans: %v", err)
	}
	if beanCount == 0 {
		t.Fatal("expected seeded beans")
	}
}
