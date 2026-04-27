package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coffee-spa/entity"

	"github.com/labstack/echo/v4"
)

func TestRequireAdmin_AdditionalCoverage(t *testing.T) {
	// M-ADMIN-ADD-01
	t.Run("admin passes", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/admin/items", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("actor", &entity.Actor{UserID: 1, Role: entity.RoleAdmin})
		called := false
		err := RequireAdmin()(func(c echo.Context) error {
			called = true
			return c.NoContent(http.StatusNoContent)
		})(c)
		if err != nil {
			t.Fatalf("handler error = %v", err)
		}
		if !called || rec.Code != http.StatusNoContent {
			t.Fatalf("called/code = %v/%d, want true/%d", called, rec.Code, http.StatusNoContent)
		}
	})

	// M-ADMIN-ADD-02
	t.Run("user rejected", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/admin/items", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("actor", &entity.Actor{UserID: 1, Role: entity.RoleUser})
		err := RequireAdmin()(func(c echo.Context) error {
			t.Fatal("next should not be called")
			return nil
		})(c)
		if err != nil {
			t.Fatalf("handler error = %v", err)
		}
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}

func TestSessionKeyAndWsMiddleware_AdditionalCoverage(t *testing.T) {
	// M-SESSION-ADD-01
	t.Run("header session key passes", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/search/sessions/1/pref", nil)
		req.Header.Set(HeaderSessionKey, "secret")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err := SessionKeyHeaderRequired()(func(c echo.Context) error {
			if got := c.Get("session_key"); got != "secret" {
				t.Fatalf("session_key = %v, want secret", got)
			}
			return c.NoContent(http.StatusNoContent)
		})(c)
		if err != nil {
			t.Fatalf("handler error = %v", err)
		}
	})

	// M-SESSION-ADD-02
	t.Run("missing header session key rejected", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/search/sessions/1/pref", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err := SessionKeyHeaderRequired()(func(c echo.Context) error {
			t.Fatal("next should not be called")
			return nil
		})(c)
		if err != nil {
			t.Fatalf("handler error = %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	// M-SESSION-ADD-03 / M-WS-ADD-01
	t.Run("query session key and ws upgrade pass", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/ws/guest/search/sessions/1?session_key=secret", nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		h := WsUpgradeCheck()(SessionKeyQueryRequired()(func(c echo.Context) error {
			if got := c.Get("session_key"); got != "secret" {
				t.Fatalf("session_key = %v, want secret", got)
			}
			return c.NoContent(http.StatusNoContent)
		}))
		if err := h(c); err != nil {
			t.Fatalf("handler error = %v", err)
		}
		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})

	// M-WS-ADD-02
	t.Run("ws upgrade rejects missing connection upgrade", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/ws/search/sessions/1", nil)
		req.Header.Set("Upgrade", "websocket")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err := WsUpgradeCheck()(func(c echo.Context) error {
			t.Fatal("next should not be called")
			return nil
		})(c)
		if err != nil {
			t.Fatalf("handler error = %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
