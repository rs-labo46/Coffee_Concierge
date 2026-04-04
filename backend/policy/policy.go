package policy

import (
	"net/url"
	"strings"

	"coffee-spa/usecase"
)

// パスワードのルール
type PwPol interface {
	Ok(pw string) error
}

// メールアドレスのルール
type EmailPol interface {
	Ok(email string) error
}

// item kindのルール
type KindPol interface {
	Ok(kind string) error
}

// URLのルール
type URLPol interface {
	Ok(raw string) error
}

// ページングのルール
type PagePol interface {
	Ok(limit int, offset int) error
}

type pwPol struct{}
type emailPol struct{}
type kindPol struct{}
type urlPol struct{}
type pagePol struct{}

func NewPwPol() PwPol {
	return &pwPol{}
}

func NewEmailPol() EmailPol {
	return &emailPol{}
}

func NewKindPol() KindPol {
	return &kindPol{}
}

func NewURLPol() URLPol {
	return &urlPol{}
}

func NewPagePol() PagePol {
	return &pagePol{}
}

func (p *pwPol) Ok(pw string) error {
	n := len(pw)
	if n < 12 || n > 72 {
		return usecase.ErrInvalidRequest
	}
	return nil
}

func (p *emailPol) Ok(email string) error {
	if email == "" {
		return usecase.ErrInvalidRequest
	}
	if len(email) > 254 {
		return usecase.ErrInvalidRequest
	}
	if !strings.Contains(email, "@") {
		return usecase.ErrInvalidRequest
	}
	return nil
}

func (p *kindPol) Ok(kind string) error {
	switch kind {
	case "news", "recipe", "deal", "shop":
		return nil
	default:
		return usecase.ErrInvalidRequest
	}
}

func (p *urlPol) Ok(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return usecase.ErrInvalidRequest
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return usecase.ErrInvalidRequest
	}

	return nil
}

func (p *pagePol) Ok(limit int, offset int) error {
	if limit < 1 || limit > 50 {
		return usecase.ErrInvalidRequest
	}
	if offset < 0 {
		return usecase.ErrInvalidRequest
	}
	return nil
}
