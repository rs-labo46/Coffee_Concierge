package validator

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Item作成・取得・一覧取得・Top 表示時の入力チェック。
type itemValidator struct{}

// Item用validatorを生成。
func NewItemValidator() usecase.ItemVal {
	return &itemValidator{}
}

// Item作成時の入力を検証。
func (v *itemValidator) Create(in usecase.CreateItemIn) error {
	return validation.ValidateStruct(&in,
		// Titleは必須で、1〜200文字以内に制限。
		validation.Field(
			&in.Title,
			validation.Required.Error("title is required"),
			validation.RuneLength(1, 200).Error("title must be 1 to 200 chars"),
		),
		// Summaryは必須で、1〜1000文字以内に制限。
		validation.Field(
			&in.Summary,
			validation.Required.Error("summary is required"),
			validation.RuneLength(1, 1000).Error("summary must be 1 to 1000 chars"),
		),
		// URLは必須で、リンクとして妥当な形式。
		validation.Field(
			&in.URL,
			validation.Required.Error("url is required"),
			is.URL.Error("url must be valid"),
		),
		// ImageURLも必須で、画像URLとして妥当な形式。
		validation.Field(
			&in.ImageURL,
			validation.Required.Error("image_url is required"),
			is.URL.Error("image_url must be valid"),
		),
		// Kindは許可されたItemKindのどれかでなければならない。
		validation.Field(
			&in.Kind,
			validation.Required.Error("kind is required"),
			validation.In(
				entity.ItemKindNews,
				entity.ItemKindRecipe,
				entity.ItemKindDeal,
				entity.ItemKindShop,
			).Error("kind is invalid"),
		),
		// SourceIDは1以上。
		validation.Field(
			&in.SourceID,
			validation.Min(uint(1)).Error("source_id must be greater than 0"),
		),
		// PublishedAtは必須。
		validation.Field(
			&in.PublishedAt,
			validation.Required.Error("published_at is required"),
		),
	)
}

// Item詳細取得時のidを検証。
func (v *itemValidator) Get(id uint) error {
	return validation.Validate(id,
		validation.Min(uint(1)).Error("id must be greater than 0"),
	)
}

// Item 一覧取得時の検索条件を検証。
func (v *itemValidator) List(q entity.ItemQ) error {
	return validation.ValidateStruct(&q,
		// Kindは空なら未指定として許可し、値が入る時だけenumを検証。
		validation.Field(
			&q.Kind,
			validation.When(
				q.Kind != "",
				validation.In(
					entity.ItemKindNews,
					entity.ItemKindRecipe,
					entity.ItemKindDeal,
					entity.ItemKindShop,
				).Error("kind is invalid"),
			),
		),
		// limitは1〜50に制限。
		validation.Field(
			&q.Limit,
			validation.Min(1).Error("limit must be at least 1"),
			validation.Max(50).Error("limit must be at most 50"),
		),
		// offsetは0以上。
		validation.Field(
			&q.Offset,
			validation.Min(0).Error("offset must be 0 or more"),
		),
	)
}

// Top表示用Item取得時の件数だけを検証。
func (v *itemValidator) Top(limit int) error {
	return validation.Validate(limit,
		validation.Min(1).Error("limit must be at least 1"),
		validation.Max(50).Error("limit must be at most 50"),
	)
}
