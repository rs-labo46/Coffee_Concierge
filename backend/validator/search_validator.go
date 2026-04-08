package validator

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// セッション開始、初期条件設定、発話追加、条件差分更新などの入力検証を担当。
type searchValidator struct{}

// Search用validatorを生成。
func NewSearchValidator() usecase.SearchVal {
	return &searchValidator{}
}

// 検索セッション開始入力を検証。
func (v *searchValidator) StartSession(in usecase.StartSessionIn) error {
	return validation.ValidateStruct(&in,
		// Titlは任意入力。1〜200文字に制限。
		validation.Field(
			&in.Title,
			validation.When(
				in.Title != "",
				validation.RuneLength(1, 200).Error("title must be 1 to 200 chars"),
			),
		),
	)
}

// 初回条件設定入力を検証。
func (v *searchValidator) SetPref(in usecase.SetPrefIn) error {
	return validation.ValidateStruct(&in,
		// SessionIDは必須で1以上。
		validation.Field(&in.SessionID, validation.Min(uint(1)).Error("session_id must be greater than 0")),
		// 5軸スコアはすべて1〜5。
		validation.Field(&in.Flavor, validation.Min(1), validation.Max(5)),
		validation.Field(&in.Acidity, validation.Min(1), validation.Max(5)),
		validation.Field(&in.Bitterness, validation.Min(1), validation.Max(5)),
		validation.Field(&in.Body, validation.Min(1), validation.Max(5)),
		validation.Field(&in.Aroma, validation.Min(1), validation.Max(5)),
		// Moodは許可enumを検証。
		validation.Field(&in.Mood, validation.In(entity.MoodMorning, entity.MoodWork, entity.MoodRelax, entity.MoodNight)),
		// Methodは許可enumを検証。
		validation.Field(&in.Method, validation.In(entity.MethodDrip, entity.MethodEspresso, entity.MethodMilk, entity.MethodIced)),
		// Sceneは許可enumを検証。
		validation.Field(&in.Scene, validation.In(entity.SceneWork, entity.SceneBreak, entity.SceneAfterMeal, entity.SceneRelax)),
		// TempPrefは許可enumを検証。
		validation.Field(&in.TempPref, validation.In(entity.TempHot, entity.TempIce)),
		// Noteは任意だが、入るなら1〜2000文字に制限。
		validation.Field(&in.Note,
			validation.When(in.Note != "", validation.RuneLength(1, 2000)),
		),
	)
}

// ユーザー発話追加入力を検証。
func (v *searchValidator) AddTurn(in usecase.AddTurnIn) error {
	return validation.ValidateStruct(&in,
		// SessionIDは1以上である必要がある。
		validation.Field(&in.SessionID, validation.Min(uint(1)).Error("session_id must be greater than 0")),
		// Bodyは必須で、1〜2000文字に制限。
		validation.Field(
			&in.Body,
			validation.Required.Error("body is required"),
			validation.RuneLength(1, 2000).Error("body must be 1 to 2000 chars"),
		),
	)
}

// 条件差分更新入力を検証。
func (v *searchValidator) PatchPref(in usecase.PatchPrefIn) error {
	// 何も更新対象がないPATCHを弾くため、まず変更有無を判定。
	changed := in.Flavor != nil ||
		in.Acidity != nil ||
		in.Bitterness != nil ||
		in.Body != nil ||
		in.Aroma != nil ||
		in.Mood != nil ||
		in.Method != nil ||
		in.Scene != nil ||
		in.TempPref != nil ||
		in.Excludes != nil ||
		in.Note != nil

	// 一つも変更フィールドが無ければ不正入力。
	if !changed {
		return validation.NewError("validation_no_fields", "at least one field is required")
	}

	return validation.ValidateStruct(&in,
		// SessionIDは必須で1以上。
		validation.Field(&in.SessionID, validation.Min(uint(1)).Error("session_id must be greater than 0")),
		// 各スコアは値がある場合だけ1〜5を検証。
		validation.Field(&in.Flavor, validation.When(in.Flavor != nil, validation.Min(1), validation.Max(5))),
		validation.Field(&in.Acidity, validation.When(in.Acidity != nil, validation.Min(1), validation.Max(5))),
		validation.Field(&in.Bitterness, validation.When(in.Bitterness != nil, validation.Min(1), validation.Max(5))),
		validation.Field(&in.Body, validation.When(in.Body != nil, validation.Min(1), validation.Max(5))),
		validation.Field(&in.Aroma, validation.When(in.Aroma != nil, validation.Min(1), validation.Max(5))),
		// Moodがある場合だけenumを検証。
		validation.Field(&in.Mood, validation.When(in.Mood != nil,
			validation.In(entity.MoodMorning, entity.MoodWork, entity.MoodRelax, entity.MoodNight))),
		// Methodがある場合だけenumを検証。
		validation.Field(&in.Method, validation.When(in.Method != nil,
			validation.In(entity.MethodDrip, entity.MethodEspresso, entity.MethodMilk, entity.MethodIced))),
		// Sceneがある場合だけenumを検証。
		validation.Field(&in.Scene, validation.When(in.Scene != nil,
			validation.In(entity.SceneWork, entity.SceneBreak, entity.SceneAfterMeal, entity.SceneRelax))),
		// TempPrefがある場合だけenumを検証。
		validation.Field(&in.TempPref, validation.When(in.TempPref != nil,
			validation.In(entity.TempHot, entity.TempIce))),
		// Noteがある場合だけ文字数を検証。
		validation.Field(&in.Note, validation.When(in.Note != nil && *in.Note != "",
			validation.RuneLength(1, 2000))),
	)
}

// セッション詳細取得入力を検証。
func (v *searchValidator) GetSession(in usecase.GetSessionIn) error {
	return validation.Validate(in.SessionID,
		validation.Min(uint(1)).Error("session_id must be greater than 0"),
	)
}

// 履歴一覧取得入力を検証。
func (v *searchValidator) ListHistory(in usecase.ListHistoryIn) error {
	return validation.ValidateStruct(&in,
		// limit/offsetの基本的なページング検証。
		validation.Field(&in.Limit, validation.Min(1), validation.Max(50)),
		validation.Field(&in.Offset, validation.Min(0)),
	)
}

// セッション終了入力を検証。
func (v *searchValidator) CloseSession(in usecase.CloseSessionIn) error {
	return validation.Validate(in.SessionID,
		validation.Min(uint(1)).Error("session_id must be greater than 0"),
	)
}
