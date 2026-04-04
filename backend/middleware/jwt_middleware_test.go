package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/repository"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type jwtUserRepoMock struct {
	getByIDFn func(id int64) (entity.User, error)
}

func (m *jwtUserRepoMock) Create(u entity.User) (entity.User, error) {
	return entity.User{}, nil
}
func (m *jwtUserRepoMock) GetByEmail(email string) (entity.User, error) {
	return entity.User{}, repository.ErrNotFound
}
func (m *jwtUserRepoMock) GetByID(id int64) (entity.User, error) {
	return m.getByIDFn(id)
}
func (m *jwtUserRepoMock) SetEmailVerified(userID int64) error               { return nil }
func (m *jwtUserRepoMock) UpdatePassHash(userID int64, newHash string) error { return nil }
func (m *jwtUserRepoMock) BumpTokenVer(userID int64) (int, error)            { return 0, nil }

func TestJWTAuth_MissingBearer_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := JWTAuth("secret")(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestJWTAuth_OK_SetsContext(t *testing.T) {
	t.Parallel()

	maker := usecase.NewJWTMaker("secret")
	token, err := maker.NewAccess(9, "admin", 4)
	if err != nil {
		t.Fatalf("token create error: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := JWTAuth("secret")(func(c echo.Context) error {
		if c.Get("user_id").(int64) != 9 {
			t.Fatalf("user_id = %v, want 9", c.Get("user_id"))
		}
		if c.Get("role").(string) != "admin" {
			t.Fatalf("role = %v, want admin", c.Get("role"))
		}
		if c.Get("tv").(int) != 4 {
			t.Fatalf("tv = %v, want 4", c.Get("tv"))
		}
		return c.NoContent(http.StatusOK)
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestTokenVersion_Mismatch_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", int64(9))
	c.Set("tv", 1)

	h := TokenVersion(&jwtUserRepoMock{
		getByIDFn: func(id int64) (entity.User, error) {
			return entity.User{ID: id, TokenVer: 2}, nil
		},
	})(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestTokenVersion_OK_Passes(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", int64(9))
	c.Set("tv", 2)

	h := TokenVersion(&jwtUserRepoMock{
		getByIDFn: func(id int64) (entity.User, error) {
			return entity.User{ID: id, TokenVer: 2}, nil
		},
	})(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	if err := h(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
