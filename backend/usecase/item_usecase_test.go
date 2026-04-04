package usecase

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

type mockItemRepo struct {
	topFn     func(cap int) (repository.TopItems, error)
	createFn  func(i entity.Item) (entity.Item, error)
	getByIDFn func(id int64) (entity.Item, error)
	listFn    func(q repository.ItemQ) ([]entity.Item, error)
}

func (m *mockItemRepo) Create(i entity.Item) (entity.Item, error) {
	if m.createFn != nil {
		return m.createFn(i)
	}
	return entity.Item{}, nil
}

func (m *mockItemRepo) GetByID(id int64) (entity.Item, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return entity.Item{}, nil
}

func (m *mockItemRepo) List(q repository.ItemQ) ([]entity.Item, error) {
	if m.listFn != nil {
		return m.listFn(q)
	}
	return nil, nil
}

func (m *mockItemRepo) Top(cap int) (repository.TopItems, error) {
	if m.topFn != nil {
		return m.topFn(cap)
	}
	return repository.TopItems{}, nil
}

type mockAuditRepo struct {
	createFn func(a entity.AuditLog) error
}

func (m *mockAuditRepo) Create(a entity.AuditLog) error {
	if m.createFn != nil {
		return m.createFn(a)
	}
	return nil
}

type mockItemVal struct {
	newItemFn  func(input AddItemIn) error
	listItemFn func(q ItemQ) error
}

func (m *mockItemVal) NewItem(input AddItemIn) error {
	if m.newItemFn != nil {
		return m.newItemFn(input)
	}
	return nil
}

func (m *mockItemVal) ListItem(q ItemQ) error {
	if m.listItemFn != nil {
		return m.listItemFn(q)
	}
	return nil
}

func TestItemUCGet_OK(t *testing.T) {
	t.Parallel()

	uc := &ItemUC{
		item: &mockItemRepo{
			getByIDFn: func(id int64) (entity.Item, error) {
				if id != 5 {
					t.Fatalf("id = %d, want 5", id)
				}
				return entity.Item{ID: 5, Title: "detail", Kind: "news"}, nil
			},
		},
		audit: &mockAuditRepo{},
		val:   &mockItemVal{},
	}

	got, err := uc.Get(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 5 {
		t.Fatalf("id = %d, want 5", got.ID)
	}
}

func TestItemUCTop_ZeroCap_ReturnsEmptyGroups(t *testing.T) {
	t.Parallel()

	uc := &ItemUC{
		item: &mockItemRepo{
			topFn: func(cap int) (repository.TopItems, error) {
				if cap != 0 {
					t.Fatalf("cap = %d, want 0", cap)
				}
				return repository.TopItems{
					News:   []entity.Item{},
					Recipe: []entity.Item{},
					Deal:   []entity.Item{},
					Shop:   []entity.Item{},
				}, nil
			},
		},
		audit: &mockAuditRepo{},
		val:   &mockItemVal{},
	}

	got, err := uc.Top(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.News == nil || got.Recipe == nil || got.Deal == nil || got.Shop == nil {
		t.Fatalf("expected all groups to be non-nil slices")
	}
}

func TestItemUCTop_RepoError_Mapped(t *testing.T) {
	t.Parallel()

	uc := &ItemUC{
		item: &mockItemRepo{
			topFn: func(cap int) (repository.TopItems, error) {
				return repository.TopItems{}, repository.ErrInternal
			},
		},
		audit: &mockAuditRepo{},
		val:   &mockItemVal{},
	}

	_, err := uc.Top(3)
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("err = %v, want ErrInternal", err)
	}
}

func TestItemUCAdd_OK(t *testing.T) {
	t.Parallel()

	var created bool
	var audited bool

	uc := &ItemUC{
		item: &mockItemRepo{
			createFn: func(i entity.Item) (entity.Item, error) {
				created = true
				i.ID = 10
				return i, nil
			},
		},
		audit: &mockAuditRepo{
			createFn: func(a entity.AuditLog) error {
				audited = true
				return nil
			},
		},
		val: &mockItemVal{},
	}

	body := "full body"
	item, err := uc.Add(Actor{UserID: 1, IP: "127.0.0.1", UA: "test"}, AddItemIn{
		Title:       "coffee",
		Body:        &body,
		Kind:        "news",
		SourceID:    1,
		PublishedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Fatalf("item create was not called")
	}
	if !audited {
		t.Fatalf("audit create was not called")
	}
	if item.ID != 10 {
		t.Fatalf("item.ID = %d, want 10", item.ID)
	}
}
