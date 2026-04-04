package policy

import (
	"errors"
	"testing"

	"coffee-spa/usecase"
)

// 正しいメールアドレス形式ならエラーにならないことを確認。
func TestEmailPolOK_Valid(t *testing.T) {
	t.Parallel()

	p := NewEmailPol()

	err := p.Ok("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 明らかに不正なメール形式を弾けることを確認。
// 「@ が無い」「形式が壊れている」かをチェック。
func TestEmailPolOK_Invalid(t *testing.T) {
	t.Parallel()

	p := NewEmailPol()

	err := p.Ok("invalid-email")
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// 正常なパスワードが通ることを確認。
func TestPwPolOK_Valid(t *testing.T) {
	t.Parallel()

	p := NewPwPol()

	err := p.Ok("CorrectPW123!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 短すぎるパスワードを弾けることを確認。
func TestPwPolOK_TooShort(t *testing.T) {
	t.Parallel()

	p := NewPwPol()

	err := p.Ok("Ab1!")
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// 許可された kind 値だけが通ることを確認。
func TestKindPolOK_ValidKinds(t *testing.T) {
	t.Parallel()

	p := NewKindPol()

	validKinds := []string{"news", "recipe", "deal", "shop"}

	for _, kind := range validKinds {
		kind := kind
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			if err := p.Ok(kind); err != nil {
				t.Fatalf("kind=%q unexpected error: %v", kind, err)
			}
		})
	}
}

// 許可されていないkindを弾けることを確認。
func TestKindPolOK_InvalidKind(t *testing.T) {
	t.Parallel()

	p := NewKindPol()

	err := p.Ok("unknown")
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// 正常なURLが通ることを確認。
func TestURLPolOK_Valid(t *testing.T) {
	t.Parallel()

	p := NewURLPol()

	err := p.Ok("https://example.com/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// URLとして壊れている値を弾けることを確認。
func TestURLPolOK_Invalid(t *testing.T) {
	t.Parallel()

	p := NewURLPol()

	err := p.Ok("not-a-url")
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// limit / offset の正常な組み合わせが通ることを確認。
func TestPagePolOK_Valid(t *testing.T) {
	t.Parallel()

	p := NewPagePol()

	err := p.Ok(20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 不正な limit を弾けることを確認。
func TestPagePolOK_InvalidLimit(t *testing.T) {
	t.Parallel()

	p := NewPagePol()

	err := p.Ok(999, 0)
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}

// offset が負数のとき弾けることを確認。
func TestPagePolOK_InvalidOffset(t *testing.T) {
	t.Parallel()

	p := NewPagePol()

	err := p.Ok(20, -1)
	if !errors.Is(err, usecase.ErrInvalidRequest) {
		t.Fatalf("err = %v, want ErrInvalidRequest", err)
	}
}
