package controller

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

func newJSONContext(method string, path string, body interface{}) (echo.Context, *httptest.ResponseRecorder, error) {
	e := echo.New()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, nil, err
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec, nil
}

func setActor(c echo.Context, actor *entity.Actor) {
	if actor != nil {
		c.Set(ContextActorKey, actor)
	}
}

func mustJSONBody(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(rec.Body.Bytes(), v); err != nil {
		t.Fatalf("unmarshal response: %v, body=%s", err, rec.Body.String())
	}
}

type authUCMock struct {
	signupFn       func(usecase.SignupIn) (usecase.SignupOut, error)
	verifyEmailFn  func(usecase.VerifyEmailIn) error
	loginFn        func(usecase.LoginIn) (usecase.LoginOut, error)
	refreshFn      func(usecase.RefreshIn) (usecase.RefreshOut, error)
	resendVerifyFn func(usecase.ResendVerifyIn) error
	logoutFn       func(entity.Actor, string) error
	forgotPwFn     func(usecase.ForgotPwIn) error
	resetPwFn      func(usecase.ResetPwIn) error
	meFn           func(entity.Actor) (entity.User, error)
}

func (m *authUCMock) Signup(in usecase.SignupIn) (usecase.SignupOut, error) { return m.signupFn(in) }
func (m *authUCMock) VerifyEmail(in usecase.VerifyEmailIn) error { return m.verifyEmailFn(in) }
func (m *authUCMock) Login(in usecase.LoginIn) (usecase.LoginOut, error) { return m.loginFn(in) }
func (m *authUCMock) Refresh(in usecase.RefreshIn) (usecase.RefreshOut, error) { return m.refreshFn(in) }
func (m *authUCMock) ResendVerify(in usecase.ResendVerifyIn) error { return m.resendVerifyFn(in) }
func (m *authUCMock) Logout(actor entity.Actor, token string) error { return m.logoutFn(actor, token) }
func (m *authUCMock) ForgotPw(in usecase.ForgotPwIn) error { return m.forgotPwFn(in) }
func (m *authUCMock) ResetPw(in usecase.ResetPwIn) error { return m.resetPwFn(in) }
func (m *authUCMock) Me(actor entity.Actor) (entity.User, error) { return m.meFn(actor) }

type rateLimiterMock struct {
	allowSignupFn       func(string) (bool, int, error)
	allowLoginIPFn      func(string) (bool, int, error)
	allowLoginFn        func(string) (bool, int, error)
	allowRefreshTokenFn func(string) (bool, int, error)
	allowResendIPFn     func(string) (bool, int, error)
	allowResendMailFn   func(string) (bool, int, error)
	allowForgotIPFn     func(string) (bool, int, error)
	allowForgotMailFn   func(string) (bool, int, error)
	allowWSFn           func(string) (bool, int, error)
}

func (m *rateLimiterMock) AllowSignup(ip string) (bool, int, error) { return m.allowSignupFn(ip) }
func (m *rateLimiterMock) AllowLoginIP(ip string) (bool, int, error) { return m.allowLoginIPFn(ip) }
func (m *rateLimiterMock) AllowLogin(emailHash string) (bool, int, error) { return m.allowLoginFn(emailHash) }
func (m *rateLimiterMock) AllowRefreshToken(tokenHash string) (bool, int, error) { return m.allowRefreshTokenFn(tokenHash) }
func (m *rateLimiterMock) AllowResendIP(ip string) (bool, int, error) { return m.allowResendIPFn(ip) }
func (m *rateLimiterMock) AllowResendMail(emailHash string) (bool, int, error) { return m.allowResendMailFn(emailHash) }
func (m *rateLimiterMock) AllowForgotIP(ip string) (bool, int, error) { return m.allowForgotIPFn(ip) }
func (m *rateLimiterMock) AllowForgotMail(emailHash string) (bool, int, error) { return m.allowForgotMailFn(emailHash) }
func (m *rateLimiterMock) AllowWS(key string) (bool, int, error) { return m.allowWSFn(key) }

type sourceUCMock struct {
	createFn func(entity.Actor, usecase.CreateSourceIn) (entity.Source, error)
	getFn    func(uint) (entity.Source, error)
	listFn   func(int, int) ([]entity.Source, error)
}

func (m *sourceUCMock) Create(actor entity.Actor, in usecase.CreateSourceIn) (entity.Source, error) { return m.createFn(actor, in) }
func (m *sourceUCMock) Get(id uint) (entity.Source, error) { return m.getFn(id) }
func (m *sourceUCMock) List(limit int, offset int) ([]entity.Source, error) { return m.listFn(limit, offset) }

type itemUCMock struct {
	createFn func(entity.Actor, usecase.CreateItemIn) (entity.Item, error)
	getFn    func(uint) (entity.Item, error)
	listFn   func(entity.ItemQ) ([]entity.Item, error)
	topFn    func(int) (entity.TopItems, error)
}

