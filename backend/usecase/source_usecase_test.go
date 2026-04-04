package usecase

import (
	"testing"

	"coffee-spa/entity"
)

type mockSourceRepo2 struct {
	createFn func(s entity.Source) (entity.Source, error)
	listFn   func() ([]entity.Source, error)
}

func (m *mockSourceRepo2) Create(s entity.Source) (entity.Source, error) {
	return m.createFn(s)
}

func (m *mockSourceRepo2) List() ([]entity.Source, error) {
	return m.listFn()
}

type mockSourceVal2 struct {
	newSourceFn func(input AddSourceIn) error
}

func (m *mockSourceVal2) NewSource(input AddSourceIn) error {
	return m.newSourceFn(input)
}

type mockAuditRepository2 struct {
	createFn func(a entity.AuditLog) error
}

func (m *mockAuditRepository2) Create(a entity.AuditLog) error {
	return m.createFn(a)
}

func TestSourceUCList_OK(t *testing.T) {
	t.Parallel()

	uc := &SourceUC{
		source: &mockSourceRepo2{
			listFn: func() ([]entity.Source, error) {
				return []entity.Source{
					{ID: 1, Name: "a"},
					{ID: 2, Name: "b"},
				}, nil
			},
			createFn: func(s entity.Source) (entity.Source, error) {
				return entity.Source{}, nil
			},
		},
		audit: &mockAuditRepository2{
			createFn: func(a entity.AuditLog) error { return nil },
		},
		val: &mockSourceVal2{
			newSourceFn: func(input AddSourceIn) error { return nil },
		},
	}

	got, err := uc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}
