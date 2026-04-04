package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCSRF_MissingCookie_ReturnsForbidden(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.Header.Set("X-CSRF-Token", "token-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := CSRF()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestCSRF_MissingHeader_ReturnsForbidden(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token-1"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := CSRF()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestCSRF_Mismatch_ReturnsForbidden(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token-1"})
	req.Header.Set("X-CSRF-Token", "token-2")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := CSRF()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestCSRF_OK_Passes(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token-1"})
	req.Header.Set("X-CSRF-Token", "token-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := CSRF()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
