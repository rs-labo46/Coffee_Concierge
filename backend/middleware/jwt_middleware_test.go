package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func TestJWTAuth_AllowsValidBearerToken(t *testing.T) {
	secret := "secret"
	claims := jwt.MapClaims{
		"sub":  "9",
		"role": "user",
		"tv":   float64(2),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+signed)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	called := false

	h := JWTAuth(secret)(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusNoContent)
	})
	if err := h(c); err != nil {
		t.Fatalf("handler error = %v", err)
	}
	if !called || rec.Code != http.StatusNoContent {
		t.Fatalf("called/code = %v/%d", called, rec.Code)
	}
	if got := c.Get("user_id"); got != uint(9) {
		t.Fatalf("user_id = %v", got)
	}
}

func TestJWTAuth_RejectsMissingBearerToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	h := JWTAuth("secret")(func(c echo.Context) error {
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
