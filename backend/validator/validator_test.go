package validator

import (
	"errors"
	"testing"

	"coffee-spa/policy"
	"coffee-spa/usecase"
)

// 正常な email / password なら Signup用バリデーションが通ることを確認。
func TestAuthValidatorSignup_OK(t *testing.T) {
	t.Parallel()

	v := NewAuthValidator(
		policy.NewEmailPol(),
		policy.NewPwPol(),
	)

	err := v.Signup("user@example.com", "CorrectPW123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 不正なメールアドレスを弾けることを確認。
func TestAuthValidatorSignup_InvalidEmail(t *testing.T) {
	t.Parallel()

	v := NewAuthValidator(
		policy.NewEmailPol(),
		policy.NewPwPol(),
	)

	err := v.Signup("bad-email", "CorrectPW123!")
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// 不正なパスワードを弾けることを確認。
func TestAuthValidatorSignup_InvalidPassword(t *testing.T) {
	t.Parallel()

	v := NewAuthValidator(
		policy.NewEmailPol(),
		policy.NewPwPol(),
	)

	err := v.Signup("user@example.com", "short")
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// 正常な item 入力が通ることを確認。
func TestItemValidatorNewItem_OK(t *testing.T) {
	t.Parallel()

	v := NewItemValidator(
		policy.NewKindPol(),
		policy.NewURLPol(),
		policy.NewPagePol(),
	)

	err := v.NewItem(usecase.AddItemIn{
		Title:       "Coffee News",
		Kind:        "news",
		SourceID:    1,
		PublishedAt: "2026-03-12T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 空白だけの title を弾けることを確認。
func TestItemValidatorNewItem_TitleTrimmedEmpty(t *testing.T) {
	t.Parallel()

	v := NewItemValidator(
		policy.NewKindPol(),
		policy.NewURLPol(),
		policy.NewPagePol(),
	)

	err := v.NewItem(usecase.AddItemIn{
		Title:       "   ",
		Kind:        "news",
		SourceID:    1,
		PublishedAt: "2026-03-12T00:00:00Z",
	})
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// 許可されていないkindを弾けることを確認。
func TestItemValidatorNewItem_InvalidKind(t *testing.T) {
	t.Parallel()

	v := NewItemValidator(
		policy.NewKindPol(),
		policy.NewURLPol(),
		policy.NewPagePol(),
	)

	err := v.NewItem(usecase.AddItemIn{
		Title:       "Coffee News",
		Kind:        "invalid-kind",
		SourceID:    1,
		PublishedAt: "2026-03-12T00:00:00Z",
	})
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// URL項目がある場合に、不正URLを弾けることを確認。
func TestItemValidatorNewItem_InvalidURL(t *testing.T) {
	t.Parallel()

	v := NewItemValidator(
		policy.NewKindPol(),
		policy.NewURLPol(),
		policy.NewPagePol(),
	)

	url := "bad-url"

	err := v.NewItem(usecase.AddItemIn{
		Title:       "Coffee News",
		Kind:        "news",
		URL:         &url,
		SourceID:    1,
		PublishedAt: "2026-03-12T00:00:00Z",
	})
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// 一覧検索の limit / offset / kind が正常なら通ることを確認。
func TestItemValidatorListItem_OK(t *testing.T) {
	t.Parallel()

	v := NewItemValidator(
		policy.NewKindPol(),
		policy.NewURLPol(),
		policy.NewPagePol(),
	)

	err := v.ListItem(usecase.ItemQ{
		Kind:   "news",
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 一覧検索のページング条件が不正なとき弾けることを確認。
func TestItemValidatorListItem_InvalidPage(t *testing.T) {
	t.Parallel()

	v := NewItemValidator(
		policy.NewKindPol(),
		policy.NewURLPol(),
		policy.NewPagePol(),
	)

	err := v.ListItem(usecase.ItemQ{
		Kind:   "news",
		Limit:  999,
		Offset: 0,
	})
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// 正常な source 入力が通ることを確認。
func TestSourceValidatorNewSource_OK(t *testing.T) {
	t.Parallel()

	v := NewSourceValidator(policy.NewURLPol())

	siteURL := "https://example.com"

	err := v.NewSource(usecase.AddSourceIn{
		Name:    "Coffee Source",
		SiteURL: &siteURL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 空白だけのnameを弾けることを確認。
func TestSourceValidatorNewSource_TrimmedEmptyName(t *testing.T) {
	t.Parallel()

	v := NewSourceValidator(policy.NewURLPol())

	err := v.NewSource(usecase.AddSourceIn{
		Name: "   ",
	})
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// SiteURL が不正なときに弾けることを確認。
func TestSourceValidatorNewSource_InvalidURL(t *testing.T) {
	t.Parallel()

	v := NewSourceValidator(policy.NewURLPol())

	siteURL := "bad-url"

	err := v.NewSource(usecase.AddSourceIn{
		Name:    "Coffee Source",
		SiteURL: &siteURL,
	})
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}
