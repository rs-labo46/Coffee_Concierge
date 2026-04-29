package gemini

import (
	"errors"
	"strings"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

// 実APIを叩かずに、発話から簡易的な条件差分・理由文・追加質問を返す。
type MockClient struct{}

func NewMockClient() usecase.GeminiClient {
	return &MockClient{}
}

var ErrEmptyInput = errors.New("empty input")

// 発話文から簡易的な条件差分候補を返す。
// 単純なキーワード一致で値を組み立てる。
func (m *MockClient) BuildConditionDiff(
	in usecase.GeminiConditionDiffIn,
) (usecase.GeminiConditionDiffOut, usecase.GeminiAuditMeta, error) {
	text := normalize(in.InputText)
	if text == "" {
		return usecase.GeminiConditionDiffOut{}, usecase.GeminiAuditMeta{
			Provider:   "gemini_mock",
			Model:      "mock",
			Status:     "error",
			DurationMS: 0,
			ErrorType:  "invalid_input",
		}, ErrEmptyInput
	}

	out := usecase.GeminiConditionDiffOut{}
	matched := false

	// body
	switch {
	case containsAny(text, "軽め", "軽い", "すっきり"):
		v := clampScore(in.Pref.Body - 1)
		out.Body = &v
		matched = true
	case containsAny(text, "重め", "重い", "コク", "濃い"):
		v := clampScore(in.Pref.Body + 1)
		out.Body = &v
		matched = true
	}

	// acidity
	switch {
	case containsAny(text, "酸味弱め", "酸味は弱め", "酸味少なめ", "酸味控えめ"):
		v := clampScore(in.Pref.Acidity - 1)
		out.Acidity = &v
		matched = true
	case containsAny(text, "酸味強め", "酸味ほしい", "酸っぱいのが好き"):
		v := clampScore(in.Pref.Acidity + 1)
		out.Acidity = &v
		matched = true
	}

	// bitterness
	switch {
	case containsAny(text, "苦味弱め", "苦味控えめ", "苦くない"):
		v := clampScore(in.Pref.Bitterness - 1)
		out.Bitterness = &v
		matched = true
	case containsAny(text, "苦め", "ビター", "苦い"):
		v := clampScore(in.Pref.Bitterness + 1)
		out.Bitterness = &v
		matched = true
	}

	// aroma
	switch {
	case containsAny(text, "香り強め", "香り高い", "華やか"):
		v := clampScore(in.Pref.Aroma + 1)
		out.Aroma = &v
		matched = true
	}

	// method
	switch {
	case containsAny(text, "ミルク", "ラテ", "カフェオレ"):
		method := entity.MethodMilk
		out.Method = &method
		matched = true
	case containsAny(text, "ドリップ", "ハンドドリップ"):
		method := entity.MethodDrip
		out.Method = &method
		matched = true
	case containsAny(text, "エスプレッソ"):
		method := entity.MethodEspresso
		out.Method = &method
		matched = true
	case containsAny(text, "アイス"):
		method := entity.MethodIced
		out.Method = &method
		matched = true
	}

	// mood
	switch {
	case containsAny(text, "朝", "朝向け", "朝に飲みたい"):
		mood := entity.MoodMorning
		out.Mood = &mood
		matched = true
	case containsAny(text, "仕事中", "仕事", "集中"):
		mood := entity.MoodWork
		out.Mood = &mood
		matched = true
	case containsAny(text, "夜", "夜向け"):
		mood := entity.MoodNight
		out.Mood = &mood
		matched = true
	case containsAny(text, "リラックス", "落ち着きたい", "ゆっくり"):
		mood := entity.MoodRelax
		out.Mood = &mood
		matched = true
	}

	// scene
	switch {
	case containsAny(text, "休憩", "一息"):
		scene := entity.SceneBreak
		out.Scene = &scene
		matched = true
	case containsAny(text, "食後"):
		scene := entity.SceneAfterMeal
		out.Scene = &scene
		matched = true
	case containsAny(text, "作業", "仕事"):
		scene := entity.SceneWork
		out.Scene = &scene
		matched = true
	case containsAny(text, "くつろぎ", "リラックス"):
		scene := entity.SceneRelax
		out.Scene = &scene
		matched = true
	}

	// temp_pref
	switch {
	case containsAny(text, "ホット", "温かい"):
		temp := entity.TempHot
		out.TempPref = &temp
		matched = true
	case containsAny(text, "アイス", "冷たい"):
		temp := entity.TempIce
		out.TempPref = &temp
		matched = true
	}

	// excludes
	excludes := make([]string, 0, 2)
	if containsAny(text, "酸味なし", "酸味いらない") {
		excludes = append(excludes, "acidic")
	}
	if containsAny(text, "深煎り以外", "深煎りは避けたい") {
		excludes = append(excludes, "dark_roast")
	}
	if len(excludes) > 0 {
		out.Excludes = excludes
		matched = true
	}

	// 元の発話は補助メモとして保持。
	if trimmed := strings.TrimSpace(in.InputText); trimmed != "" {
		note := trimmed
		out.Note = &note
	}
	// errorにはせずno-op successで返す。
	if !matched {
		return out, usecase.GeminiAuditMeta{
			Provider:   "gemini_mock",
			Model:      "mock",
			Status:     "success",
			DurationMS: 0,
			ErrorType:  "",
		}, nil
	}

	return out, usecase.GeminiAuditMeta{
		Provider:   "gemini_mock",
		Model:      "mock",
		Status:     "success",
		DurationMS: 0,
		ErrorType:  "",
	}, nil
}

// 検索bundleをモックで返す。Bean選定、理由文、追加質問を1回で返す。
func (m *MockClient) BuildSearchBundle(
	in usecase.GeminiSearchBundleIn,
) (usecase.GeminiSearchBundleOut, usecase.GeminiAuditMeta, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}

	selections := make([]usecase.GeminiBeanSelection, 0, limit)
	for _, bean := range in.Candidates {
		if bean.ID == 0 {
			continue
		}
		score := 100 - absInt(in.Pref.Flavor-bean.Flavor)*4 - absInt(in.Pref.Acidity-bean.Acidity)*4 - absInt(in.Pref.Bitterness-bean.Bitterness)*4 - absInt(in.Pref.Body-bean.Body)*4 - absInt(in.Pref.Aroma-bean.Aroma)*4
		if score < 0 {
			score = 0
		}
		selections = append(selections, usecase.GeminiBeanSelection{
			BeanID: bean.ID,
			Rank:   len(selections) + 1,
			Score:  score,
			Reason: bean.Name + " は、" + string(bean.Roast) + "の焙煎と味覚バランスが現在の条件に合いやすい候補です。",
		})
		if len(selections) >= limit {
			break
		}
	}

	return usecase.GeminiSearchBundleOut{
			Selections: selections,
			Followups: []string{
				"酸味はもう少し控えめが好みですか？",
				"ホットとアイスならどちらで飲みたいですか？",
			},
		}, usecase.GeminiAuditMeta{
			Provider:   "gemini_mock",
			Model:      "mock",
			Status:     "success",
			DurationMS: 0,
			ErrorType:  "",
		}, nil
}

