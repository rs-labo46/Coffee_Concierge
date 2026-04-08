package validator

import (
	"coffee-spa/usecase"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// 監査ログ一覧取得時の検索条件を検証。
type auditValidator struct{}

// Audit用validatorを生成。
func NewAuditValidator() usecase.AuditVal {
	return &auditValidator{}
}

// Listは監査ログ一覧取得入力を検証。
func (v *auditValidator) List(in usecase.AuditListIn) error {
	return validation.ValidateStruct(&in,
		// Typeは任意入力。
		// 指定されるなら1〜100文字に制限。
		validation.Field(
			&in.Type,
			validation.When(
				in.Type != "",
				validation.RuneLength(1, 100).Error("type must be 1 to 100 chars"),
			),
		),
		// UserIDは任意入力。
		// 指定される場合だけ1以上かを見る。
		validation.Field(
			&in.UserID,
			validation.When(
				in.UserID != nil,
				validation.Min(uint(1)).Error("user_id must be greater than 0"),
			),
		),
		// 一覧系のページング検証。
		validation.Field(&in.Limit, validation.Min(1), validation.Max(100)),
		validation.Field(&in.Offset, validation.Min(0)),
	)
}
