package validator

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Bean作成・更新・取得・一覧取得の入力検証。
type beanValidator struct{}

// Bean用のvalidatorを生成。
func NewBeanValidator() usecase.BeanVal {
	return &beanValidator{}
}

// Bean作成入力を検証。
func (v *beanValidator) Create(in usecase.CreateBeanIn) error {
	// CreateとUpdateで共通の本体チェックを使い回す。
	return v.validateCreateOrUpdate(
		in.Name,
		in.Roast,
		in.Origin,
		in.Flavor,
		in.Acidity,
		in.Bitterness,
		in.Body,
		in.Aroma,
		in.Desc,
		in.BuyURL,
	)
}

// Bean更新入力を検証。
func (v *beanValidator) Update(in usecase.UpdateBeanIn) error {
	// 更新：まず対象idが1以上か確認。
	if err := validation.Validate(in.ID,
		validation.Min(uint(1)).Error("id must be greater than 0"),
	); err != nil {
		return err
	}

	// 更新内容の本体をCreateと同じルールで検証。
	return v.validateCreateOrUpdate(
		in.Name,
		in.Roast,
		in.Origin,
		in.Flavor,
		in.Acidity,
		in.Bitterness,
		in.Body,
		in.Aroma,
		in.Desc,
		in.BuyURL,
	)
}

// Bean詳細取得時のidを検証。
func (v *beanValidator) Get(id uint) error {
	return validation.Validate(id,
		validation.Min(uint(1)).Error("id must be greater than 0"),
	)
}

// Bean一覧取得時の条件を検証。
func (v *beanValidator) List(in usecase.BeanListIn) error {
	return validation.ValidateStruct(&in,
		// Roastは未指定なら許可し、指定された時だけenumを検証。
		validation.Field(
			&in.Roast,
			validation.When(
				in.Roast != "",
				validation.In(entity.RoastLight, entity.RoastMedium, entity.RoastDark).Error("roast is invalid"),
			),
		),
		// limitは1〜50に制限。
		validation.Field(
			&in.Limit,
			validation.Min(1).Error("limit must be at least 1"),
			validation.Max(50).Error("limit must be at most 50"),
		),
		// offsetは0以上に制限。
		validation.Field(
			&in.Offset,
			validation.Min(0).Error("offset must be 0 or more"),
		),
	)
}

// Beanの作成・更新で共通の本体項目を検証。
func (v *beanValidator) validateCreateOrUpdate(
	name string,
	roast entity.Roast,
	origin string,
	flavor int,
	acidity int,
	bitterness int,
	body int, aroma int,
	desc string,
	buyURL string,
) error {
	// ozzo-validation の Field は、ValidateStruct に渡した構造体のフィールドポインタを渡す必要がある。
	// ローカル変数のポインタを直接渡すと、field cannot be found になるため、一度構造体へ詰める。
	in := struct {
		Name       string
		Roast      entity.Roast
		Origin     string
		Flavor     int
		Acidity    int
		Bitterness int
		Body       int
		Aroma      int
		Desc       string
		BuyURL     string
	}{
		Name:       name,
		Roast:      roast,
		Origin:     origin,
		Flavor:     flavor,
		Acidity:    acidity,
		Bitterness: bitterness,
		Body:       body,
		Aroma:      aroma,
		Desc:       desc,
		BuyURL:     buyURL,
	}

	return validation.ValidateStruct(&in,
		// Nameは必須で1〜100文字。
		validation.Field(&in.Name,
			validation.Required.Error("name is required"),
			validation.RuneLength(1, 100).Error("name must be 1 to 100 chars"),
		),
		// Roastは必須で、許可enumのどれかである必要がある。
		validation.Field(&in.Roast,
			validation.Required.Error("roast is required"),
			validation.In(entity.RoastLight, entity.RoastMedium, entity.RoastDark).Error("roast is invalid"),
		),
		// Originは必須で1〜100文字。
		validation.Field(&in.Origin,
			validation.Required.Error("origin is required"),
			validation.RuneLength(1, 100).Error("origin must be 1 to 100 chars"),
		),
		// 各味覚スコアは必須で1〜5。
		validation.Field(&in.Flavor, validation.Required.Error("flavor is required"), validation.Min(1), validation.Max(5)),
		validation.Field(&in.Acidity, validation.Required.Error("acidity is required"), validation.Min(1), validation.Max(5)),
		validation.Field(&in.Bitterness, validation.Required.Error("bitterness is required"), validation.Min(1), validation.Max(5)),
		validation.Field(&in.Body, validation.Required.Error("body is required"), validation.Min(1), validation.Max(5)),
		validation.Field(&in.Aroma, validation.Required.Error("aroma is required"), validation.Min(1), validation.Max(5)),
		// Descは必須で1〜1000文字。
		validation.Field(&in.Desc,
			validation.Required.Error("desc is required"),
			validation.RuneLength(1, 1000).Error("desc must be 1 to 1000 chars"),
		),
		// BuyURLは必須でURL形式でなければならない。
		validation.Field(&in.BuyURL,
			validation.Required.Error("buy_url is required"),
			is.URL.Error("buy_url must be valid"),
		),
	)
}
