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
- note must be written in Japanese if present.
- Omit keys you are not confident about.
- Interpret common Japanese coffee preference expressions as condition hints.
- If the user asks for "苦め", "苦い", "ビター", "苦味強め", "苦味しっかり", "しっかり苦味", "苦味が欲しい", "苦味のある", "苦味を感じる", "パンチがある", "キリッと苦い", "深煎りっぽい苦さ", "大人っぽい苦味", or "甘くない苦味", set bitterness to 5.
- If the user asks for "少し苦め", "やや苦め", "ほんのり苦い", "苦味も少し欲しい", or "苦味は中くらい", set bitterness to 4.
- If the user asks for "苦味控えめ", "苦味弱め", "苦くない", "苦味少なめ", "苦味は少なく", "苦味を抑えたい", "苦すぎない", "苦いのは苦手", "苦味が苦手", "まろやか", "やさしい苦味", or "飲みやすい苦味", set bitterness to 1 or 2.
- If the user asks for "酸味強め", "酸っぱい", "明るい酸味", "フルーティーな酸味", "爽やかな酸味", "華やかな酸味", "柑橘っぽい", "ベリーっぽい", "果実感", "フルーティー", or "明るい味", set acidity to 5.
- If the user asks for "少し酸味", "酸味も少し", "酸味は中くらい", "ほどよい酸味", or "バランスのある酸味", set acidity to 3 or 4.
- If the user asks for "酸味弱め", "酸味控えめ", "酸味少なめ", "酸っぱくない", "酸っぱいのは苦手", "酸味が苦手", "酸味を抑えたい", "酸味なし", or "まろやかな味", set acidity to 1 or 2.
- If the user asks for "軽め", "すっきり", "さっぱり", "あっさり", "軽やか", "さらっと", "ごくごく飲める", "重くない", "飲み疲れしない", "朝に軽く", "透明感", or "クリーン", set body to 2.
- If the user asks for "少しコク", "ほどよいコク", "中くらいのコク", "バランスのよいコク", or "飲みごたえも少し", set body to 3 or 4.
- If the user asks for "重め", "どっしり", "深いコク", "濃厚", "しっかりしたコク", "飲みごたえ", "厚みがある", "力強い", "濃いめ", "深み", "余韻が長い", "リッチ", or "こってり", set body to 5.
- If the user asks for "香り高い", "香りが強い", "香り重視", "華やか", "アロマ強め", "いい香り", "香ばしい香り", "フローラル", "花っぽい", "ナッツの香り", "チョコっぽい香り", or "香りを楽しみたい", set aroma to 5.
- If the user asks for "香り控えめ", "香りは弱め", "香りは普通", or "香りより味重視", set aroma to 2 or 3.
- If the user asks for "甘み", "甘い余韻", "甘さ", "ナッツ感", "チョコ感", "キャラメル感", "まろやかな甘み", "丸い味", "バランスがいい", or "飲みやすい", set flavor to 4 or 5.
- If the user asks for "個性的", "特徴が強い", "クセがある", "変わった味", "印象に残る", or "複雑な味", set flavor to 5.
- If the user asks for "普通", "無難", "クセが少ない", "シンプル", or "毎日飲める", set flavor to 3.
- If the user asks for "朝", "朝に飲みたい", "起きた時", "目覚め", "眠気覚まし", "出勤前", or "朝食と一緒", set mood to "morning".
- If the user asks for "仕事", "作業", "集中", "勉強", "デスクワーク", "仕事中", "作業用", "長時間飲む", or "PC作業", set mood to "work" and scene to "work".
- If the user asks for "休憩", "一息", "気分転換", "昼休み", "午後の休憩", "リフレッシュ", or "ブレイク", set scene to "break".
- If the user asks for "リラックス", "くつろぎ", "ゆっくり", "休日", "落ち着きたい", "癒されたい", "寝る前以外の夜", or "のんびり", set mood to "relax" and scene to "relax".
- If the user asks for "夜", "夜に飲みたい", "寝る前", "夕食後", "夜の作業", or "夜向け", set mood to "night".
- If the user asks for "食後", "ご飯の後", "食事の後", "デザートと一緒", "甘いものと", "スイーツと", or "ケーキと", set scene to "after_meal".
- If the user asks for "ホット", "温かい", "あたたかい", "熱い", "冬", "寒い日", or "温まりたい", set temp_pref to "hot".
- If the user asks for "アイス", "冷たい", "冷やして", "暑い日", "夏", "さっぱり冷たく", "氷", or "アイスコーヒー", set method to "iced" and temp_pref to "ice".
- If the user asks for "ドリップ", "ハンドドリップ", "ペーパー", "家で淹れる", "ゆっくり淹れたい", or "ブラックで飲む", set method to "drip".
- If the user asks for "エスプレッソ", "濃縮", "ラテのベース", "カフェラテ用", or "短時間で濃い", set method to "espresso".
- If the user asks for "ミルク", "牛乳", "カフェラテ", "ラテ", "オレ", "ミルクに合う", "ミルク割り", "豆乳", "オーツミルク", or "ミルク向け", set method to "milk".
- If the user asks for both "軽め" and "苦め", set bitterness to 5 and body to 2. Do not ignore either condition.
- If the user asks for both "酸味弱め" and "フルーティー", prefer acidity 2 or 3 and flavor 4. Do not set acidity to 5 unless the user explicitly asks for strong acidity.
- If the user asks for both "ミルク" and "苦め", set method to "milk" and bitterness to 5.
- If the user asks for both "朝" and "すっきり", set mood to "morning" and body to 2.
- If the user asks for both "夜" and "軽め", set mood to "night" and body to 2.
- If the user says "苦すぎない", "酸っぱすぎない", "重すぎない", or similar "too much" expressions, avoid setting the corresponding score to 5.
- Do not overwrite an existing preference unless the latest user input clearly changes it.
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
- Keep the Japanese natural, concise, and suitable for a consumer coffee app.
- Each reason must be a complete Japanese sentence ending with "です。" or "ます。".

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

Language rules:
- Write every question in Japanese.
- Do not use English unless it is a proper noun already provided.
- Use short, natural Japanese suitable for a consumer coffee app.

Rules:
- Do not ask broad or unrelated questions.
- Keep questions short.
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
