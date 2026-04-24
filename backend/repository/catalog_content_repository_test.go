package repository

import (
	"testing"
	"time"

	"coffee-spa/entity"
)

func TestSourceAndItemRepository_CreateListTopSearchRelated(t *testing.T) {
	db := openTestDB(t)
	srcRepo := NewSourceRepository(db)
	itemRepo := NewItemRepository(db)

	src := &entity.Source{Name: "Coffee Media", SiteURL: "https://example.com"}
	if err := srcRepo.Create(src); err != nil {
		t.Fatalf("create source: %v", err)
	}
	if src.ID == 0 {
		t.Fatalf("source id was not set")
	}

	item := &entity.Item{
		Title: "Kenya morning drip", Summary: "bright Kenya coffee", URL: "https://example.com/item", ImageURL: "https://example.com/image.jpg",
		Kind: entity.ItemKindRecipe, SourceID: src.ID, PublishedAt: time.Now().Add(-time.Hour),
	}
	if err := itemRepo.Create(item); err != nil {
		t.Fatalf("create item: %v", err)
	}
	got, err := itemRepo.GetByID(item.ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	if got.Source.ID != src.ID {
		t.Fatalf("source was not preloaded: %#v", got.Source)
	}
	items, err := itemRepo.List(ItemListQ{Q: "Kenya", Kind: entity.ItemKindRecipe, Limit: 10})
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	top, err := itemRepo.Top(3)
	if err != nil {
		t.Fatalf("top items: %v", err)
	}
	if len(top.Recipe) != 1 {
		t.Fatalf("expected recipe top item, got %#v", top)
	}
	related, err := itemRepo.SearchRelated("Kenya", entity.RoastMedium, "Kenya", entity.MoodMorning, entity.MethodDrip, 3, time.Now())
	if err != nil {
		t.Fatalf("search related: %v", err)
	}
	if len(related) == 0 {
		t.Fatalf("expected related items")
	}
}

func TestBeanAndRecipeRepository_CreateSearchAndFindPrimary(t *testing.T) {
	db := openTestDB(t)
	beanRepo := NewBeanRepository(db)
	recipeRepo := NewRecipeRepository(db)

	bean := &entity.Bean{
		Name: "Kenya AA", Roast: entity.RoastLight, Origin: "Kenya",
		Flavor: 4, Acidity: 5, Bitterness: 2, Body: 3, Aroma: 4,
		Desc: "fruity", BuyURL: "https://example.com/bean", Active: true,
	}
	if err := beanRepo.Create(bean); err != nil {
		t.Fatalf("create bean: %v", err)
	}
	beans, err := beanRepo.SearchByPref(entity.Pref{Flavor: 4, Acidity: 5, Bitterness: 2, Body: 3, Aroma: 4}, 10)
	if err != nil {
		t.Fatalf("search beans: %v", err)
	}
	if len(beans) != 1 || beans[0].ID != bean.ID {
		t.Fatalf("unexpected bean search result: %#v", beans)
	}
	recipe := &entity.Recipe{
		BeanID: bean.ID, Name: "Hot Drip", Method: entity.MethodDrip, TempPref: entity.TempHot,
		Grind: "medium", Ratio: "1:15", Temp: 90, TimeSec: 180, Steps: []string{"bloom", "pour"}, Desc: "desc", Active: true,
	}
	if err := recipeRepo.Create(recipe); err != nil {
		t.Fatalf("create recipe: %v", err)
	}
	primary, err := recipeRepo.FindPrimaryByBean(bean.ID, entity.MethodDrip, entity.TempHot)
	if err != nil {
		t.Fatalf("find primary recipe: %v", err)
	}
	if primary.ID != recipe.ID {
		t.Fatalf("unexpected primary recipe: %#v", primary)
	}
}
