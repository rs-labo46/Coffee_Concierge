package usecase

import (
	"errors"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

var ErrInvalidRequest = errors.New("invalid_request")
var ErrUnauthorized = errors.New("unauthorized")
var ErrForbidden = errors.New("forbidden")
var ErrNotFound = errors.New("not_found")
var ErrConflict = errors.New("conflict")
var ErrInternal = errors.New("internal")

type ErrRateLimited struct {
	RetryAfterSec int
}

func (e ErrRateLimited) Error() string {
	return "rate_limited"
}

type Actor struct {
	UserID int64
	Role   string
	IP     string
	UA     string
}

type ItemQ = repository.ItemQ

type TopItems = repository.TopItems

type SignupIn struct {
	Email string
	Pw    string
	IP    string
	UA    string
}

type LoginIn struct {
	Email string
	Pw    string
	IP    string
	UA    string
}

type RefreshIn struct {
	RefreshToken string
	IP           string
	UA           string
}

type LogoutIn struct {
	UserID       int64
	RefreshToken string
	IP           string
	UA           string
}

type ForgotPwIn struct {
	Email string
	IP    string
	UA    string
}

type ResetPwIn struct {
	Token string
	NewPw string
	IP    string
	UA    string
}

type VerifyEmailIn struct {
	Token string
	IP    string
	UA    string
}

type ResendVerifyIn struct {
	Email string
	IP    string
	UA    string
}

type AuthOut struct {
	AccessToken  string
	RefreshToken string
	CsrfToken    string
	User         entity.User
}

type AddItemIn struct {
	Title       string
	Summary     *string
	Body        *string
	URL         *string
	ImageURL    *string
	Kind        string
	SourceID    int64
	PublishedAt string
}

type AddSourceIn struct {
	Name    string
	SiteURL *string
}

type AuthUsecase interface {
	Signup(in SignupIn) (entity.User, error)
	VerifyEmail(in VerifyEmailIn) error
	ResendVerify(in ResendVerifyIn) error
	Login(in LoginIn) (AuthOut, error)
	Refresh(in RefreshIn) (AuthOut, error)
	Logout(in LogoutIn) error
	ForgotPw(in ForgotPwIn) error
	ResetPw(in ResetPwIn) error
	Me(userID int64) (entity.User, error)
}

type ItemUsecase interface {
	Add(actor Actor, in AddItemIn) (entity.Item, error)
	Get(id int64) (entity.Item, error)
	Search(q ItemQ) ([]entity.Item, error)
	Top(limit int) (TopItems, error)
}

type SourceUsecase interface {
	Add(actor Actor, in AddSourceIn) (entity.Source, error)
	List() ([]entity.Source, error)
}

type HealthUsecase interface {
	Check() error
}
