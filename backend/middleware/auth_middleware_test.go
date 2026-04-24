package middleware

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"coffee-spa/entity"

	"github.com/labstack/echo/v4"
)

func TestJWTAuth_ValidToken_SetsActorAndCallsNext(t *testing.T) {
	secret := "secret"
	c, rec := newTestContext(http.MethodGet, "/me", "")
	token := makeAccessToken(t, secret, 12, entity.RoleAdmin, 7, time.Now().Add(time.Hour))
	c.Request().Header.Set(echo.HeaderAuthorization, "Bearer "+token)

	err := JWTAuth(secret)(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)

	actor, ok := c.Get("actor").(*entity.Actor)
	if !ok || actor == nil {
		t.Fatalf("actor was not set")
	}
	if actor.UserID != 12 || actor.Role != entity.RoleAdmin || actor.TokenVer != 7 {
		t.Fatalf("actor mismatch: %#v", actor)
	}
}

func TestJWTAuth_MissingBearer_ReturnsUnauthorized(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/me", "")
	err := JWTAuth("secret")(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestJWTAuth_InvalidSignature_ReturnsUnauthorized(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/me", "")
	token := makeAccessToken(t, "other-secret", 1, entity.RoleUser, 1, time.Now().Add(time.Hour))
	c.Request().Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	err := JWTAuth("secret")(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestJWTAuth_ExpiredToken_ReturnsUnauthorized(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/me", "")
	token := makeAccessToken(t, "secret", 1, entity.RoleUser, 1, time.Now().Add(-time.Hour))
	c.Request().Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	err := JWTAuth("secret")(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestTokenVersion_MatchedTokenVersion_CallsNext(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/me", "")
	c.Set("user_id", uint(10))
	c.Set("tv", 3)
	repo := &tvReaderMock{user: &entity.User{ID: 10, TokenVer: 3}}
	err := TokenVersion(repo)(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
	if len(repo.ids) != 1 || repo.ids[0] != 10 {
		t.Fatalf("GetByID was not called correctly: %#v", repo.ids)
	}
}

func TestTokenVersion_MismatchedTokenVersion_ReturnsUnauthorized(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/me", "")
	c.Set("user_id", uint(10))
	c.Set("tv", 2)
	repo := &tvReaderMock{user: &entity.User{ID: 10, TokenVer: 3}}
	err := TokenVersion(repo)(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestTokenVersion_RepositoryError_ReturnsUnauthorized(t *testing.T) {
	c, rec := newTestContext(http.MethodGet, "/me", "")
	c.Set("user_id", uint(10))
	c.Set("tv", 3)
	repo := &tvReaderMock{err: errors.New("db down")}
	err := TokenVersion(repo)(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusUnauthorized)
}

func TestRequireAdmin_AdminActor_CallsNext(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/admin/beans", "")
	c.Set("actor", &entity.Actor{UserID: 1, Role: entity.RoleAdmin})
	err := RequireAdmin()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusOK)
}

func TestRequireAdmin_UserActor_ReturnsForbidden(t *testing.T) {
	c, rec := newTestContext(http.MethodPost, "/admin/beans", "")
	c.Set("actor", &entity.Actor{UserID: 1, Role: entity.RoleUser})
	err := RequireAdmin()(okHandler)(c)
	if err != nil {
		t.Fatalf("middleware returned error: %v", err)
	}
	assertStatus(t, rec, http.StatusForbidden)
}
