package gemini

import (
	"strings"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

// GeminiClient のモック実装で、本物のGeminiAPIを呼ばず、簡単なキーワード判定で条件差分・理由文・追加質問を返す。
type MockClient struct{}

func NewMockClient() usecase.GeminiClient {
	return &MockClient{}
}

// 発話文から簡易的に条件差分候補を返す。
func (m *MockClient) BuildConditionDiff(in usecase.GeminiConditionDiffIn) (usecase.GeminiConditionDiffOut, error) {
	text := normalize(in.InputText)

	out := usecase.GeminiConditionDiffOut{}

	// body
	if containsAny(text, "軽め", "軽い", "すっきり") {
		v := clampScore(in.Pref.Body - 1)
		out.Body = &v
	}
	if containsAny(text, "重め", "重い", "コク", "濃い") {
		v := clampScore(in.Pref.Body + 1)
		out.Body = &v
	}

	// acidity
	if containsAny(text, "酸味弱め", "酸味は弱め", "酸味少なめ", "酸味控えめ") {
		v := clampScore(in.Pref.Acidity - 1)
		out.Acidity = &v
	}
	if containsAny(text, "酸味強め", "酸味ほしい", "酸っぱいのが好き") {
		v := clampScore(in.Pref.Acidity + 1)
		out.Acidity = &v
	}

	// bitterness
	if containsAny(text, "苦め", "ビター", "苦い") {
		v := clampScore(in.Pref.Bitterness + 1)
		out.Bitterness = &v
	}
	if containsAny(text, "苦味弱め", "苦味控えめ", "苦くない") {
		v := clampScore(in.Pref.Bitterness - 1)
		out.Bitterness = &v
	}

	// aroma
	if containsAny(text, "香り強め", "香り高い", "華やか") {
		v := clampScore(in.Pref.Aroma + 1)
		out.Aroma = &v
	}

	// method
	if containsAny(text, "ミルク", "ラテ", "カフェオレ") {
		method := entity.MethodMilk
		out.Method = &method
	}
	if containsAny(text, "ドリップ", "ハンドドリップ") {
		method := entity.MethodDrip
		out.Method = &method
	}
	if containsAny(text, "エスプレッソ") {
		method := entity.MethodEspresso
		out.Method = &method
	}
	if containsAny(text, "アイス") {
		method := entity.MethodIced
		out.Method = &method
	}

	// mood
	if containsAny(text, "朝", "朝向け", "朝に飲みたい") {
		mood := entity.MoodMorning
		out.Mood = &mood
	}
	if containsAny(text, "仕事中", "仕事", "集中") {
		mood := entity.MoodWork
		out.Mood = &mood
	}
	if containsAny(text, "夜", "夜向け") {
		mood := entity.MoodNight
		out.Mood = &mood
	}
	if containsAny(text, "リラックス", "落ち着きたい", "ゆっくり") {
		mood := entity.MoodRelax
		out.Mood = &mood
	}

	// scene
	if containsAny(text, "休憩", "一息") {
		scene := entity.SceneBreak
		out.Scene = &scene
	}
	if containsAny(text, "食後") {
		scene := entity.SceneAfterMeal
		out.Scene = &scene
	}
	if containsAny(text, "作業", "仕事") {
		scene := entity.SceneWork
		out.Scene = &scene
	}
	if containsAny(text, "くつろぎ", "リラックス") {
		scene := entity.SceneRelax
		out.Scene = &scene
	}

	// temp_pref
	if containsAny(text, "ホット", "温かい") {
		temp := entity.TempHot
		out.TempPref = &temp
	}
	if containsAny(text, "アイス", "冷たい") {
		temp := entity.TempIce
		out.TempPref = &temp
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
	}

	// note
	if trimmed := strings.TrimSpace(in.InputText); trimmed != "" {
		note := trimmed
		out.Note = &note
	}

	return out, nil
}

// suggestionごとの簡単な理由文を返す。
func (m *MockClient) BuildReasons(in usecase.GeminiReasonIn) ([]usecase.GeminiReason, error) {
	reasons := make([]usecase.GeminiReason, 0, len(in.Suggestions))

	for _, s := range in.Suggestions {
		beanName := findBeanNameByID(in.Beans, s.BeanID)
		reasonText := buildReasonText(beanName, in.Pref, in.ExplainLevel)

		reasons = append(reasons, usecase.GeminiReason{
			Rank:   s.Rank,
			Reason: reasonText,
		})
	}

	return reasons, nil
}

// 追加質問候補を返す。
func (m *MockClient) BuildFollowups(in usecase.GeminiFollowupIn) ([]string, error) {
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
		return out[:3], nil
	}

	return out, nil
}

func normalize(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func containsAny(text string, words ...string) bool {
	for _, word := range words {
		if strings.Contains(text, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

func clampScore(v int) int {
	if v < 1 {
		return 1
	}
	if v > 5 {
		return 5
	}
	return v
}

func findBeanNameByID(beans []entity.Bean, beanID uint) string {
	for _, bean := range beans {
		if bean.ID == beanID {
			return bean.Name
		}
	}
	return "この候補"
}

func buildReasonText(beanName string, pref entity.Pref, explainLevel string) string {
	if explainLevel == "easy" {
		return beanName + " は、今ほしい味に近く、飲みやすい候補です。"
	}

	parts := make([]string, 0, 3)

	if pref.Body >= 4 {
		parts = append(parts, "コクのある方向")
	} else if pref.Body > 0 {
		parts = append(parts, "重すぎない飲み口")
	}

	if pref.Acidity >= 4 {
		parts = append(parts, "明るい酸味")
	} else if pref.Acidity > 0 && pref.Acidity <= 2 {
		parts = append(parts, "酸味を抑えた方向")
	}

	if pref.Method == entity.MethodMilk {
		parts = append(parts, "ミルクと合わせやすい")
	}
	if pref.Method == entity.MethodDrip {
		parts = append(parts, "ドリップで香りを出しやすい")
	}
	if pref.Mood == entity.MoodMorning {
		parts = append(parts, "朝に合わせやすい")
	}
	if pref.Mood == entity.MoodRelax {
		parts = append(parts, "落ち着きたい場面に合いやすい")
	}

	if len(parts) == 0 {
		return beanName + " は、今の条件に近いバランスの候補です。"
	}

	return beanName + " は、" + strings.Join(parts, "、") + "候補です。"
}
