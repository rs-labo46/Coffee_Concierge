package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coffee-spa/controller"
	"coffee-spa/entity"
	"coffee-spa/router"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type healthOK struct{}

func (healthOK) Check() error { return nil }

type tokenVersionReaderStub struct{}

func (tokenVersionReaderStub) GetByID(id uint) (*entity.User, error) {
	return &entity.User{ID: id, TokenVer: 1, Role: entity.RoleUser}, nil
}

type rateLimiterStub struct {
	allowed    bool
	retryAfter int
}

func (m *rateLimiterStub) result() (bool, int, error) {
	if m.retryAfter == 0 {
		m.retryAfter = 60
	}
	return m.allowed, m.retryAfter, nil
}
func (m *rateLimiterStub) AllowSignup(ip string) (bool, int, error)       { return m.result() }
func (m *rateLimiterStub) AllowLoginIP(ip string) (bool, int, error)      { return m.result() }
func (m *rateLimiterStub) AllowLogin(emailHash string) (bool, int, error) { return m.result() }
func (m *rateLimiterStub) AllowRefreshToken(tokenHash string) (bool, int, error) {
	return m.result()
}
func (m *rateLimiterStub) AllowResendIP(ip string) (bool, int, error) { return m.result() }
func (m *rateLimiterStub) AllowResendMail(emailHash string) (bool, int, error) {
	return m.result()
}
func (m *rateLimiterStub) AllowForgotIP(ip string) (bool, int, error) { return m.result() }
func (m *rateLimiterStub) AllowForgotMail(emailHash string) (bool, int, error) {
	return m.result()
}
func (m *rateLimiterStub) AllowWS(key string) (bool, int, error) { return m.result() }

type authUCStub struct{}

func (authUCStub) Signup(in usecase.SignupIn) (usecase.SignupOut, error) {
	return usecase.SignupOut{User: entity.User{ID: 1, Email: in.Email, Role: entity.RoleUser}}, nil
}
func (authUCStub) VerifyEmail(in usecase.VerifyEmailIn) error { return nil }
func (authUCStub) Login(in usecase.LoginIn) (usecase.LoginOut, error) {
	return usecase.LoginOut{User: entity.User{ID: 1, Email: in.Email, Role: entity.RoleUser}, AccessToken: "access", RefreshToken: "refresh"}, nil
}
func (authUCStub) Refresh(in usecase.RefreshIn) (usecase.RefreshOut, error) {
	return usecase.RefreshOut{User: entity.User{ID: 1, Role: entity.RoleUser}, AccessToken: "access", RefreshToken: "refresh"}, nil
}
func (authUCStub) ResendVerify(in usecase.ResendVerifyIn) error         { return nil }
func (authUCStub) Logout(actor entity.Actor, refreshToken string) error { return nil }
func (authUCStub) ForgotPw(in usecase.ForgotPwIn) error                 { return nil }
func (authUCStub) ResetPw(in usecase.ResetPwIn) error                   { return nil }
func (authUCStub) Me(actor entity.Actor) (entity.User, error) {
	return entity.User{ID: actor.UserID, Role: actor.Role, TokenVer: actor.TokenVer}, nil
}

type itemUCStub struct{}

func (itemUCStub) Create(actor entity.Actor, in usecase.CreateItemIn) (entity.Item, error) {
	return entity.Item{ID: 1, Title: in.Title, Kind: in.Kind, SourceID: in.SourceID}, nil
}
func (itemUCStub) Get(id uint) (entity.Item, error) { return entity.Item{ID: id, Title: "item"}, nil }
func (itemUCStub) List(q entity.ItemQ) ([]entity.Item, error) {
	return []entity.Item{{ID: 1, Title: "item", Kind: q.Kind}}, nil
}
func (itemUCStub) Top(limit int) (entity.TopItems, error) {
	return entity.TopItems{News: []entity.Item{{ID: 1, Kind: entity.ItemKindNews}}}, nil
}

type sourceUCStub struct{}

func (sourceUCStub) Create(actor entity.Actor, in usecase.CreateSourceIn) (entity.Source, error) {
	return entity.Source{ID: 1, Name: in.Name, SiteURL: in.SiteURL}, nil
}
func (sourceUCStub) Get(id uint) (entity.Source, error) {
	return entity.Source{ID: id, Name: "source"}, nil
}
func (sourceUCStub) List(limit int, offset int) ([]entity.Source, error) {
	return []entity.Source{{ID: 1, Name: "source"}}, nil
}

type beanUCStub struct{}

func (beanUCStub) Create(actor entity.Actor, in usecase.CreateBeanIn) (entity.Bean, error) {
	return entity.Bean{ID: 1, Name: in.Name, Roast: in.Roast, Active: in.Active}, nil
}
func (beanUCStub) Update(actor entity.Actor, in usecase.UpdateBeanIn) (entity.Bean, error) {
	return entity.Bean{ID: in.ID, Name: in.Name, Roast: in.Roast, Active: in.Active}, nil
}
func (beanUCStub) Get(id uint) (entity.Bean, error) { return entity.Bean{ID: id, Name: "bean"}, nil }
func (beanUCStub) List(in usecase.BeanListIn) ([]entity.Bean, error) {
	return []entity.Bean{{ID: 1, Name: "bean"}}, nil
}

type recipeUCStub struct{}