func (m *itemUCMock) Create(actor entity.Actor, in usecase.CreateItemIn) (entity.Item, error) { return m.createFn(actor, in) }
func (m *itemUCMock) Get(id uint) (entity.Item, error) { return m.getFn(id) }
func (m *itemUCMock) List(q entity.ItemQ) ([]entity.Item, error) { return m.listFn(q) }
func (m *itemUCMock) Top(limit int) (entity.TopItems, error) { return m.topFn(limit) }

type beanUCMock struct {
	createFn func(entity.Actor, usecase.CreateBeanIn) (entity.Bean, error)
	updateFn func(entity.Actor, usecase.UpdateBeanIn) (entity.Bean, error)
	getFn    func(uint) (entity.Bean, error)
	listFn   func(usecase.BeanListIn) ([]entity.Bean, error)
}

func (m *beanUCMock) Create(actor entity.Actor, in usecase.CreateBeanIn) (entity.Bean, error) { return m.createFn(actor, in) }
func (m *beanUCMock) Update(actor entity.Actor, in usecase.UpdateBeanIn) (entity.Bean, error) { return m.updateFn(actor, in) }
func (m *beanUCMock) Get(id uint) (entity.Bean, error) { return m.getFn(id) }
func (m *beanUCMock) List(in usecase.BeanListIn) ([]entity.Bean, error) { return m.listFn(in) }

type recipeUCMock struct {
	createFn func(entity.Actor, usecase.CreateRecipeIn) (entity.Recipe, error)
	updateFn func(entity.Actor, usecase.UpdateRecipeIn) (entity.Recipe, error)
	getFn    func(uint) (entity.Recipe, error)
	listFn   func(usecase.RecipeListIn) ([]entity.Recipe, error)
}

func (m *recipeUCMock) Create(actor entity.Actor, in usecase.CreateRecipeIn) (entity.Recipe, error) { return m.createFn(actor, in) }
func (m *recipeUCMock) Update(actor entity.Actor, in usecase.UpdateRecipeIn) (entity.Recipe, error) { return m.updateFn(actor, in) }
func (m *recipeUCMock) Get(id uint) (entity.Recipe, error) { return m.getFn(id) }
func (m *recipeUCMock) List(in usecase.RecipeListIn) ([]entity.Recipe, error) { return m.listFn(in) }

type savedUCMock struct {
	saveFn   func(usecase.SaveSuggestionIn) (entity.SavedSuggestion, error)
	listFn   func(usecase.ListSavedIn) ([]entity.SavedSuggestion, error)
	deleteFn func(usecase.DeleteSavedIn) error
}

func (m *savedUCMock) Save(in usecase.SaveSuggestionIn) (entity.SavedSuggestion, error) { return m.saveFn(in) }
func (m *savedUCMock) List(in usecase.ListSavedIn) ([]entity.SavedSuggestion, error) { return m.listFn(in) }
func (m *savedUCMock) Delete(in usecase.DeleteSavedIn) error { return m.deleteFn(in) }

type auditUCMock struct { listFn func(entity.Actor, usecase.AuditListIn) ([]entity.AuditLog, error) }
func (m *auditUCMock) List(actor entity.Actor, in usecase.AuditListIn) ([]entity.AuditLog, error) { return m.listFn(actor, in) }

type healthUCMock struct { checkFn func() error }
func (m *healthUCMock) Check() error { return m.checkFn() }

type searchFlowUCMock struct {
	startSessionFn func(usecase.StartSessionIn) (usecase.StartSessionOut, error)
	setPrefFn      func(usecase.SetPrefIn) (usecase.SetPrefOut, error)
	addTurnFn      func(usecase.AddTurnIn) (usecase.AddTurnOut, error)
	patchPrefFn    func(usecase.PatchPrefIn) (usecase.PatchPrefOut, error)
}
func (m *searchFlowUCMock) StartSession(in usecase.StartSessionIn) (usecase.StartSessionOut, error) { return m.startSessionFn(in) }
func (m *searchFlowUCMock) SetPref(in usecase.SetPrefIn) (usecase.SetPrefOut, error) { return m.setPrefFn(in) }
func (m *searchFlowUCMock) AddTurn(in usecase.AddTurnIn) (usecase.AddTurnOut, error) { return m.addTurnFn(in) }
func (m *searchFlowUCMock) PatchPref(in usecase.PatchPrefIn) (usecase.PatchPrefOut, error) { return m.patchPrefFn(in) }

type sessionUCMock struct {
	getSessionFn   func(usecase.GetSessionIn) (usecase.GetSessionOut, error)
	listHistoryFn  func(usecase.ListHistoryIn) ([]entity.Session, error)
	closeSessionFn func(usecase.CloseSessionIn) error
}
func (m *sessionUCMock) GetSession(in usecase.GetSessionIn) (usecase.GetSessionOut, error) { return m.getSessionFn(in) }
func (m *sessionUCMock) ListHistory(in usecase.ListHistoryIn) ([]entity.Session, error) { return m.listHistoryFn(in) }
func (m *sessionUCMock) CloseSession(in usecase.CloseSessionIn) error { return m.closeSessionFn(in) }
