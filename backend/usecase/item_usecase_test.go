package usecase

import (
	"testing"
	"time"

	"coffee-spa/entity"
)

func TestItemUsecase_Create(t *testing.T) {
	now := time.Date(2026, 4, 23, 10, 0, 0, 0, time.UTC)
	uc := NewItemUsecase(
		itemRepoMock{createFn: func(item *entity.Item) error { item.ID = 21; return nil }},
		sourceRepoMock{getByIDFn: func(id uint) (*entity.Source, error) { return &entity.Source{ID: id}, nil }},
		&auditRepoMock{},
		itemValMock{},
	)
	out, err := uc.Create(entity.Actor{UserID: 1, Role: entity.RoleAdmin}, CreateItemIn{Title: "T", Summary: "S", URL: "https://x", ImageURL: "https://i", Kind: entity.ItemKindNews, SourceID: 2, PublishedAt: now})
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if out.ID != 21 || out.SourceID != 2 {
		t.Fatalf("unexpected output:%+v", out)
	}
}

func TestItemUsecase_Top(t *testing.T) {
	uc := NewItemUsecase(itemRepoMock{topFn: func(limit int) (*entity.TopItems, error) {
		return &entity.TopItems{News: []entity.Item{{ID: 1}}}, nil
	}}, sourceRepoMock{}, &auditRepoMock{}, itemValMock{})
	out, err := uc.Top(3)
	if err != nil {
		t.Fatalf("unexpected error:%v", err)
	}
	if len(out.News) != 1 || out.News[0].ID != 1 {
		t.Fatalf("unexpected output:%+v", out)
	}
}
