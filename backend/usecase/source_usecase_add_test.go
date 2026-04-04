package usecase

import (
	"errors"
	"testing"

	"coffee-spa/entity"
)

func TestSourceUCAdd_OK(t *testing.T) {
	t.Parallel()

	var gotAuditType string

	uc := &SourceUC{
		source: &mockSourceRepo2{
			createFn: func(s entity.Source) (entity.Source, error) {
				if s.Name != "Coffee Media" {
					t.Fatalf("name = %q, want Coffee Media", s.Name)
				}
				if s.SiteURL == nil || *s.SiteURL != "https://example.com" {
					t.Fatalf("site_url = %+v, want https://example.com", s.SiteURL)
				}
				s.ID = 5
				return s, nil
			},
			listFn: func() ([]entity.Source, error) { return nil, nil },
		},
		audit: &mockAuditRepository2{
			createFn: func(a entity.AuditLog) error {
				gotAuditType = a.Type
				return nil
			},
		},
		val: &mockSourceVal2{
			newSourceFn: func(input AddSourceIn) error { return nil },
		},
	}

	siteURL := " https://example.com "
	got, err := uc.Add(Actor{UserID: 9, IP: "127.0.0.1", UA: "test"}, AddSourceIn{
		Name:    " Coffee Media ",
		SiteURL: &siteURL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 5 {
		t.Fatalf("source id = %d, want 5", got.ID)
	}
	if gotAuditType != "admin.sources.create" {
		t.Fatalf("audit type = %q, want admin.sources.create", gotAuditType)
	}
}

func TestSourceUCAdd_InvalidRequest(t *testing.T) {
	t.Parallel()

	uc := &SourceUC{
		source: &mockSourceRepo2{
			createFn: func(s entity.Source) (entity.Source, error) { return entity.Source{}, nil },
			listFn:   func() ([]entity.Source, error) { return nil, nil },
		},
		audit: &mockAuditRepository2{createFn: func(a entity.AuditLog) error { return nil }},
		val: &mockSourceVal2{
			newSourceFn: func(input AddSourceIn) error { return errors.New("invalid") },
		},
	}

	_, err := uc.Add(Actor{}, AddSourceIn{Name: ""})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}
