package controller

import (
	"net/http"
	"strings"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

func TestAuthCtlSignupCreated(t *testing.T) {
	uc := &authUCMock{signupFn: func(in usecase.SignupIn) (usecase.SignupOut, error) {
		if in.Email != "a@test.com" || in.Password != "pw" {
			t.Fatalf("unexpected signup input: %+v", in)
		}
		return usecase.SignupOut{User: entity.User{ID: 1, Email: in.Email, Role: entity.RoleUser}}, nil
	}}
	ctl := NewAuthCtl(uc, &rateLimiterMock{})
	c, rec, err := newJSONContext(http.MethodPost, "/auth/signup", map[string]string{"email": "a@test.com", "password": "pw"})
	if err != nil { t.Fatal(err) }

	if err := ctl.Signup(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if !strings.Contains(rec.Body.String(), `"email":"a@test.com"`) { t.Fatalf("body=%s", rec.Body.String()) }
}

func TestAuthCtlLoginSetsRefreshCookie(t *testing.T) {
	rl := &rateLimiterMock{
		allowLoginIPFn: func(string) (bool, int, error) { return true, 0, nil },
		allowLoginFn:   func(string) (bool, int, error) { return true, 0, nil },
	}
	uc := &authUCMock{loginFn: func(in usecase.LoginIn) (usecase.LoginOut, error) {
		return usecase.LoginOut{User: entity.User{ID: 2, Email: in.Email, Role: entity.RoleUser}, AccessToken: "acc", RefreshToken: "ref"}, nil
	}}
	ctl := NewAuthCtl(uc, rl)
	c, rec, err := newJSONContext(http.MethodPost, "/auth/login", map[string]string{"email": "a@test.com", "password": "pw"})
	if err != nil { t.Fatal(err) }

	if err := ctl.Login(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if got := rec.Header().Get("Set-Cookie"); !strings.Contains(got, refreshCookieName+"=ref") { t.Fatalf("set-cookie=%s", got) }
}

func TestAuthCtlRefreshRateLimited(t *testing.T) {
	rl := &rateLimiterMock{allowRefreshTokenFn: func(string) (bool, int, error) { return false, 9, nil }}
	uc := &authUCMock{refreshFn: func(usecase.RefreshIn) (usecase.RefreshOut, error) { t.Fatal("refresh should not be called"); return usecase.RefreshOut{}, nil }}
	ctl := NewAuthCtl(uc, rl)
	c, rec, err := newJSONContext(http.MethodPost, "/auth/refresh", nil)
	if err != nil { t.Fatal(err) }
	c.SetCookie(&http.Cookie{Name: refreshCookieName, Value: "token"})

	if err := ctl.Refresh(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusTooManyRequests { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestAuthCtlMeSetsNoStore(t *testing.T) {
	actor := &entity.Actor{UserID: 7, Role: entity.RoleUser}
	uc := &authUCMock{meFn: func(a entity.Actor) (entity.User, error) {
		if a.UserID != 7 { t.Fatalf("unexpected actor: %+v", a) }
		return entity.User{ID: 7, Email: "me@test.com", Role: entity.RoleUser}, nil
	}}
	ctl := NewAuthCtl(uc, &rateLimiterMock{})
	c, rec, err := newJSONContext(http.MethodGet, "/me", nil)
	if err != nil { t.Fatal(err) }
	setActor(c, actor)

	if err := ctl.Me(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusOK { t.Fatalf("status=%d", rec.Code) }
	if rec.Header().Get("Cache-Control") != "no-store" { t.Fatalf("cache-control=%s", rec.Header().Get("Cache-Control")) }
}
