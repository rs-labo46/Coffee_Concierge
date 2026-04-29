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

// Geminiへ渡す公開済みBeanの軽量JSON。内部監査IDや不要な秘密情報は含めない。
type GeminiBeanCandidate struct {
	ID         uint         `json:"id"`
	Name       string       `json:"name"`
	Roast      entity.Roast `json:"roast"`
	Origin     string       `json:"origin"`
	Flavor     int          `json:"flavor"`
	Acidity    int          `json:"acidity"`
	Bitterness int          `json:"bitterness"`
	Body       int          `json:"body"`
	Aroma      int          `json:"aroma"`
	Desc       string       `json:"desc"`
}

// Geminiに登録済みBean候補から選定させる入力。
type GeminiBeanSelectionIn struct {
	InputText  string
	Pref       entity.Pref
	Turns      []entity.Turn
	Candidates []GeminiBeanCandidate
	Limit      int
}

// Geminiが選定したBean候補。usecase側でID存在・rank・score・reasonを必ず検証する。
type GeminiBeanSelection struct {
	BeanID uint
	Rank   int
	Score  int
	Reason string
}

// Geminiに登録済みBean候補から、選定・理由文・追加質問を1回で作らせる入力。
type GeminiSearchBundleIn struct {
	InputText  string
	Pref       entity.Pref
	Turns      []entity.Turn
	Candidates []GeminiBeanCandidate
	Limit      int
}

// Geminiの1回応答で得る検索補助結果。usecase側で必ず検証してから使う。
type GeminiSearchBundleOut struct {
	Selections []GeminiBeanSelection
	Followups  []string
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
	SelectBeans(in GeminiBeanSelectionIn) ([]GeminiBeanSelection, GeminiAuditMeta, error)
	BuildSearchBundle(in GeminiSearchBundleIn) (GeminiSearchBundleOut, GeminiAuditMeta, error)
	BuildReasons(in GeminiReasonIn) ([]GeminiReason, GeminiAuditMeta, error)
	BuildFollowups(in GeminiFollowupIn) ([]string, GeminiAuditMeta, error)
}