func (recipeUCStub) Create(actor entity.Actor, in usecase.CreateRecipeIn) (entity.Recipe, error) {
	return entity.Recipe{ID: 1, BeanID: in.BeanID, Name: in.Name, Method: in.Method, TempPref: in.TempPref}, nil
}
func (recipeUCStub) Update(actor entity.Actor, in usecase.UpdateRecipeIn) (entity.Recipe, error) {
	return entity.Recipe{ID: in.ID, BeanID: in.BeanID, Name: in.Name}, nil
}
func (recipeUCStub) Get(id uint) (entity.Recipe, error) {
	return entity.Recipe{ID: id, Name: "recipe"}, nil
}
func (recipeUCStub) List(in usecase.RecipeListIn) ([]entity.Recipe, error) {
	return []entity.Recipe{{ID: 1, Name: "recipe"}}, nil
}

type searchFlowUCStub struct{}

func (searchFlowUCStub) StartSession(in usecase.StartSessionIn) (usecase.StartSessionOut, error) {
	return usecase.StartSessionOut{Session: entity.Session{ID: 1}, SessionKey: "guest-key"}, nil
}
func (searchFlowUCStub) SetPref(in usecase.SetPrefIn) (usecase.SetPrefOut, error) {
	return usecase.SetPrefOut{Pref: entity.Pref{ID: 1, SessionID: in.SessionID}}, nil
}
func (searchFlowUCStub) AddTurn(in usecase.AddTurnIn) (usecase.AddTurnOut, error) {
	return usecase.AddTurnOut{Turn: entity.Turn{ID: 1, SessionID: in.SessionID, Body: in.Body}, Pref: entity.Pref{SessionID: in.SessionID}}, nil
}
func (searchFlowUCStub) PatchPref(in usecase.PatchPrefIn) (usecase.PatchPrefOut, error) {
	return usecase.PatchPrefOut{Pref: entity.Pref{SessionID: in.SessionID}}, nil
}

type sessionUCStub struct{}

func (sessionUCStub) GetSession(in usecase.GetSessionIn) (usecase.GetSessionOut, error) {
	return usecase.GetSessionOut{Session: entity.Session{ID: in.SessionID}}, nil
}
func (sessionUCStub) ListHistory(in usecase.ListHistoryIn) ([]entity.Session, error) {
	uid := in.Actor.UserID
	return []entity.Session{{ID: 1, UserID: &uid}}, nil
}
func (sessionUCStub) CloseSession(in usecase.CloseSessionIn) error { return nil }

type savedUCStub struct{}

func (savedUCStub) Save(in usecase.SaveSuggestionIn) (entity.SavedSuggestion, error) {
	return entity.SavedSuggestion{ID: 1, UserID: in.Actor.UserID, SessionID: in.SessionID, SuggestionID: in.SuggestionID}, nil
}
func (savedUCStub) List(in usecase.ListSavedIn) ([]entity.SavedSuggestion, error) {
	return []entity.SavedSuggestion{}, nil
}
func (savedUCStub) Delete(in usecase.DeleteSavedIn) error { return nil }

type auditUCStub struct{}

func (auditUCStub) List(actor entity.Actor, in usecase.AuditListIn) ([]entity.AuditLog, error) {
	return []entity.AuditLog{{ID: 1, Type: in.Type}}, nil
}

func newTestRouter() *echo.Echo {
	e := echo.New()
	rl := &rateLimiterStub{allowed: true}
	sf := searchFlowUCStub{}
	sess := sessionUCStub{}
	router.New(
		e,
		controller.NewHealthCtl(healthOK{}),
		controller.NewAuthCtl(authUCStub{}, rl),
		controller.NewItemCtl(itemUCStub{}),
		controller.NewSourceCtl(sourceUCStub{}),
		controller.NewBeanCtl(beanUCStub{}),
		controller.NewRecipeCtl(recipeUCStub{}),
		controller.NewSearchCtl(sf, sess),
		controller.NewSavedCtl(savedUCStub{}),
		controller.NewAuditCtl(auditUCStub{}),
		controller.NewWsCtl(sf, sess),
		"test-secret",
		tokenVersionReaderStub{},
		"http://localhost:3000",
		rl,
		rl,
	)
	return e
}

func TestRouter_RequiredRoutesAndProtection(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		want   int
	}{
		{name: "health registered", method: http.MethodGet, path: "/health", want: http.StatusOK},
		{name: "public items top registered", method: http.MethodGet, path: "/items/top", want: http.StatusOK},
		{name: "public sources registered", method: http.MethodGet, path: "/sources", want: http.StatusOK},
		{name: "private me requires jwt", method: http.MethodGet, path: "/me", want: http.StatusUnauthorized},
		{name: "admin items requires jwt before handler", method: http.MethodPost, path: "/admin/items", want: http.StatusUnauthorized},
		{name: "search session start requires login without token", method: http.MethodPost, path: "/search/sessions", want: http.StatusUnauthorized},
		{name: "pref requires login without token", method: http.MethodPost, path: "/search/sessions/1/pref", want: http.StatusUnauthorized},
		{name: "patch pref requires login without token", method: http.MethodPatch, path: "/search/sessions/1/pref", want: http.StatusUnauthorized},
		{name: "guest session detail requires session key", method: http.MethodGet, path: "/search/guest/sessions/1", want: http.StatusBadRequest},
		{name: "refresh route protected by csrf", method: http.MethodPost, path: "/auth/refresh", want: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestRouter()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestRouter_WSGuestRequiresUpgradeAndSessionKey(t *testing.T) {
	e := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/ws/guest/search/sessions/1?session_key=guest", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUpgradeRequired {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
