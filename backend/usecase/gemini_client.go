package usecase

import "coffee-spa/entity"

// AI呼び出しの監査に使うメタ情報。
type GeminiAuditMeta struct {
	Provider   string
	Model      string
	Status     string
	DurationMS int64
	ErrorType  string
}

// 発話から条件差分候補を作る
type GeminiConditionDiffIn struct {
	InputText string
	Pref      entity.Pref
	Turns     []entity.Turn
}

// Geminiが提案した条件差分の候補=>validator/usecase側で必ず検証
type GeminiConditionDiffOut struct {
	Flavor     *int
	Acidity    *int
	Bitterness *int
	Body       *int
	Aroma      *int
	Mood       *entity.Mood
	Method     *entity.Method
	Scene      *entity.Scene
	TempPref   *entity.TempPref
	Excludes   []string
	Note       *string
}

// suggestion向け。理由文生成の入力。
type GeminiReasonIn struct {
	InputText    string
	Pref         entity.Pref
	Suggestions  []entity.Suggestion
	Beans        []entity.Bean
	Recipes      []entity.Recipe
	Items        []entity.Item
	ExplainLevel string
}

// suggestionごとの理由の文。
type GeminiReason struct {
	Rank   int
	Reason string
}

// 追加質問の生成入力。
type GeminiFollowupIn struct {
	InputText string
	Pref      entity.Pref
	Beans     []entity.Bean
}

// geminiの呼び出し。iinterfaceにのみ依存しHTTPには依存しない
type GeminiClient interface {
	Info() (provider string, model string)
	BuildConditionDiff(in GeminiConditionDiffIn) (GeminiConditionDiffOut, GeminiAuditMeta, error)
	BuildReasons(in GeminiReasonIn) ([]GeminiReason, GeminiAuditMeta, error)
	BuildFollowups(in GeminiFollowupIn) ([]string, GeminiAuditMeta, error)
}
