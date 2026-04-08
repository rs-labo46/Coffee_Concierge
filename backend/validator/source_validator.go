package validator

import (
	"coffee-spa/usecase"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Source作成・取得・一覧取得で使う入力チェックをまとめる。
type sourceValidator struct{}

// NewSourceValidator は Source 用 validator を生成。
// usecase側にはinterfaceを返し、具体実装を隠す。
func NewSourceValidator() usecase.SourceVal {
	return &sourceValidator{}
}

// Source作成時の入力を検証。
func (v *sourceValidator) Create(in usecase.CreateSourceIn) error {
	return validation.ValidateStruct(&in,
		// Nameは必須で、1〜100文字に制限。
		validation.Field(
			&in.Name,
			validation.Required.Error("name is required"),
			validation.RuneLength(1, 100).Error("name must be 1 to 100 chars"),
		),
		// SiteURLは必須で、URLとして正しい形式かを見る。
		validation.Field(
			&in.SiteURL,
			validation.Required.Error("site_url is required"),
			is.URL.Error("site_url must be valid url"),
		),
	)
}

// Source詳細取得時のidを検証。
func (v *sourceValidator) Get(id uint) error {
	return validation.Validate(id,
		// id は1以上。
		validation.Min(uint(1)).Error("id must be greater than 0"),
	)
}

// Source 一覧取得時のlimit / offsetを検証。
func (v *sourceValidator) List(limit int, offset int) error {
	return validation.ValidateStruct(&struct {
		// Limitは取得件数。
		Limit int
		// Offsetは読み飛ばし件数。
		Offset int
	}{
		Limit:  limit,
		Offset: offset,
	},
		// limitは最低1、最大50。
		validation.Field(
			&limit,
			validation.Required,
			validation.Min(1).Error("limit must be at least 1"),
			validation.Max(50).Error("limit must be at most 50"),
		),
		// offsetは0以上。
		validation.Field(
			&offset,
			validation.Min(0).Error("offset must be 0 or more"),
		),
	)
}
