package gemini

import (
	"encoding/json"
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
Taste scale definitions:
- flavor: overall sweetness, pleasant flavor impression, balance, and memorability. 1 is weak/plain, 3 is balanced/standard, 5 is very expressive or sweet-impression rich.
- acidity: bright sourness, citrus, berry, and fruity sharpness. 1 is very low, 3 is moderate, 5 is very bright and acidic.
- bitterness: bitter impression, roast bitterness, and dark chocolate-like sharpness. 1 is very low, 3 is moderate, 5 is very bitter.
- body: weight, thickness, richness, and mouthfeel. 1 is very light, 3 is medium, 5 is heavy and rich.
- aroma: fragrance intensity and aromatic impression. 1 is quiet, 3 is moderate, 5 is very aromatic.

Rules:
- flavor/acidity/bitterness/body/aroma are integers 1..5 if present.
- mood is one of: morning, work, relax, night.
- method is one of: drip, espresso, milk, iced.
- scene is one of: work, break, after_meal, relax.
- temp_pref is one of: hot, ice.
- excludes is an array of strings.
- note is a short string.
- note must be written in Japanese if present.
- Omit keys you are not confident about.
- Interpret Japanese coffee preference expressions by meaning, not by exact keyword matching.
- Use the taste scale definitions to infer numeric values.
- If the user gives multiple conditions, preserve all non-conflicting conditions.
- If conditions conflict, prefer the latest explicit user input.
- If the user says “not too bitter”, “not too sour”, “not too heavy”, or similar, do not set the corresponding score to 5.
- Do not overwrite an existing preference unless the latest user input clearly changes it.
`

// 検索bundle生成用の固定ルール。Bean選定・理由文・追加質問を1回のGemini呼び出しで返す。
const searchBundleSystemInstruction = `
You are a search bundle generator for a coffee search engine.
Return JSON only.
Do not explain outside JSON.
Select only from the provided registered beans.
Never invent a bean, recipe, article, URL, ID, origin, or score.
Return up to the requested limit.

Output shape:
{
  "selections": [
    {"bean_id": 1, "rank": 1, "score": 92, "reason": "..."}
  ],
  "followup_questions": ["..."]
}

Taste scale definitions:
- flavor: overall sweetness, pleasant flavor impression, balance, and memorability. 1 is weak/plain, 3 is balanced/standard, 5 is very expressive or sweet-impression rich.
- acidity: bright sourness, citrus, berry, and fruity sharpness. 1 is very low, 3 is moderate, 5 is very bright and acidic.
- bitterness: bitter impression, roast bitterness, and dark chocolate-like sharpness. 1 is very low, 3 is moderate, 5 is very bitter.
- body: weight, thickness, richness, and mouthfeel. 1 is very light, 3 is medium, 5 is heavy and rich.
- aroma: fragrance intensity and aromatic impression. 1 is quiet, 3 is moderate, 5 is very aromatic.

Selection rules:
- Select from the provided registered beans only.
- Return exactly min(limit, number of provided candidates) selections whenever possible.
- If all provided candidates are valid beans, include every candidate up to the limit.
- Rank the candidates by how well they match the current pref and latest input.
- Score must be 0..100.
- Rank must start at 1 and be unique.
- Reason must be Japanese and grounded only in the provided bean JSON and current pref.
- Each reason must mention at least two concrete matching points from roast, origin, flavor, acidity, bitterness, body, aroma, method, mood, scene, or temp_pref.
- Do not use vague reason text like 「今の条件に近い候補です」 by itself.
- Do not reuse the same reason text across multiple selections.
- Each reason must explain why that specific bean was selected compared with the other provided candidates.
- Each reason must mention the bean name or a concrete attribute unique to that bean.
- If multiple candidates have similar scores, still describe their differences using origin, roast, flavor, acidity, bitterness, body, aroma, drinking scene, temperature, or brewing method.
- Avoid generic phrases that can apply to every candidate.

Followup rules:
- followup_questions must be Japanese.
- Return 0 to 3 questions.
- Write each question in polite, natural Japanese for general coffee drinkers.
- Keep each question to one sentence and short enough to display as a UI chip.
- Use simple words that beginners can understand.
- Avoid technical coffee jargon unless it is already provided by the user.
- Ask only questions that help narrow the current coffee preference.
- Prefer concrete preference questions about taste strength, acidity, bitterness, body, aroma, drinking scene, temperature, or brewing method.
- Do not ask broad or unrelated questions.
- Do not repeat a question that is already answered by the current pref or recent turns.
`

// 登録済みBean選定用の固定ルール。AIは渡された候補IDの中からだけ選ぶ。
const beanSelectionSystemInstruction = `
You are a bean selector for a coffee search engine.
Return JSON only.
Do not explain outside JSON.
Select only from the provided registered beans.
Never invent a bean, recipe, article, URL, ID, origin, or score.
Return up to the requested limit.

Taste scale definitions:
- flavor: overall sweetness, pleasant flavor impression, balance, and memorability. 1 is weak/plain, 3 is balanced/standard, 5 is very expressive or sweet-impression rich.
- acidity: bright sourness, citrus, berry, and fruity sharpness. 1 is very low, 3 is moderate, 5 is very bright and acidic.
- bitterness: bitter impression, roast bitterness, and dark chocolate-like sharpness. 1 is very low, 3 is moderate, 5 is very bitter.
- body: weight, thickness, richness, and mouthfeel. 1 is very light, 3 is medium, 5 is heavy and rich.
- aroma: fragrance intensity and aromatic impression. 1 is quiet, 3 is moderate, 5 is very aromatic.

Selection rules:
- Select from the provided registered beans only.
- Return exactly min(limit, number of provided candidates) selections whenever possible.
- If all provided candidates are valid beans, include every candidate up to the limit.
- Rank the candidates by how well they match the current pref and latest input.
- Score must be 0..100.
- Rank must start at 1 and be unique.
- Reason must be Japanese and grounded only in the provided bean JSON and current pref.
- Each reason must mention at least two concrete matching points from roast, origin, flavor, acidity, bitterness, body, aroma, method, mood, scene, or temp_pref.
- Do not use vague reason text like 「今の条件に近い候補です」 by itself.
- Do not reuse the same reason text across multiple selections.
- Each reason must explain why that specific bean was selected compared with the other provided candidates.
- Each reason must mention the bean name or a concrete attribute unique to that bean.
- If multiple candidates have similar scores, still describe their differences using origin, roast, flavor, acidity, bitterness, body, aroma, drinking scene, temperature, or brewing method.
- Avoid generic phrases that can apply to every candidate.
`

// 候補理由文生成用の固定ルール。
const reasonSystemInstruction = `
You generate short reasons for coffee recommendations.
Return JSON only.
Each candidate must have:
- suggestion_id
- reason

Language rules:
- Write every reason in Japanese.
- Do not use English unless it is a proper noun already provided, such as a bean name, origin, recipe name, or method value.
- Write in polite, natural Japanese for general coffee drinkers.
- Use simple words that beginners can understand.
- Avoid overly technical cupping terms unless they are already provided.
- Explain why the bean matches the user's preference, not just that it matches.
- Keep each reason to 1 or 2 sentences.
- Do not sound like an advertisement or a sales message.
- Each reason must be a complete Japanese sentence ending with "です。" or "ます。".

Rules:
- Reason must be grounded only in the provided candidate summary, pref, and recent turns.
- Do not invent beans, recipes, articles, or facts.
- Mention at least two concrete matching points from the candidate values, such as roast, origin, acidity, bitterness, body, aroma, method, or recipe.
- Avoid vague phrases like 「今の条件に近い候補です」 unless you also explain which condition matched.
- Keep each reason concise and readable.
`

// 追加質問生成用の固定ルール。
const followupSystemInstruction = `
You generate follow-up questions for a coffee search flow.
Return JSON only.
Output only the next questions needed to narrow the search.

Language rules:
- Write every question in Japanese.
- Do not use English unless it is a proper noun already provided.
- Write questions in polite, natural Japanese for general coffee drinkers.
- Keep each question to one sentence.
- Use simple words that beginners can understand.
- Avoid technical coffee jargon unless it is already provided by the user.
- Do not sound like an advertisement or a sales message.
- Ask in a helpful concierge tone, not a casual chatbot tone.

Rules:
- Return 0 to 3 questions.
- Ask only questions that help narrow the current coffee preference.
- Prefer concrete preference questions about taste strength, acidity, bitterness, body, aroma, drinking scene, temperature, or brewing method.
- Do not ask broad or unrelated questions.
- Do not repeat a question that is already answered by the current pref or recent turns.
- Keep questions short enough to display as UI chips.
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

// 検索bundle生成用のuser prompt。Bean候補JSON、現在条件、ユーザー入力を1回で渡す。
func buildSearchBundlePrompt(in usecase.GeminiSearchBundleIn) string {
	var b strings.Builder
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}

	b.WriteString("Return JSON only. Select registered beans and generate reasons and followup questions.\n")
	b.WriteString(fmt.Sprintf("Limit: %d\n", limit))
	b.WriteString("Current pref:\n")
	b.WriteString(fmt.Sprintf(
		`{"flavor":%d,"acidity":%d,"bitterness":%d,"body":%d,"aroma":%d,"mood":"%s","method":"%s","scene":"%s","temp_pref":"%s","note":"%s"}`+"\n",
		in.Pref.Flavor,
		in.Pref.Acidity,
		in.Pref.Bitterness,
		in.Pref.Body,
		in.Pref.Aroma,
		in.Pref.Mood,
		in.Pref.Method,
		in.Pref.Scene,
		in.Pref.TempPref,
		strings.ReplaceAll(in.Pref.Note, `"`, `'`),
	))

	b.WriteString("Recent turns:\n")
	for _, t := range in.Turns {
		b.WriteString(fmt.Sprintf("- role=%s body=%s\n", t.Role, t.Body))
	}

	b.WriteString("Latest user input:\n")
	b.WriteString(in.InputText)
	b.WriteString("\n")

	b.WriteString("Registered bean candidates JSON:\n")
	raw, err := json.Marshal(in.Candidates)
	if err != nil {
		b.WriteString("[]")
		return b.String()
	}
	b.Write(raw)
	return b.String()
}

