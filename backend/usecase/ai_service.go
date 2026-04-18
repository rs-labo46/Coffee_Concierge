package usecase

import "coffee-spa/entity"

// ユーザー発話を条件差分へ変換するためのinterfaceで、Geminiはこのinterfaceの具体実装になる。
type PrefParser interface {
	ParseConditionDiff(in ParseConditionDiffIn) (ParseConditionDiffOut, error)
}

// 候補ごとの理由文を生成するためのinterface。
type ExplainSvc interface {
	BuildReasons(in BuildReasonsIn) ([]ReasonResult, bool, error)
}

// 次に聞くべき質問候補を生成するためのinterface。
type FollowupSvc interface {
	BuildQuestions(in BuildQuestionsIn) ([]string, bool, error)
}

// 発話から条件差分を作るための入力。
type ParseConditionDiffIn struct {
	// ユーザー発話本文。
	InputText string

	// 現在の構造化条件。
	Pref entity.Pref

	// 直近の会話履歴。
	Turns []entity.Turn
}

// Geminiが提案した条件差分で、ここで返った値をそのまま採用せず、usecase側で、必ずvalidatorに通す。
type ParseConditionDiffOut struct {
	Flavor     *int
	Acidity    *int
	Bitterness *int
	Body       *int
	Aroma      *int

	Mood     *entity.Mood
	Method   *entity.Method
	Scene    *entity.Scene
	TempPref *entity.TempPref

	Excludes []string
	Note     *string

	// フォールバックを使ったかどうか。
	FallbackUsed bool
}

// 候補理由文生成の入力。
type BuildReasonsIn struct {
	// 現在条件。
	Pref entity.Pref

	// 直近の会話履歴。
	Turns []entity.Turn

	// 理由文を生成したい候補群。
	Candidates []ReasonCandidate
}

// 理由文生成用にGeminiに渡す。
type ReasonCandidate struct {
	SuggestionID uint
	BeanName     string
	Roast        string
	Origin       string
	RecipeName   string
	Method       string
	ItemTitles   []string
}

// 候補1件分の理由文結果。
type ReasonResult struct {
	SuggestionID uint
	Reason       string
}

// 追加質問生成の入力。
type BuildQuestionsIn struct {
	// 現在の条件。
	Pref entity.Pref

	// 直近の会話履歴。
	Turns []entity.Turn

	// 現時点の候補群。
	Candidates []QuestionCandidate

	// 返す最大件数。
	Limit int
}

// 質問生成用にGeminiへ渡す候補要約。
type QuestionCandidate struct {
	BeanName   string
	Method     string
	ReasonHint string
}