// 登録済みBean候補から、モックでは入力順に最大Limit件を返す。
func (m *MockClient) SelectBeans(
	in usecase.GeminiBeanSelectionIn,
) ([]usecase.GeminiBeanSelection, usecase.GeminiAuditMeta, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}

	out := make([]usecase.GeminiBeanSelection, 0, limit)
	for _, bean := range in.Candidates {
		if bean.ID == 0 {
			continue
		}
		score := 100 - absInt(in.Pref.Flavor-bean.Flavor)*4 - absInt(in.Pref.Acidity-bean.Acidity)*4 - absInt(in.Pref.Bitterness-bean.Bitterness)*4 - absInt(in.Pref.Body-bean.Body)*4 - absInt(in.Pref.Aroma-bean.Aroma)*4
		if score < 0 {
			score = 0
		}
		out = append(out, usecase.GeminiBeanSelection{
			BeanID: bean.ID,
			Rank:   len(out) + 1,
			Score:  score,
			Reason: bean.Name + " は、登録済み豆の中で現在の条件に近い候補です。",
		})
		if len(out) >= limit {
			break
		}
	}

	return out, usecase.GeminiAuditMeta{
		Provider:   "gemini_mock",
		Model:      "mock",
		Status:     "success",
		DurationMS: 0,
		ErrorType:  "",
	}, nil
}

