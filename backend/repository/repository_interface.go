package repository

import (
	"errors"
	"time"

	"coffee-spa/entity"
)

// repository共通エラー。
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidState = errors.New("invalid state")
	ErrRateLimited  = errors.New("rate limited")
	ErrInternal     = errors.New("internal")
)

// Beanの一覧・検索用の条件(qはname/originなどの部分一致検索)。
type BeanListQ struct {
	Q      string
	Roast  entity.Roast
	Active *bool
	Limit  int
	Offset int
}

// Recipe 一覧・検索用の条件(Beanの詳細からの一覧取得と、抽出条件による検索の両方で使う)。
type RecipeListQ struct {
	BeanID   *uint
	Method   entity.Method
	TempPref entity.TempPref
	Active   *bool
	Limit    int
	Offset   int
}

// Source一覧取得条件。
type SourceListQ struct {
	Limit  int
	Offset int
}

// Item一覧・検索条件。
// qはtitle/summaryの部分一致、kindはカテゴリ絞り込み。
type ItemListQ struct {
	Q      string
	Kind   entity.ItemKind
	Limit  int
	Offset int
}

// 認証ユーザーの履歴一覧取得条件。
type HistoryQ struct {
	UserID uint
	Limit  int
	Offset int
}

// 保存済み提案一覧取得条件。
type SavedListQ struct {
	UserID uint
	Limit  int
	Offset int
}

// 監査一覧取得条件。
type AuditListQ struct {
	Type   string
	UserID *uint
	Limit  int
	Offset int
}

// 認証・取得・token version更新に使う。
type UserRepository interface {
	Create(user *entity.User) error
	GetByID(id uint) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	Update(user *entity.User) error
	UpdateTokenVer(userID uint, tokenVer int) error
}

// メール認証 tokenの発行・取得・使用済み。
type EmailVerifyRepository interface {
	Create(v *entity.EmailVerify) error
	GetByTokenHash(tokenHash string) (*entity.EmailVerify, error)
	MarkUsed(id uint, usedAt time.Time) error
	DeleteExpired(now time.Time) error
}

// パスワード再設定 tokenの発行・取得・使用済み。
type PwResetRepository interface {
	Create(r *entity.PwReset) error
	GetByTokenHash(tokenHash string) (*entity.PwReset, error)
	MarkUsed(id uint, usedAt time.Time) error
	DeleteExpired(now time.Time) error
}

// refresh rotation、reuse検知、family 失効の土台。
type RtRepository interface {
	Create(rt *entity.Rt) error
	GetByTokenHash(tokenHash string) (*entity.Rt, error)
	Update(rt *entity.Rt) error
	RevokeFamily(familyID string, revokedAt time.Time) error
	DeleteExpired(now time.Time) error
}

// 主要イベントの追加と管理画面向け一覧取得。
type AuditRepository interface {
	Create(log *entity.AuditLog) error
	List(q AuditListQ) ([]entity.AuditLog, error)
}

// sources。
type SourceRepository interface {
	Create(src *entity.Source) error
	GetByID(id uint) (*entity.Source, error)
	List(q SourceListQ) ([]entity.Source, error)
}

// 一覧、詳細、top 表示、関連Item検索。
type ItemRepository interface {
	Create(item *entity.Item) error
	GetByID(id uint) (*entity.Item, error)
	List(q ItemListQ) ([]entity.Item, error)
	Top(limit int) (*entity.TopItems, error)
	SearchRelated(
		beanName string,
		roast entity.Roast,
		origin string,
		mood entity.Mood,
		method entity.Method,
		limit int,
		now time.Time,
	) ([]entity.Item, error)
}

// BeanのCRUDと条件検索。
type BeanRepository interface {
	Create(bean *entity.Bean) error
	Update(bean *entity.Bean) error
	GetByID(id uint) (*entity.Bean, error)
	List(q BeanListQ) ([]entity.Bean, error)
	SearchByPref(pref entity.Pref, limit int) ([]entity.Bean, error)
}

// BeanごとのRecipe選定や一覧取得。
type RecipeRepository interface {
	Create(recipe *entity.Recipe) error
	Update(recipe *entity.Recipe) error
	GetByID(id uint) (*entity.Recipe, error)
	List(q RecipeListQ) ([]entity.Recipe, error)
	FindPrimaryByBean(beanID uint, method entity.Method, tempPref entity.TempPref) (*entity.Recipe, error)
}

// sessions/turns/prefs/suggestions 検索対話フローの永続。
type SessionRepository interface {
	CreateSession(session *entity.Session) error
	GetSessionByID(id uint) (*entity.Session, error)
	GetGuestSessionByID(id uint, sessionKeyHash string, now time.Time) (*entity.Session, error)
	ListHistory(q HistoryQ) ([]entity.Session, error)
	CloseSession(id uint) error

	CreateTurn(turn *entity.Turn) error
	ListTurns(sessionID uint) ([]entity.Turn, error)

	CreatePref(pref *entity.Pref) error
	UpdatePref(pref *entity.Pref) error
	GetPrefBySessionID(sessionID uint) (*entity.Pref, error)

	ReplaceSuggestions(sessionID uint, suggestions []entity.Suggestion) error
	ListSuggestions(sessionID uint) ([]entity.Suggestion, error)
	GetSuggestionByID(id uint) (*entity.Suggestion, error)
}

// 保存・一覧・削除。
type SavedRepository interface {
	Create(saved *entity.SavedSuggestion) error
	List(q SavedListQ) ([]entity.SavedSuggestion, error)
	DeleteByUserAndSuggestionID(userID uint, suggestionID uint) error
	GetByUserAndSuggestionID(userID uint, suggestionID uint) (*entity.SavedSuggestion, error)
}
