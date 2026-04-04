package validator

import (
	"strings"

	"coffee-spa/policy"
	"coffee-spa/usecase"
)

type AuthValidator struct {
	email policy.EmailPol
	pw    policy.PwPol
}

type ItemValidator struct {
	kind policy.KindPol
	url  policy.URLPol
	page policy.PagePol
}

type SourceValidator struct {
	url policy.URLPol
}

func NewAuthValidator(
	email policy.EmailPol,
	pw policy.PwPol,
) *AuthValidator {
	return &AuthValidator{
		email: email,
		pw:    pw,
	}
}

func NewItemValidator(
	kind policy.KindPol,
	url policy.URLPol,
	page policy.PagePol,
) *ItemValidator {
	return &ItemValidator{
		kind: kind,
		url:  url,
		page: page,
	}
}

func NewSourceValidator(
	url policy.URLPol,
) *SourceValidator {
	return &SourceValidator{
		url: url,
	}
}

func (v *AuthValidator) Signup(email string, pw string) error {
	if err := v.EmailOnly(email); err != nil {
		return err
	}
	if err := v.pw.Ok(pw); err != nil {
		return err
	}
	return nil
}
func (v *AuthValidator) EmailOnly(email string) error {
	return v.email.Ok(strings.TrimSpace(email))
}

func (v *AuthValidator) Login(email string, pw string) error {
	return v.Signup(email, pw)
}

func (v *AuthValidator) NewPw(pw string) error {
	return v.pw.Ok(pw)
}

func (v *ItemValidator) NewItem(itemInput usecase.AddItemIn) error {
	title := strings.TrimSpace(itemInput.Title)
	if title == "" || len(title) > 120 {
		return usecase.ErrInvalidRequest
	}

	if itemInput.Summary != nil {
		summary := strings.TrimSpace(*itemInput.Summary)
		if len(summary) > 500 {
			return usecase.ErrInvalidRequest
		}
	}

	if itemInput.URL != nil {
		raw := strings.TrimSpace(*itemInput.URL)
		if err := v.url.Ok(raw); err != nil {
			return err
		}
	}

	if itemInput.ImageURL != nil {
		raw := strings.TrimSpace(*itemInput.ImageURL)
		if err := v.url.Ok(raw); err != nil {
			return err
		}
	}

	if err := v.kind.Ok(strings.TrimSpace(itemInput.Kind)); err != nil {
		return err
	}

	if itemInput.SourceID <= 0 {
		return usecase.ErrInvalidRequest
	}

	if strings.TrimSpace(itemInput.PublishedAt) == "" {
		return usecase.ErrInvalidRequest
	}

	return nil
}

func (v *ItemValidator) ListItem(q usecase.ItemQ) error {
	if err := v.page.Ok(q.Limit, q.Offset); err != nil {
		return err
	}

	if q.Kind == "" {
		return nil
	}

	return v.kind.Ok(strings.TrimSpace(q.Kind))
}

func (v *SourceValidator) NewSource(sourceInput usecase.AddSourceIn) error {
	name := strings.TrimSpace(sourceInput.Name)
	if name == "" || len(name) > 80 {
		return usecase.ErrInvalidRequest
	}

	if sourceInput.SiteURL == nil {
		return nil
	}

	return v.url.Ok(strings.TrimSpace(*sourceInput.SiteURL))
}
