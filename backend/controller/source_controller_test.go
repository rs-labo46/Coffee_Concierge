package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type mockSourceUC struct {
	listFn func() ([]entity.Source, error)
	addFn  func(actor usecase.Actor, in usecase.AddSourceIn) (entity.Source, error)

	addCalled bool
}

func (m *mockSourceUC) List() ([]entity.Source, error) {
	return m.listFn()
}

func (m *mockSourceUC) Add(actor usecase.Actor, in usecase.AddSourceIn) (entity.Source, error) {
	m.addCalled = true
	return m.addFn(actor, in)
}

func TestSrcCtlList_OK(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/sources", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctl := NewSrcCtl(&mockSourceUC{
		listFn: func() ([]entity.Source, error) {
			return []entity.Source{{ID: 1, Name: "Coffee Media"}}, nil
		},
		addFn: func(actor usecase.Actor, in usecase.AddSourceIn) (entity.Source, error) {
			return entity.Source{}, nil
		},
	})

	if err := ctl.List(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Coffee Media") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSrcCtlCreate_OK(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(
		http.MethodPost,
		"/sources",
		strings.NewReader(`{"name":"Coffee Media","site_url":"https://example.com"}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", int64(7))
	c.Set("role", "admin")

	ctl := NewSrcCtl(&mockSourceUC{
		listFn: func() ([]entity.Source, error) { return nil, nil },
		addFn: func(actor usecase.Actor, in usecase.AddSourceIn) (entity.Source, error) {
			if actor.UserID != 7 || actor.Role != "admin" {
				t.Fatalf("unexpected actor: %+v", actor)
			}
			if in.Name != "Coffee Media" {
				t.Fatalf("name = %q, want Coffee Media", in.Name)
			}
			if in.SiteURL == nil || *in.SiteURL != "https://example.com" {
				t.Fatalf("site_url = %+v", in.SiteURL)
			}
			return entity.Source{ID: 11, Name: in.Name, SiteURL: in.SiteURL}, nil
		},
	})

	if err := ctl.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Coffee Media") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSrcCtlCreate_InvalidJSON_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(
		http.MethodPost,
		"/sources",
		strings.NewReader(`{"name":"Coffee Media",`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", int64(1))
	c.Set("role", "admin")

	mockUC := &mockSourceUC{
		listFn: func() ([]entity.Source, error) { return nil, nil },
		addFn: func(actor usecase.Actor, in usecase.AddSourceIn) (entity.Source, error) {
			return entity.Source{}, nil
		},
	}
	ctl := NewSrcCtl(mockUC)

	if err := ctl.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if mockUC.addCalled {
		t.Fatal("usecase.Add must not be called on invalid json")
	}
}