// 登録済みBean候補選定用のuser prompt。
func buildBeanSelectionPrompt(in usecase.GeminiBeanSelectionIn) string {
	var b strings.Builder
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}

	b.WriteString("Return JSON only. Select registered beans from candidates.\n")
	b.WriteString(fmt.Sprintf("Limit: %d\n", limit))
	b.WriteString("Current pref:\n")
	b.WriteString(fmt.Sprintf(
		`{"flavor":%d,"acidity":%d,"bitterness":%d,"body":%d,"aroma":%d,"mood":"%s","method":"%s","scene":"%s","temp_pref":"%s","note":"%s"}`+"\n",
		in.Pref.Flavor,
		in.Pref.Acidity,
		in.Pref.Bitterness,
		in.Pref.Body,
		in.Pref.Aroma,
		in.Pref.Mood,
		in.Pref.Method,
		in.Pref.Scene,
		in.Pref.TempPref,
		strings.ReplaceAll(in.Pref.Note, `"`, `'`),
	))

	b.WriteString("Recent turns:\n")
	for _, t := range in.Turns {
		b.WriteString(fmt.Sprintf("- role=%s body=%s\n", t.Role, t.Body))
	}

	b.WriteString("Latest user input:\n")
	b.WriteString(in.InputText)
	b.WriteString("\n")

	b.WriteString("Registered bean candidates JSON:\n")
	raw, err := json.Marshal(in.Candidates)
	if err != nil {
		b.WriteString("[]")
		return b.String()
	}
	b.Write(raw)
	return b.String()
}

// 理由文生成用のuser prompt。
func buildReasonsPrompt(in usecase.BuildReasonsIn) string {
	var b strings.Builder
	b.WriteString("Output language: Japanese only. Return JSON only.\n")
	b.WriteString("Each reason must be one or two complete Japanese sentences ending with です。 or ます。\n")
	b.WriteString("Explain why the displayed bean matches the current pref, using roast/origin and at least two taste values from the same candidate line.\n")
	b.WriteString("Avoid vague phrases like 今の条件に近い候補です unless you explain which condition matched.\n")
	b.WriteString("Do not describe another bean. Do not invent facts not present in the candidate line.\n")
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
			`- suggestion_id=%d bean="%s" roast="%s" origin="%s" flavor=%d acidity=%d bitterness=%d body=%d aroma=%d recipe="%s" method="%s" items="%s"`+"\n",
			c.SuggestionID,
			c.BeanName,
			c.Roast,
			c.Origin,
			c.Flavor,
			c.Acidity,
			c.Bitterness,
			c.Body,
			c.Aroma,
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
	b.WriteString("Output language: Japanese only. Return JSON only.\n")
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
