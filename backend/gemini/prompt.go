package gemini

import (
	"fmt"
	"strings"

	"coffee-spa/usecase"
)

// 条件差分生成用の固定ルール。AIは提案だけを返し、最終反映は backend 側が行う前提。
const conditionDiffSystemInstruction = `
You are a condition-diff generator for a coffee search engine.
Return JSON only.
Do not explain.
Do not add keys not requested.
Allowed keys:
flavor, acidity, bitterness, body, aroma,
mood, method, scene, temp_pref, excludes, note.

Rules:
- flavor/acidity/bitterness/body/aroma are integers 1..5 if present.
- mood is one of: morning, work, relax, night.
- method is one of: drip, espresso, milk, iced.
- scene is one of: work, break, after_meal, relax.
- temp_pref is one of: hot, ice.
- excludes is an array of strings.
- note is a short string.
- Omit keys you are not confident about.
`

// 候補理由文生成用の固定ルール。
const reasonSystemInstruction = `
You generate short reasons for coffee recommendations.
Return JSON only.
Each candidate must have:
- suggestion_id
- reason

Rules:
- Reason must be grounded only in the provided candidate summary, pref, and recent turns.
- Do not invent beans, recipes, articles, or facts.
- Keep each reason concise and readable.
`

// 追加質問生成用の固定ルール。
const followupSystemInstruction = `
You generate follow-up questions for a coffee search flow.
Return JSON only.
Output only the next questions needed to narrow the search.
Do not ask broad or unrelated questions.
Keep questions short.
`

// 条件差分生成用のuser prompt。
func buildConditionDiffPrompt(in usecase.ParseConditionDiffIn) string {
	var b strings.Builder

	b.WriteString("Current pref:\n")
	b.WriteString(fmt.Sprintf(
		`{"flavor":%d,"acidity":%d,"bitterness":%d,"body":%d,"aroma":%d,"mood":"%s","method":"%s","scene":"%s","temp_pref":"%s","excludes_count":%d}`+"\n",
		in.Pref.Flavor,
		in.Pref.Acidity,
		in.Pref.Bitterness,
		in.Pref.Body,
		in.Pref.Aroma,
		in.Pref.Mood,
		in.Pref.Method,
		in.Pref.Scene,
		in.Pref.TempPref,
		len(in.Pref.Excludes),
	))

	b.WriteString("Recent turns:\n")
	for _, t := range in.Turns {
		b.WriteString(fmt.Sprintf("- role=%s body=%s\n", t.Role, t.Body))
	}

	b.WriteString("Latest user input:\n")
	b.WriteString(in.InputText)

	return b.String()
}

// 理由文生成用のuser prompt。
func buildReasonsPrompt(in usecase.BuildReasonsIn) string {
	var b strings.Builder

	b.WriteString("Current pref:\n")
	b.WriteString(fmt.Sprintf(
		`{"flavor":%d,"acidity":%d,"bitterness":%d,"body":%d,"aroma":%d,"mood":"%s","method":"%s","scene":"%s","temp_pref":"%s"}`+"\n",
		in.Pref.Flavor,
		in.Pref.Acidity,
		in.Pref.Bitterness,
		in.Pref.Body,
		in.Pref.Aroma,
		in.Pref.Mood,
		in.Pref.Method,
		in.Pref.Scene,
		in.Pref.TempPref,
	))

	b.WriteString("Recent turns:\n")
	for _, t := range in.Turns {
		b.WriteString(fmt.Sprintf("- role=%s body=%s\n", t.Role, t.Body))
	}

	b.WriteString("Candidates:\n")
	for _, c := range in.Candidates {
		b.WriteString(fmt.Sprintf(
			`- suggestion_id=%d bean="%s" roast="%s" origin="%s" recipe="%s" method="%s" items="%s"`+"\n",
			c.SuggestionID,
			c.BeanName,
			c.Roast,
			c.Origin,
			c.RecipeName,
			c.Method,
			strings.Join(c.ItemTitles, ", "),
		))
	}

	return b.String()
}

// 追加質問生成用のuser prompt。
func buildFollowupPrompt(in usecase.BuildQuestionsIn) string {
	var b strings.Builder

	b.WriteString("Current pref:\n")
	b.WriteString(fmt.Sprintf(
		`{"flavor":%d,"acidity":%d,"bitterness":%d,"body":%d,"aroma":%d,"mood":"%s","method":"%s","scene":"%s","temp_pref":"%s"}`+"\n",
		in.Pref.Flavor,
		in.Pref.Acidity,
		in.Pref.Bitterness,
		in.Pref.Body,
		in.Pref.Aroma,
		in.Pref.Mood,
		in.Pref.Method,
		in.Pref.Scene,
		in.Pref.TempPref,
	))

	b.WriteString("Recent turns:\n")
	for _, t := range in.Turns {
		b.WriteString(fmt.Sprintf("- role=%s body=%s\n", t.Role, t.Body))
	}

	b.WriteString("Candidates:\n")
	for _, c := range in.Candidates {
		b.WriteString(fmt.Sprintf(
			`- bean="%s" method="%s" hint="%s"`+"\n",
			c.BeanName,
			c.Method,
			c.ReasonHint,
		))
	}

	b.WriteString(fmt.Sprintf("Limit: %d\n", in.Limit))

	return b.String()
}
