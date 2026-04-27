package repository_test

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/apperr"
	"coffee-spa/db"
	"coffee-spa/entity"
	"coffee-spa/repository"
	"coffee-spa/testutil/dbtest"

	"gorm.io/gorm"
)

func openMigratedDB(t *testing.T) *gorm.DB {
	t.Helper()
	g := dbtest.OpenPostgres(t)
	if err := db.Migrate(db.DB{G: g}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := g.Exec("TRUNCATE saved_suggestions, suggestions, prefs, turns, sessions, recipes, beans, items, sources, audit_logs, refresh_tokens, pw_resets, email_verifies, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return g
}

func TestRepositoryDBBehavior_RequiredCoverage(t *testing.T) {
	g := openMigratedDB(t)

	t.Run("user email unique and lookup", func(t *testing.T) {
		repo := repository.NewUserRepository(g)
		u := &entity.User{Email: "user@example.com", PassHash: "hash", Role: entity.RoleUser, TokenVer: 1, EmailVerified: true}
		if err := repo.Create(u); err != nil {
			t.Fatalf("create: %v", err)
		}
		if _, err := repo.GetByEmail("user@example.com"); err != nil {
			t.Fatalf("get: %v", err)
		}
		dup := &entity.User{Email: "user@example.com", PassHash: "hash", Role: entity.RoleUser, TokenVer: 1}
		if err := repo.Create(dup); !errors.Is(err, apperr.ErrConflict) {
			t.Fatalf("dup err=%v", err)
		}
	})

	t.Run("content repositories create and filter", func(t *testing.T) {
		srcRepo := repository.NewSourceRepository(g)
		itemRepo := repository.NewItemRepository(g)
		src := &entity.Source{Name: "Source", SiteURL: "https://example.com"}
		if err := srcRepo.Create(src); err != nil {
			t.Fatalf("source create: %v", err)
		}
		if _, err := srcRepo.GetByID(src.ID); err != nil {
			t.Fatalf("source get: %v", err)
		}
		item := &entity.Item{Title: "Coffee News", Summary: "Light roast", URL: "https://example.com/news", Kind: entity.ItemKindNews, SourceID: src.ID, PublishedAt: time.Now().Add(-time.Hour)}
		if err := itemRepo.Create(item); err != nil {
			t.Fatalf("item create: %v", err)
		}
		items, err := itemRepo.List(repository.ItemListQ{Q: "Coffee", Kind: entity.ItemKindNews, Limit: 10})
		if err != nil {
			t.Fatalf("item list: %v", err)
		}
		if len(items) == 0 {
			t.Fatal("expected item list result")
		}
	})

	t.Run("bean and recipe repositories enforce parent and filters", func(t *testing.T) {
		beanRepo := repository.NewBeanRepository(g)
		recipeRepo := repository.NewRecipeRepository(g)
		bean := &entity.Bean{Name: "Brazil", Roast: entity.RoastMedium, Origin: "BR", Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3, Active: true}
		if err := beanRepo.Create(bean); err != nil {
			t.Fatalf("bean create: %v", err)
		}
		recipe := &entity.Recipe{BeanID: bean.ID, Name: "Drip", Method: entity.MethodDrip, TempPref: entity.TempHot, Grind: "medium", Ratio: "1:15", Temp: 90, TimeSec: 180, Steps: []string{"brew"}, Active: true}
		if err := recipeRepo.Create(recipe); err != nil {
			t.Fatalf("recipe create: %v", err)
		}
		recipes, err := recipeRepo.List(repository.RecipeListQ{BeanID: &bean.ID, Method: entity.MethodDrip, TempPref: entity.TempHot, Limit: 10})
		if err != nil {
			t.Fatalf("recipe list: %v", err)
		}
		if len(recipes) == 0 {
			t.Fatal("expected recipe list result")
		}
	})
}