// suggestionごとの簡単な理由文を返す。
// ExplainLevelに応じて少し文体を変える。
func (m *MockClient) BuildReasons(
	in usecase.GeminiReasonIn,
) ([]usecase.GeminiReason, usecase.GeminiAuditMeta, error) {
	reasons := make([]usecase.GeminiReason, 0, len(in.Suggestions))

	for _, s := range in.Suggestions {
		beanName := findBeanNameByID(in.Beans, s.BeanID)
		reasonText := buildReasonText(beanName, in.Pref, in.ExplainLevel)

		reasons = append(reasons, usecase.GeminiReason{
			Rank:   s.Rank,
			Reason: reasonText,
		})
	}

	return reasons, usecase.GeminiAuditMeta{
		Provider:   "gemini_mock",
		Model:      "mock",
		Status:     "success",
		DurationMS: 0,
		ErrorType:  "",
	}, nil
}

// 追加質問候補を最大3件返す。
// 今のPrefで未確定な観点を優先して質問化する。
func (m *MockClient) BuildFollowups(
	in usecase.GeminiFollowupIn,
) ([]string, usecase.GeminiAuditMeta, error) {
	out := make([]string, 0, 3)

	if in.Pref.Method == "" {
		out = append(out, "ドリップ、ミルク、アイスのどれで飲みたいですか？")
	}
	if in.Pref.Mood == "" {
		out = append(out, "朝向け、仕事中向け、リラックス向けのどれに近いですか？")
	}
	if in.Pref.Scene == "" {
		out = append(out, "仕事中、休憩、食後、リラックスのどの場面で飲みたいですか？")
	}
	if len(out) < 3 {
		out = append(out, "もう少し軽め、酸味弱め、ミルク向けなどの希望はありますか？")
	}
	if len(out) > 3 {
		out = out[:3]
	}

	return out, usecase.GeminiAuditMeta{
		Provider:   "gemini_mock",
		Model:      "mock",
		Status:     "success",
		DurationMS: 0,
		ErrorType:  "",
	}, nil
}

// 比較しやすいように文字列を小文字・前後trimに寄せる。
func normalize(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

// 文字列が候補語のどれかを含むかを返す。
func containsAny(text string, words ...string) bool {
	for _, word := range words {
		if strings.Contains(text, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// 1~5の範囲に丸める。
func clampScore(v int) int {
	if v < 1 {
		return 1
	}
	if v > 5 {
		return 5
	}
	return v
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// beanIDに対応するBean名を返す。見つからない場合は空文字を返す。
func findBeanNameByID(beans []entity.Bean, beanID uint) string {
	for _, bean := range beans {
		if bean.ID == beanID {
			return bean.Name
		}
	}
	return ""
}

// モック用の簡単な理由文を組み立てる。easy の場合は、よりやさしい表現にする。
func buildReasonText(
	beanName string,
	pref entity.Pref,
	explainLevel string,
) string {
	name := beanName
	if strings.TrimSpace(name) == "" {
		name = "この豆"
	}

	if explainLevel == "easy" {
		return name + " は、今の好みに近くて飲みやすそうです。"
	}

	parts := make([]string, 0, 3)

	if pref.Body >= 4 {
		parts = append(parts, "コク寄り")
	} else if pref.Body <= 2 {
		parts = append(parts, "軽め")
	}

	if pref.Acidity >= 4 {
		parts = append(parts, "酸味寄り")
	} else if pref.Acidity <= 2 {
		parts = append(parts, "酸味控えめ")
	}

	if pref.Bitterness >= 4 {
		parts = append(parts, "苦味寄り")
	} else if pref.Bitterness <= 2 {
		parts = append(parts, "苦味控えめ")
	}

	if len(parts) == 0 {
		return name + " は、今の条件に近い候補です。"
	}

	return name + " は、" + strings.Join(parts, "で、") + "の好みに合いやすい候補です。"
}

// 監査用にprovider/modelを返す。
func (m *MockClient) Info() (string, string) {
	return "gemini_mock", "mock"
}
