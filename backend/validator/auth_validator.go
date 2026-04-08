package validator

import (
	"coffee-spa/usecase"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// authValidator は AuthVal の具体実装。
// サインアップ、ログイン、パスワード再設定、トークン入力の検証を担当。
type authValidator struct{}

// Auth用validatorを生成。
// usecase側にはinterfaceを返して、具体実装を隠す。
func NewAuthValidator() usecase.AuthVal {
	return &authValidator{}
}

// サインアップ時のemail/passwordを検証。
func (v *authValidator) Signup(email string, pw string) error {
	in := struct {
		Email    string
		Password string
	}{
		Email:    email,
		Password: pw,
	}

	return validation.ValidateStruct(&in,
		// Emailは必須。
		// メール形式も確認。
		validation.Field(
			&in.Email,
			validation.Required.Error("email is required"),
			validation.RuneLength(1, 254).Error("email must be 1 to 254 chars"),
			is.Email.Error("email must be valid"),
		),
		// Passwordは必須。
		//72文字以下、最低8文字以上に制限。
		validation.Field(
			&in.Password,
			validation.Required.Error("password is required"),
			validation.RuneLength(8, 72).Error("password must be 8 to 72 chars"),
		),
	)
}

// ログイン時の email / password を検証。
func (v *authValidator) Login(email string, pw string) error {
	in := struct {
		Email    string
		Password string
	}{
		Email:    email,
		Password: pw,
	}

	return validation.ValidateStruct(&in,
		// Emailは必須かつメール形式。
		validation.Field(
			&in.Email,
			validation.Required.Error("email is required"),
			validation.RuneLength(1, 254).Error("email must be 1 to 254 chars"),
			is.Email.Error("email must be valid"),
		),
		// Passwordは必須。
		validation.Field(
			&in.Password,
			validation.Required.Error("password is required"),
			validation.RuneLength(1, 72).Error("password must be 1 to 72 chars"),
		),
	)
}

// 新しいパスワードの形式を検証。
func (v *authValidator) NewPw(pw string) error {
	return validation.Validate(pw,
		// 新しいパスワードは必須。
		validation.Required.Error("password is required"),
		validation.RuneLength(8, 72).Error("password must be 8 to 72 chars"),
	)
}

// Tokenはverify/reset用などのtoken文字列を検証。
func (v *authValidator) Token(token string) error {
	return validation.Validate(token,
		// tokenは必須。
		validation.Required.Error("token is required"),
		validation.RuneLength(16, 512).Error("token must be 16 to 512 chars"),
	)
}
