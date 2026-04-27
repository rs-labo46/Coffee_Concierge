package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/testutil/middlewaremock"

	"github.com/labstack/echo/v4"
)

func TestTokenVersion_AllowsMatchingTokenVersion(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", uint(1))
	c.Set("tv", 2)

	reader := middlewaremock.TokenVersionReader{GetByIDFn: func(id uint) (*entity.User, error) {
		return &entity.User{ID: id, TokenVer: 2}, nil
	}}
	called := false
	h := TokenVersion(reader)(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})

	if err := h(c); err != nil {
		t.Fatalf("handler error = %v", err)
	}
	if !called || rec.Code != http.StatusNoContent {
		t.Fatalf("called/code = %v/%d", called, rec.Code)
	}
}

func TestTokenVersion_RejectsMissingUser(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", uint(1))
	c.Set("tv", 2)

	reader := middlewaremock.TokenVersionReader{GetByIDFn: func(id uint) (*entity.User, error) {
		return nil, errors.New("not found")
	}}
	h := TokenVersion(reader)(func(c echo.Context) error {
		t.Fatal("next should not be called")
		return nil
	})

	if err := h(c); err != nil {
		t.Fatalf("handler error = %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
