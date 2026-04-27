package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestWsUpgradeCheck_AllowsUpgradeRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws/guest/search/sessions/1", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	called := false

	h := WsUpgradeCheck()(func(c echo.Context) error {
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

func TestWsUpgradeCheck_RejectsNormalHTTPRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ws/guest/search/sessions/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := WsUpgradeCheck()(func(c echo.Context) error {
		t.Fatal("next should not be called")
		return nil
	})
	if err := h(c); err != nil {
		t.Fatalf("handler error = %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
