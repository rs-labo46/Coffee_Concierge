package repomock

import (
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

type Audit struct {
	T        *testing.T
	CreateFn func(*entity.AuditLog) error
	ListFn   func(repository.AuditListQ) ([]entity.AuditLog, error)
}

func (m *Audit) Create(log *entity.AuditLog) error {
	if m.CreateFn == nil {
		return nil
	}
	return m.CreateFn(log)
}

func (m *Audit) List(q repository.AuditListQ) ([]entity.AuditLog, error) {
	if m.ListFn == nil {
		return []entity.AuditLog{}, nil
	}
	return m.ListFn(q)
}

type User struct {
	T                *testing.T
	CreateFn         func(*entity.User) error
	GetByIDFn        func(uint) (*entity.User, error)
	GetByEmailFn     func(string) (*entity.User, error)
	UpdateFn         func(*entity.User) error
	UpdateTokenVerFn func(uint, int) error
}

func (m *User) Create(user *entity.User) error {
	if m.CreateFn == nil {
		return nil
	}
	return m.CreateFn(user)
}
func (m *User) GetByID(id uint) (*entity.User, error) {
	if m.GetByIDFn == nil {
		return &entity.User{ID: id, Role: entity.RoleUser, TokenVer: 1}, nil
	}
	return m.GetByIDFn(id)
}
func (m *User) GetByEmail(email string) (*entity.User, error) {
	if m.GetByEmailFn == nil {
		return &entity.User{ID: 1, Email: email, Role: entity.RoleUser, TokenVer: 1, EmailVerified: true}, nil
	}
	return m.GetByEmailFn(email)
}
func (m *User) Update(user *entity.User) error {
	if m.UpdateFn == nil {
		return nil
	}
	return m.UpdateFn(user)
}
func (m *User) UpdateTokenVer(userID uint, tokenVer int) error {
	if m.UpdateTokenVerFn == nil {
		return nil
	}
	return m.UpdateTokenVerFn(userID, tokenVer)
}

type EmailVerify struct {
	CreateFn         func(*entity.EmailVerify) error
	GetByTokenHashFn func(string) (*entity.EmailVerify, error)
	MarkUsedFn       func(uint, time.Time) error
	DeleteExpiredFn  func(time.Time) error
}

func (m *EmailVerify) Create(v *entity.EmailVerify) error {
	if m.CreateFn == nil {
		return nil
	}
	return m.CreateFn(v)
}
func (m *EmailVerify) GetByTokenHash(tokenHash string) (*entity.EmailVerify, error) {
	if m.GetByTokenHashFn == nil {
		return &entity.EmailVerify{ID: 1, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	return m.GetByTokenHashFn(tokenHash)
}
func (m *EmailVerify) MarkUsed(id uint, usedAt time.Time) error {
	if m.MarkUsedFn == nil {
		return nil
	}
	return m.MarkUsedFn(id, usedAt)
}
func (m *EmailVerify) DeleteExpired(now time.Time) error {
	if m.DeleteExpiredFn == nil {
		return nil
	}
	return m.DeleteExpiredFn(now)
}

type PwReset struct {
	CreateFn         func(*entity.PwReset) error
	GetByTokenHashFn func(string) (*entity.PwReset, error)
	MarkUsedFn       func(uint, time.Time) error
	DeleteExpiredFn  func(time.Time) error
}

func (m *PwReset) Create(v *entity.PwReset) error {
	if m.CreateFn == nil {
		return nil
	}
	return m.CreateFn(v)
}
func (m *PwReset) GetByTokenHash(tokenHash string) (*entity.PwReset, error) {
	if m.GetByTokenHashFn == nil {
		return &entity.PwReset{ID: 1, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	return m.GetByTokenHashFn(tokenHash)
}
func (m *PwReset) MarkUsed(id uint, usedAt time.Time) error {
	if m.MarkUsedFn == nil {
		return nil
	}
	return m.MarkUsedFn(id, usedAt)
}
func (m *PwReset) DeleteExpired(now time.Time) error {
	if m.DeleteExpiredFn == nil {
		return nil
	}
	return m.DeleteExpiredFn(now)
}

type Rt struct {
	CreateFn         func(*entity.Rt) error
	GetByTokenHashFn func(string) (*entity.Rt, error)
	UpdateFn         func(*entity.Rt) error
	RevokeFamilyFn   func(string, time.Time) error
	DeleteExpiredFn  func(time.Time) error
}

func (m *Rt) Create(v *entity.Rt) error {
	if m.CreateFn == nil {
		return nil
	}
	return m.CreateFn(v)
}
func (m *Rt) GetByTokenHash(tokenHash string) (*entity.Rt, error) {
	if m.GetByTokenHashFn == nil {
		return &entity.Rt{ID: 1, UserID: 1, FamilyID: "family", TokenHash: tokenHash, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}
	return m.GetByTokenHashFn(tokenHash)
}
func (m *Rt) Update(v *entity.Rt) error {
	if m.UpdateFn == nil {
		return nil
	}
	return m.UpdateFn(v)
}
func (m *Rt) RevokeFamily(familyID string, revokedAt time.Time) error {
	if m.RevokeFamilyFn == nil {
		return nil
	}
	return m.RevokeFamilyFn(familyID, revokedAt)
}
func (m *Rt) DeleteExpired(now time.Time) error {
	if m.DeleteExpiredFn == nil {
		return nil
	}
	return m.DeleteExpiredFn(now)
}

type Session struct {
	CreateSessionFn       func(*entity.Session) error
	GetSessionByIDFn      func(uint) (*entity.Session, error)
	GetGuestSessionByIDFn func(uint, string, time.Time) (*entity.Session, error)
	ListHistoryFn         func(repository.HistoryQ) ([]entity.Session, error)
	CloseSessionFn        func(uint) error
	CreateTurnFn          func(*entity.Turn) error
	ListTurnsFn           func(uint) ([]entity.Turn, error)
	CreatePrefFn          func(*entity.Pref) error
	UpdatePrefFn          func(*entity.Pref) error
	GetPrefBySessionIDFn  func(uint) (*entity.Pref, error)
	ReplaceSuggestionsFn  func(uint, []entity.Suggestion) error
	ListSuggestionsFn     func(uint) ([]entity.Suggestion, error)
	GetSuggestionByIDFn   func(uint) (*entity.Suggestion, error)
}

func (m *Session) CreateSession(s *entity.Session) error {
	if m.CreateSessionFn == nil {
		if s.ID == 0 {
			s.ID = 1
		}
		return nil
	}
	return m.CreateSessionFn(s)
}
func (m *Session) GetSessionByID(id uint) (*entity.Session, error) {
	if m.GetSessionByIDFn == nil {
		uid := uint(1)
		return &entity.Session{ID: id, UserID: &uid, Status: entity.SessionActive}, nil
	}
	return m.GetSessionByIDFn(id)
}
func (m *Session) GetGuestSessionByID(id uint, h string, now time.Time) (*entity.Session, error) {
	if m.GetGuestSessionByIDFn == nil {
		return &entity.Session{ID: id, Status: entity.SessionActive, SessionKeyHash: h}, nil
	}
	return m.GetGuestSessionByIDFn(id, h, now)
}
func (m *Session) ListHistory(q repository.HistoryQ) ([]entity.Session, error) {
	if m.ListHistoryFn == nil {
		return []entity.Session{{ID: 1, UserID: &q.UserID}}, nil
	}
	return m.ListHistoryFn(q)
}
func (m *Session) CloseSession(id uint) error {
	if m.CloseSessionFn == nil {
		return nil
	}
	return m.CloseSessionFn(id)
}
func (m *Session) CreateTurn(t *entity.Turn) error {
	if m.CreateTurnFn == nil {
		if t.ID == 0 {
			t.ID = 1
		}
		return nil
	}
	return m.CreateTurnFn(t)
}
func (m *Session) ListTurns(id uint) ([]entity.Turn, error) {
	if m.ListTurnsFn == nil {
		return []entity.Turn{}, nil
	}
	return m.ListTurnsFn(id)
}
func (m *Session) CreatePref(p *entity.Pref) error {
	if m.CreatePrefFn == nil {
		if p.ID == 0 {
			p.ID = 1
		}
		return nil
	}
	return m.CreatePrefFn(p)
}
func (m *Session) UpdatePref(p *entity.Pref) error {
	if m.UpdatePrefFn == nil {
		return nil
	}
	return m.UpdatePrefFn(p)
}
func (m *Session) GetPrefBySessionID(id uint) (*entity.Pref, error) {
	if m.GetPrefBySessionIDFn == nil {
		return &entity.Pref{ID: 1, SessionID: id, Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3, Mood: entity.MoodRelax, Method: entity.MethodDrip, Scene: entity.SceneRelax, TempPref: entity.TempHot}, nil
	}
	return m.GetPrefBySessionIDFn(id)
}
func (m *Session) ReplaceSuggestions(id uint, s []entity.Suggestion) error {
	if m.ReplaceSuggestionsFn == nil {
		return nil
	}
	return m.ReplaceSuggestionsFn(id, s)
}
func (m *Session) ListSuggestions(id uint) ([]entity.Suggestion, error) {
	if m.ListSuggestionsFn == nil {
		return []entity.Suggestion{}, nil
	}
	return m.ListSuggestionsFn(id)
}
func (m *Session) GetSuggestionByID(id uint) (*entity.Suggestion, error) {
	if m.GetSuggestionByIDFn == nil {
		return &entity.Suggestion{ID: id, SessionID: 1, BeanID: 1}, nil
	}
	return m.GetSuggestionByIDFn(id)
}

type Saved struct {
	CreateFn                      func(*entity.SavedSuggestion) error
	ListFn                        func(repository.SavedListQ) ([]entity.SavedSuggestion, error)
	DeleteByUserAndSuggestionIDFn func(uint, uint) error
	GetByUserAndSuggestionIDFn    func(uint, uint) (*entity.SavedSuggestion, error)
}

func (m *Saved) Create(s *entity.SavedSuggestion) error {
	if m.CreateFn == nil {
		if s.ID == 0 {
			s.ID = 1
		}
		return nil
	}
	return m.CreateFn(s)
}
func (m *Saved) List(q repository.SavedListQ) ([]entity.SavedSuggestion, error) {
	if m.ListFn == nil {
		return []entity.SavedSuggestion{}, nil
	}
	return m.ListFn(q)
}
func (m *Saved) DeleteByUserAndSuggestionID(u uint, s uint) error {
	if m.DeleteByUserAndSuggestionIDFn == nil {
		return nil
	}
	return m.DeleteByUserAndSuggestionIDFn(u, s)
}
func (m *Saved) GetByUserAndSuggestionID(u uint, s uint) (*entity.SavedSuggestion, error) {
	if m.GetByUserAndSuggestionIDFn == nil {
		return &entity.SavedSuggestion{ID: 1, UserID: u, SuggestionID: s}, nil
	}
	return m.GetByUserAndSuggestionIDFn(u, s)
}

type RateLimit struct {
	AllowFn      func(string, float64, float64, float64, time.Time) (bool, int, error)
	LastKey      string
	LastRate     float64
	LastCapacity float64
	LastCost     float64
}

func (m *RateLimit) Allow(key string, rate float64, capacity float64, cost float64, now time.Time) (bool, int, error) {
	m.LastKey = key
	m.LastRate = rate
	m.LastCapacity = capacity
	m.LastCost = cost
	if m.AllowFn == nil {
		return true, 0, nil
	}
	return m.AllowFn(key, rate, capacity, cost, now)
}

type Bean struct {
	CreateFn       func(*entity.Bean) error
	UpdateFn       func(*entity.Bean) error
	GetByIDFn      func(uint) (*entity.Bean, error)
	ListFn         func(repository.BeanListQ) ([]entity.Bean, error)
	SearchByPrefFn func(entity.Pref, int) ([]entity.Bean, error)
}

func (m *Bean) Create(b *entity.Bean) error {
	if m.CreateFn == nil {
		if b.ID == 0 {
			b.ID = 1
		}
		return nil
	}
	return m.CreateFn(b)
}
func (m *Bean) Update(b *entity.Bean) error {
	if m.UpdateFn == nil {
		return nil
	}
	return m.UpdateFn(b)
}
func (m *Bean) GetByID(id uint) (*entity.Bean, error) {
	if m.GetByIDFn == nil {
		return &entity.Bean{ID: id, Name: "bean", Active: true, Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3}, nil
	}
	return m.GetByIDFn(id)
}
func (m *Bean) List(q repository.BeanListQ) ([]entity.Bean, error) {
	if m.ListFn == nil {
		return []entity.Bean{}, nil
	}
	return m.ListFn(q)
}
func (m *Bean) SearchByPref(p entity.Pref, limit int) ([]entity.Bean, error) {
	if m.SearchByPrefFn == nil {
		return []entity.Bean{{ID: 1, Name: "bean", Active: true, Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3}}, nil
	}
	return m.SearchByPrefFn(p, limit)
}

type Recipe struct {
	CreateFn            func(*entity.Recipe) error
	UpdateFn            func(*entity.Recipe) error
	GetByIDFn           func(uint) (*entity.Recipe, error)
	ListFn              func(repository.RecipeListQ) ([]entity.Recipe, error)
	FindPrimaryByBeanFn func(uint, entity.Method, entity.TempPref) (*entity.Recipe, error)
}

func (m *Recipe) Create(r *entity.Recipe) error {
	if m.CreateFn == nil {
		if r.ID == 0 {
			r.ID = 1
		}
		return nil
	}
	return m.CreateFn(r)
}
func (m *Recipe) Update(r *entity.Recipe) error {
	if m.UpdateFn == nil {
		return nil
	}
	return m.UpdateFn(r)
}
func (m *Recipe) GetByID(id uint) (*entity.Recipe, error) {
	if m.GetByIDFn == nil {
		return &entity.Recipe{ID: id, BeanID: 1, Name: "recipe"}, nil
	}
	return m.GetByIDFn(id)
}
func (m *Recipe) List(q repository.RecipeListQ) ([]entity.Recipe, error) {
	if m.ListFn == nil {
		return []entity.Recipe{}, nil
	}
	return m.ListFn(q)
}
func (m *Recipe) FindPrimaryByBean(id uint, method entity.Method, temp entity.TempPref) (*entity.Recipe, error) {
	if m.FindPrimaryByBeanFn == nil {
		return &entity.Recipe{ID: 1, BeanID: id, Method: method, TempPref: temp}, nil
	}
	return m.FindPrimaryByBeanFn(id, method, temp)
}

type Item struct {
	CreateFn        func(*entity.Item) error
	GetByIDFn       func(uint) (*entity.Item, error)
	ListFn          func(repository.ItemListQ) ([]entity.Item, error)
	TopFn           func(int) (*entity.TopItems, error)
	SearchRelatedFn func(string, entity.Roast, string, entity.Mood, entity.Method, int, time.Time) ([]entity.Item, error)
}

func (m *Item) Create(i *entity.Item) error {
	if m.CreateFn == nil {
		if i.ID == 0 {
			i.ID = 1
		}
		return nil
	}
	return m.CreateFn(i)
}
func (m *Item) GetByID(id uint) (*entity.Item, error) {
	if m.GetByIDFn == nil {
		return &entity.Item{ID: id, Title: "item"}, nil
	}
	return m.GetByIDFn(id)
}
func (m *Item) List(q repository.ItemListQ) ([]entity.Item, error) {
	if m.ListFn == nil {
		return []entity.Item{}, nil
	}
	return m.ListFn(q)
}
func (m *Item) Top(limit int) (*entity.TopItems, error) {
	if m.TopFn == nil {
		return &entity.TopItems{}, nil
	}
	return m.TopFn(limit)
}
func (m *Item) SearchRelated(name string, roast entity.Roast, origin string, mood entity.Mood, method entity.Method, limit int, now time.Time) ([]entity.Item, error) {
	if m.SearchRelatedFn == nil {
		return []entity.Item{}, nil
	}
	return m.SearchRelatedFn(name, roast, origin, mood, method, limit, now)
}

type Source struct {
	CreateFn  func(*entity.Source) error
	GetByIDFn func(uint) (*entity.Source, error)
	ListFn    func(repository.SourceListQ) ([]entity.Source, error)
}

func (m *Source) Create(s *entity.Source) error {
	if m.CreateFn == nil {
		if s.ID == 0 {
			s.ID = 1
		}
		return nil
	}
	return m.CreateFn(s)
}
func (m *Source) GetByID(id uint) (*entity.Source, error) {
	if m.GetByIDFn == nil {
		return &entity.Source{ID: id, Name: "source"}, nil
	}
	return m.GetByIDFn(id)
}
func (m *Source) List(q repository.SourceListQ) ([]entity.Source, error) {
	if m.ListFn == nil {
		return []entity.Source{}, nil
	}
	return m.ListFn(q)
}
