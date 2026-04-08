package validator

import (
	"coffee-spa/usecase"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// 保存済み提案の保存・一覧・削除の入力検証。
type savedValidator struct{}

// Saved用validatorを生成。
func NewSavedValidator() usecase.SavedVal {
	return &savedValidator{}
}

// Saveは提案保存入力を検証。
func (v *savedValidator) Save(in usecase.SaveSuggestionIn) error {
	return validation.ValidateStruct(&in,
		// SessionIDは1以上。
		validation.Field(&in.SessionID, validation.Min(uint(1)).Error("session_id must be greater than 0")),
		// SuggestionIDも1以上。
		validation.Field(&in.SuggestionID, validation.Min(uint(1)).Error("suggestion_id must be greater than 0")),
	)
}

// Listは保存済み提案一覧取得入力を検証。
func (v *savedValidator) List(in usecase.ListSavedIn) error {
	return validation.ValidateStruct(&in,
		// 一覧系の基本ページング検証。
		validation.Field(&in.Limit, validation.Min(1), validation.Max(50)),
		validation.Field(&in.Offset, validation.Min(0)),
	)
}

// Deleteは保存済み提案削除入力を検証。
func (v *savedValidator) Delete(in usecase.DeleteSavedIn) error {
	return validation.Validate(in.SuggestionID,
		validation.Min(uint(1)).Error("suggestion_id must be greater than 0"),
	)
}
