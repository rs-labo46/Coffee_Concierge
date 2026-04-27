package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCsrfCheck_AllowsMatchingCookieAndHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "token"})
	req.Header.Set(csrfHeaderName, "token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	called := false

	h := CsrfCheck()(func(c echo.Context) error {
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

func TestCsrfCheck_RejectsMismatch(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "cookie"})
	req.Header.Set(csrfHeaderName, "header")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := CsrfCheck()(func(c echo.Context) error {
		t.Fatal("next should not be called")
		return nil
	})
	if err := h(c); err != nil {
		t.Fatalf("handler error = %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
