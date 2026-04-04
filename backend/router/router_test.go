package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"coffee-spa/controller"
	"coffee-spa/entity"
	"coffee-spa/repository"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type mockHealthUC struct{}

func (m *mockHealthUC) Check() error { return nil }

type mockAuthUCForRouter struct{}

func (m *mockAuthUCForRouter) Signup(in usecase.SignupIn) (entity.User, error) {
	return entity.User{}, nil
}
func (m *mockAuthUCForRouter) VerifyEmail(in usecase.VerifyEmailIn) error   { return nil }
func (m *mockAuthUCForRouter) ResendVerify(in usecase.ResendVerifyIn) error { return nil }
func (m *mockAuthUCForRouter) Login(in usecase.LoginIn) (usecase.AuthOut, error) {
	return usecase.AuthOut{}, nil
}
func (m *mockAuthUCForRouter) Refresh(in usecase.RefreshIn) (usecase.AuthOut, error) {
	return usecase.AuthOut{}, nil
}
func (m *mockAuthUCForRouter) Logout(in usecase.LogoutIn) error     { return nil }
func (m *mockAuthUCForRouter) ForgotPw(in usecase.ForgotPwIn) error { return nil }
func (m *mockAuthUCForRouter) ResetPw(in usecase.ResetPwIn) error   { return nil }
func (m *mockAuthUCForRouter) Me(userID int64) (entity.User, error) {
	return entity.User{ID: userID, Email: "user@example.com", Role: "user", EmailVerified: true}, nil
}

type mockItemUCForRouter struct{}

func (m *mockItemUCForRouter) Add(actor usecase.Actor, in usecase.AddItemIn) (entity.Item, error) {
	return entity.Item{}, nil
}
func (m *mockItemUCForRouter) Get(id int64) (entity.Item, error) {
	return entity.Item{ID: id, Title: "detail", Kind: "news", Source: entity.Source{ID: 1, Name: "Coffee Daily"}}, nil
}
func (m *mockItemUCForRouter) Search(q usecase.ItemQ) ([]entity.Item, error) {
	return []entity.Item{}, nil
}
func (m *mockItemUCForRouter) Top(limit int) (usecase.TopItems, error) {
	return usecase.TopItems{News: []entity.Item{}, Recipe: []entity.Item{}, Deal: []entity.Item{}, Shop: []entity.Item{}}, nil
}

type mockSourceUCForRouter struct{}

func (m *mockSourceUCForRouter) Add(actor usecase.Actor, in usecase.AddSourceIn) (entity.Source, error) {
	return entity.Source{}, nil
}
func (m *mockSourceUCForRouter) List() ([]entity.Source, error) { return []entity.Source{}, nil }

type mockUserRepoForRouter struct{}

func (m *mockUserRepoForRouter) Create(u entity.User) (entity.User, error) { return entity.User{}, nil }
func (m *mockUserRepoForRouter) GetByID(id int64) (entity.User, error) {
	return entity.User{ID: id, Role: "user", TokenVer: 1}, nil
}
func (m *mockUserRepoForRouter) GetByEmail(email string) (entity.User, error) {
	return entity.User{}, repository.ErrNotFound
}
func (m *mockUserRepoForRouter) SetEmailVerified(userID int64) error                { return nil }
func (m *mockUserRepoForRouter) UpdatePassHash(userID int64, passHash string) error { return nil }
func (m *mockUserRepoForRouter) BumpTokenVer(userID int64) (int, error)             { return 2, nil }

func newTestEcho() *echo.Echo {
	e := echo.New()

	healthCtl := controller.NewHealthCtl(&mockHealthUC{})
	authCtl := controller.NewAuthCtl(&mockAuthUCForRouter{})
	itemCtl := controller.NewItemCtl(&mockItemUCForRouter{})
	srcCtl := controller.NewSrcCtl(&mockSourceUCForRouter{})

	New(e, healthCtl, authCtl, itemCtl, srcCtl, "test-secret", &mockUserRepoForRouter{}, "http://localhost:3000")
	return e
}

func TestRouter_PublicItemsTop_Exists(t *testing.T) {
	t.Parallel()
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodGet, "/items/top", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body controller.TopItemsRes
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
}

func TestRouter_PublicItemsDetail_Exists(t *testing.T) {
	t.Parallel()
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodGet, "/items/9", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body controller.ItemDetailRes
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if body.Item.ID != 9 || body.Source.Name != "Coffee Daily" {
		t.Fatalf("unexpected detail body: %+v", body)
	}
}

func TestRouter_PublicItemsList_Exists(t *testing.T) {
	t.Parallel()
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRouter_PublicSources_Exists(t *testing.T) {
	t.Parallel()
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodGet, "/sources", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestRouter_RefreshRoute_RequiresRefreshCookie(t *testing.T) {
	t.Parallel()
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.Header.Set("X-CSRF-Token", "csrf")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "csrf"})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestRouter_LogoutRoute_RequiresJWT(t *testing.T) {
	t.Parallel()
	e := newTestEcho()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
