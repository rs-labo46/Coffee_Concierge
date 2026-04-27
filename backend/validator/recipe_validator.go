package validator

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Recipeの作成・更新・取得・一覧取得の入力検証。
type recipeValidator struct{}

// Recipe用validatorを生成。
func NewRecipeValidator() usecase.RecipeVal {
	return &recipeValidator{}
}

// Recipe作成入力を検証。
func (v *recipeValidator) Create(in usecase.CreateRecipeIn) error {
	// 紐づく BeanID が正しいか先に見る。
	if err := validation.Validate(in.BeanID,
		validation.Min(uint(1)).Error("bean_id must be greater than 0"),
	); err != nil {
		return err
	}

	// 本体項目の検証は共通関数へ寄せる。
	return v.validateCreateOrUpdate(
		in.Name,
		in.Method,
		in.TempPref,
		in.Grind,
		in.Ratio,
		in.Temp,
		in.TimeSec,
		in.Steps,
		in.Desc,
	)
}

// Recipeの更新入力を検証。
func (v *recipeValidator) Update(in usecase.UpdateRecipeIn) error {
	// 更新対象のidをチェック。
	if err := validation.Validate(in.ID,
		validation.Min(uint(1)).Error("id must be greater than 0"),
	); err != nil {
		return err
	}

	// 紐づくBeanIDも正しい必要がある。
	if err := validation.Validate(in.BeanID,
		validation.Min(uint(1)).Error("bean_id must be greater than 0"),
	); err != nil {
		return err
	}

	// 本体項目はCreateと同じルールで検証。
	return v.validateCreateOrUpdate(
		in.Name,
		in.Method,
		in.TempPref,
		in.Grind,
		in.Ratio,
		in.Temp,
		in.TimeSec,
		in.Steps,
		in.Desc,
	)
}

// Recipe詳細取得時のidを検証。
func (v *recipeValidator) Get(id uint) error {
	return validation.Validate(id,
		validation.Min(uint(1)).Error("id must be greater than 0"),
	)
}

// Recipeの一覧取得時の条件を検証。
func (v *recipeValidator) List(in usecase.RecipeListIn) error {
	return validation.ValidateStruct(&in,
		// BeanIDは指定された時だけ1以上かを見る。
		validation.Field(
			&in.BeanID,
			validation.When(in.BeanID != nil, validation.Min(uint(1)).Error("bean_id must be greater than 0")),
		),
		// Methodは未指定なら許可、指定時だけenumを検証。
		validation.Field(
			&in.Method,
			validation.When(
				in.Method != "",
				validation.In(
					entity.MethodDrip,
					entity.MethodEspresso,
					entity.MethodMilk,
					entity.MethodIced,
				).Error("method is invalid"),
			),
		),
		// TempPrefも未指定なら許可、指定時だけenumを検証。
		validation.Field(
			&in.TempPref,
			validation.When(
				in.TempPref != "",
				validation.In(entity.TempHot, entity.TempIce).Error("temp_pref is invalid"),
			),
		),
		// ページング値の検証。
		validation.Field(&in.Limit, validation.Min(1), validation.Max(50)),
		validation.Field(&in.Offset, validation.Min(0)),
	)
}

// Recipeの本体項目を共通で検証。
func (v *recipeValidator) validateCreateOrUpdate(
	name string,
	method entity.Method,
	tempPref entity.TempPref,
	grind string,
	ratio string,
	temp int,
	timeSec int,
	steps []string,
	desc string,
) error {
	// ozzo-validation の Field は、ValidateStruct に渡した構造体のフィールドポインタを渡す必要がある。
	// ローカル変数のポインタを直接渡すと、field cannot be found になるため、一度構造体へ詰める。
	in := struct {
		Name     string
		Method   entity.Method
		TempPref entity.TempPref
		Grind    string
		Ratio    string
		Temp     int
		TimeSec  int
		Steps    []string
		Desc     string
	}{
		Name:     name,
		Method:   method,
		TempPref: tempPref,
		Grind:    grind,
		Ratio:    ratio,
		Temp:     temp,
		TimeSec:  timeSec,
		Steps:    steps,
		Desc:     desc,
	}

	return validation.ValidateStruct(&in,
		// Nameは必須で1〜100文字。
		validation.Field(&in.Name,
			validation.Required.Error("name is required"),
			validation.RuneLength(1, 100).Error("name must be 1 to 100 chars"),
		),
		// Methodは必須で、許可されたenumのどれかでなければならない。
		validation.Field(&in.Method,
			validation.Required.Error("method is required"),
			validation.In(
				entity.MethodDrip,
				entity.MethodEspresso,
				entity.MethodMilk,
				entity.MethodIced,
			).Error("method is invalid"),
		),
		// TempPref必須でenumを検証。
		validation.Field(&in.TempPref,
			validation.Required.Error("temp_pref is required"),
			validation.In(entity.TempHot, entity.TempIce).Error("temp_pref is invalid"),
		),
		// Grind必須で1〜200文字。
		validation.Field(&in.Grind,
			validation.Required.Error("grind is required"),
			validation.RuneLength(1, 200).Error("grind must be 1 to 200 chars"),
		),
		// Ratio必須で1〜200文字。
		validation.Field(&in.Ratio,
			validation.Required.Error("ratio is required"),
			validation.RuneLength(1, 200).Error("ratio must be 1 to 200 chars"),
		),
		// 温度は必須で60〜100に制限。
		validation.Field(&in.Temp, validation.Required.Error("temp is required"), validation.Min(60), validation.Max(100)),
		// 秒数は必須で1〜600秒に制限。
		validation.Field(&in.TimeSec, validation.Required.Error("time_sec is required"), validation.Min(1), validation.Max(600)),
		// steps必須で、1〜20件に制限。
		validation.Field(&in.Steps,
			validation.Required.Error("steps is required"),
			validation.Length(1, 20).Error("steps count must be 1 to 20"),
		),
		// Desc必須で1〜1000文字。
		validation.Field(&in.Desc,
			validation.Required.Error("desc is required"),
			validation.RuneLength(1, 1000).Error("desc must be 1 to 1000 chars"),
		),
	)
}
