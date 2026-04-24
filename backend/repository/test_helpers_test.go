package repository

import (
	"os"
	"testing"
	"time"

	"coffee-spa/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is not set; repository integration tests are skipped")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&entity.User{},
		&entity.EmailVerify{},
		&entity.PwReset{},
		&entity.Rt{},
		&entity.Source{},
		&entity.Item{},
		&entity.Bean{},
		&entity.Recipe{},
		&entity.Session{},
		&entity.Turn{},
		&entity.Pref{},
		&entity.Suggestion{},
		&entity.SavedSuggestion{},
		&entity.AuditLog{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	cleanTestDB(t, db)
	t.Cleanup(func() { cleanTestDB(t, db) })
	return db
}

func cleanTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	tables := []string{
		"saved_suggestions",
		"suggestions",
		"prefs",
		"turns",
		"sessions",
		"recipes",
		"items",
		"sources",
		"beans",
		"audit_logs",
		"refresh_tokens",
		"pw_resets",
		"email_verifies",
		"users",
	}
	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			t.Fatalf("clean %s: %v", table, err)
		}
	}
}

func seedUser(t *testing.T, db *gorm.DB, email string, role entity.Role) entity.User {
	t.Helper()
	user := entity.User{Email: email, PassHash: "hash", Role: role, TokenVer: 1, EmailVerified: true}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user
}

func seedSource(t *testing.T, db *gorm.DB) entity.Source {
	t.Helper()
	src := entity.Source{Name: "source", SiteURL: "https://example.com"}
	if err := db.Create(&src).Error; err != nil {
		t.Fatalf("seed source: %v", err)
	}
	return src
}

func seedBean(t *testing.T, db *gorm.DB, name string) entity.Bean {
	t.Helper()
	bean := entity.Bean{
		Name: name, Roast: entity.RoastMedium, Origin: "Kenya",
		Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3,
		Desc: "desc", BuyURL: "https://example.com/bean", Active: true,
	}
	if err := db.Create(&bean).Error; err != nil {
		t.Fatalf("seed bean: %v", err)
	}
	return bean
}

func seedRecipe(t *testing.T, db *gorm.DB, beanID uint) entity.Recipe {
	t.Helper()
	recipe := entity.Recipe{
		BeanID: beanID, Name: "drip recipe", Method: entity.MethodDrip, TempPref: entity.TempHot,
		Grind: "medium", Ratio: "1:15", Temp: 90, TimeSec: 180, Steps: []string{"pour"}, Desc: "desc", Active: true,
	}
	if err := db.Create(&recipe).Error; err != nil {
		t.Fatalf("seed recipe: %v", err)
	}
	return recipe
}

func seedItem(t *testing.T, db *gorm.DB, sourceID uint, kind entity.ItemKind, title string) entity.Item {
	t.Helper()
	item := entity.Item{
		Title: title, Summary: "Kenya morning drip", URL: "https://example.com/item", ImageURL: "https://example.com/image.jpg",
		Kind: kind, SourceID: sourceID, PublishedAt: time.Now().Add(-time.Hour),
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed item: %v", err)
	}
	return item
}

func seedSession(t *testing.T, db *gorm.DB, userID *uint) entity.Session {
	t.Helper()
	expires := time.Now().Add(24 * time.Hour)
	session := entity.Session{
		UserID: userID, Title: "test session", Status: entity.SessionActive,
		SessionKeyHash: "guest-key-hash", GuestExpiresAt: &expires,
	}
	if userID != nil {
		session.SessionKeyHash = ""
		session.GuestExpiresAt = nil
	}
	if err := db.Create(&session).Error; err != nil {
		t.Fatalf("seed session: %v", err)
	}
	return session
}
