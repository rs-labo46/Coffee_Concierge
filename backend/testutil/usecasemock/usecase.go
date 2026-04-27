package usecasemock

import (
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

type Auth struct {
	T              *testing.T
	SignupFn       func(usecase.SignupIn) (usecase.SignupOut, error)
	VerifyEmailFn  func(usecase.VerifyEmailIn) error
	LoginFn        func(usecase.LoginIn) (usecase.LoginOut, error)
	RefreshFn      func(usecase.RefreshIn) (usecase.RefreshOut, error)
	ResendVerifyFn func(usecase.ResendVerifyIn) error
	LogoutFn       func(entity.Actor, string) error
	ForgotPwFn     func(usecase.ForgotPwIn) error
	ResetPwFn      func(usecase.ResetPwIn) error
	MeFn           func(entity.Actor) (entity.User, error)
}

func (m *Auth) Signup(in usecase.SignupIn) (usecase.SignupOut, error) {
	if m.SignupFn == nil {
		return usecase.SignupOut{User: entity.User{ID: 1, Email: in.Email, Role: entity.RoleUser}}, nil
	}
	return m.SignupFn(in)
}
func (m *Auth) VerifyEmail(in usecase.VerifyEmailIn) error {
	if m.VerifyEmailFn == nil {
		return nil
	}
	return m.VerifyEmailFn(in)
}
func (m *Auth) Login(in usecase.LoginIn) (usecase.LoginOut, error) {
	if m.LoginFn == nil {
		return usecase.LoginOut{User: entity.User{ID: 1, Email: in.Email, Role: entity.RoleUser}, AccessToken: "access", RefreshToken: "refresh"}, nil
	}
	return m.LoginFn(in)
}
func (m *Auth) Refresh(in usecase.RefreshIn) (usecase.RefreshOut, error) {
	if m.RefreshFn == nil {
		return usecase.RefreshOut{User: entity.User{ID: 1, Role: entity.RoleUser}, AccessToken: "access", RefreshToken: "refresh"}, nil
	}
	return m.RefreshFn(in)
}
func (m *Auth) ResendVerify(in usecase.ResendVerifyIn) error {
	if m.ResendVerifyFn == nil {
		return nil
	}
	return m.ResendVerifyFn(in)
}
func (m *Auth) Logout(a entity.Actor, refreshToken string) error {
	if m.LogoutFn == nil {
		return nil
	}
	return m.LogoutFn(a, refreshToken)
}
func (m *Auth) ForgotPw(in usecase.ForgotPwIn) error {
	if m.ForgotPwFn == nil {
		return nil
	}
	return m.ForgotPwFn(in)
}
func (m *Auth) ResetPw(in usecase.ResetPwIn) error {
	if m.ResetPwFn == nil {
		return nil
	}
	return m.ResetPwFn(in)
}
func (m *Auth) Me(a entity.Actor) (entity.User, error) {
	if m.MeFn == nil {
		return entity.User{ID: a.UserID, Email: "me@example.com", Role: a.Role, TokenVer: a.TokenVer}, nil
	}
	return m.MeFn(a)
}

type SearchFlow struct {
	StartSessionFn func(usecase.StartSessionIn) (usecase.StartSessionOut, error)
	SetPrefFn      func(usecase.SetPrefIn) (usecase.SetPrefOut, error)
	AddTurnFn      func(usecase.AddTurnIn) (usecase.AddTurnOut, error)
	PatchPrefFn    func(usecase.PatchPrefIn) (usecase.PatchPrefOut, error)
}

func (m *SearchFlow) StartSession(in usecase.StartSessionIn) (usecase.StartSessionOut, error) {
	if m.StartSessionFn == nil {
		return usecase.StartSessionOut{Session: entity.Session{ID: 1}, SessionKey: "guest-key"}, nil
	}
	return m.StartSessionFn(in)
}
func (m *SearchFlow) SetPref(in usecase.SetPrefIn) (usecase.SetPrefOut, error) {
	if m.SetPrefFn == nil {
		return usecase.SetPrefOut{Pref: entity.Pref{ID: 1, SessionID: in.SessionID}}, nil
	}
	return m.SetPrefFn(in)
}
func (m *SearchFlow) AddTurn(in usecase.AddTurnIn) (usecase.AddTurnOut, error) {
	if m.AddTurnFn == nil {
		return usecase.AddTurnOut{Turn: entity.Turn{ID: 1, SessionID: in.SessionID, Body: in.Body}, Pref: entity.Pref{SessionID: in.SessionID}}, nil
	}
	return m.AddTurnFn(in)
}
func (m *SearchFlow) PatchPref(in usecase.PatchPrefIn) (usecase.PatchPrefOut, error) {
	if m.PatchPrefFn == nil {
		return usecase.PatchPrefOut{Pref: entity.Pref{SessionID: in.SessionID}}, nil
	}
	return m.PatchPrefFn(in)
}

type Session struct {
	GetSessionFn   func(usecase.GetSessionIn) (usecase.GetSessionOut, error)
	ListHistoryFn  func(usecase.ListHistoryIn) ([]entity.Session, error)
	CloseSessionFn func(usecase.CloseSessionIn) error
}

func (m *Session) GetSession(in usecase.GetSessionIn) (usecase.GetSessionOut, error) {
	if m.GetSessionFn == nil {
		return usecase.GetSessionOut{Session: entity.Session{ID: in.SessionID}}, nil
	}
	return m.GetSessionFn(in)
}
func (m *Session) ListHistory(in usecase.ListHistoryIn) ([]entity.Session, error) {
	if m.ListHistoryFn == nil {
		uid := in.Actor.UserID
		return []entity.Session{{ID: 1, UserID: &uid}}, nil
	}
	return m.ListHistoryFn(in)
}
func (m *Session) CloseSession(in usecase.CloseSessionIn) error {
	if m.CloseSessionFn == nil {
		return nil
	}
	return m.CloseSessionFn(in)
}

type Saved struct {
	SaveFn   func(usecase.SaveSuggestionIn) (entity.SavedSuggestion, error)
	ListFn   func(usecase.ListSavedIn) ([]entity.SavedSuggestion, error)
	DeleteFn func(usecase.DeleteSavedIn) error
}

func (m *Saved) Save(in usecase.SaveSuggestionIn) (entity.SavedSuggestion, error) {
	if m.SaveFn == nil {
		return entity.SavedSuggestion{ID: 1, UserID: in.Actor.UserID, SessionID: in.SessionID, SuggestionID: in.SuggestionID}, nil
	}
	return m.SaveFn(in)
}
func (m *Saved) List(in usecase.ListSavedIn) ([]entity.SavedSuggestion, error) {
	if m.ListFn == nil {
		return []entity.SavedSuggestion{}, nil
	}
	return m.ListFn(in)
}
func (m *Saved) Delete(in usecase.DeleteSavedIn) error {
	if m.DeleteFn == nil {
		return nil
	}
	return m.DeleteFn(in)
}
