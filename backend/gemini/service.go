package gemini

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

//	Gemini REST APIを叩く本実装。
//
// usecase.GeminiClientを満たし、usecaseからSDK/RESTの詳細を隠す。
type Service struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// envからGemini設定を読み込んで生成する。
// モデル種別はenvから受け、コードへ直書きしない。
func NewService(apiKey string, model string) (*Service, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("GEMINI_API_KEY is empty")
	}

	if strings.TrimSpace(model) == "" {
		return nil, errors.New("GEMINI_MODEL is empty")
	}

	return &Service{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// 発話から条件差分候補を生成する。usecase側のSuggestion/Bean/Recipe/Itemを、理由生成用の軽い要約へ詰め替える。
func (s *Service) BuildConditionDiff(
	in usecase.GeminiConditionDiffIn,
) (usecase.GeminiConditionDiffOut, usecase.GeminiAuditMeta, error) {
	start := time.Now()

	prompt := buildConditionDiffPrompt(usecase.ParseConditionDiffIn{
		InputText: in.InputText,
		Pref:      in.Pref,
		Turns:     in.Turns,
	})

	var out conditionDiffResponse
	if err := s.generateJSON(
		conditionDiffSystemInstruction,
		prompt,
		conditionDiffSchema,
		&out,
	); err != nil {
		return usecase.GeminiConditionDiffOut{}, usecase.GeminiAuditMeta{
			Provider:   "gemini",
			Model:      s.model,
			Status:     "failed",
			DurationMS: time.Since(start).Milliseconds(),
			ErrorType:  "generate_condition_diff_failed",
		}, err
	}

	return usecase.GeminiConditionDiffOut{
			Flavor:     out.Flavor,
			Acidity:    out.Acidity,
			Bitterness: out.Bitterness,
			Body:       out.Body,
			Aroma:      out.Aroma,
			Mood:       out.Mood,
			Method:     out.Method,
			Scene:      out.Scene,
			TempPref:   out.TempPref,
			Excludes:   out.Excludes,
			Note:       out.Note,
		}, usecase.GeminiAuditMeta{
			Provider:   "gemini",
			Model:      s.model,
			Status:     "success",
			DurationMS: time.Since(start).Milliseconds(),
			ErrorType:  "",
		}, nil
}

// 検索1回分のAI補助を1リクエストで生成する。
// Bean選定、理由文、追加質問をまとめて返し、無料枠の消費を抑える。
func (s *Service) BuildSearchBundle(
	in usecase.GeminiSearchBundleIn,
) (usecase.GeminiSearchBundleOut, usecase.GeminiAuditMeta, error) {
	start := time.Now()

	prompt := buildSearchBundlePrompt(in)

	var out searchBundleResponse
	if err := s.generateJSON(
		searchBundleSystemInstruction,
		prompt,
		searchBundleSchema,
		&out,
	); err != nil {
		return usecase.GeminiSearchBundleOut{}, usecase.GeminiAuditMeta{
			Provider:   "gemini",
			Model:      s.model,
			Status:     "failed",
			DurationMS: time.Since(start).Milliseconds(),
			ErrorType:  "build_search_bundle_failed",
		}, err
	}

	selections := make([]usecase.GeminiBeanSelection, 0, len(out.Selections))
	for _, selection := range out.Selections {
		reason := strings.TrimSpace(selection.Reason)
		if selection.BeanID == 0 || selection.Rank <= 0 || selection.Score < 0 || selection.Score > 100 || reason == "" {
			continue
		}
		selections = append(selections, usecase.GeminiBeanSelection{
			BeanID: selection.BeanID,
			Rank:   selection.Rank,
			Score:  selection.Score,
			Reason: reason,
		})
	}

	followups := make([]string, 0, 3)
	for _, q := range out.FollowupQuestions {
		v := strings.TrimSpace(q)
		if v == "" {
			continue
		}
		followups = append(followups, v)
		if len(followups) >= 3 {
			break
		}
	}

	return usecase.GeminiSearchBundleOut{
			Selections: selections,
			Followups:  followups,
		}, usecase.GeminiAuditMeta{
			Provider:   "gemini",
			Model:      s.model,
			Status:     "success",
			DurationMS: time.Since(start).Milliseconds(),
			ErrorType:  "",
		}, nil
}

// 登録済みBean JSONの中から、Geminiに最大10件を選定させる。
func (s *Service) SelectBeans(
	in usecase.GeminiBeanSelectionIn,
) ([]usecase.GeminiBeanSelection, usecase.GeminiAuditMeta, error) {
	start := time.Now()

	prompt := buildBeanSelectionPrompt(in)

	var out beanSelectionListResponse
	if err := s.generateJSON(
		beanSelectionSystemInstruction,
		prompt,
		beanSelectionListSchema,
		&out,
	); err != nil {
		return nil, usecase.GeminiAuditMeta{
			Provider:   "gemini",
			Model:      s.model,
			Status:     "failed",
			DurationMS: time.Since(start).Milliseconds(),
			ErrorType:  "select_beans_failed",
		}, err
	}

	results := make([]usecase.GeminiBeanSelection, 0, len(out.Selections))
	for _, selection := range out.Selections {
		if selection.BeanID == 0 || selection.Rank <= 0 || selection.Score < 0 || selection.Score > 100 {
			continue
		}
		if strings.TrimSpace(selection.Reason) == "" {
			continue
		}
		results = append(results, usecase.GeminiBeanSelection{
			BeanID: selection.BeanID,
			Rank:   selection.Rank,
			Score:  selection.Score,
			Reason: strings.TrimSpace(selection.Reason),
		})
	}

	return results, usecase.GeminiAuditMeta{
		Provider:   "gemini",
		Model:      s.model,
		Status:     "success",
		DurationMS: time.Since(start).Milliseconds(),
		ErrorType:  "",
	}, nil
}

// suggestionごとの理由文を生成。
// usecase.GeminiReasonInをprompt用の候補要約へ変換して送る。
func (s *Service) BuildReasons(
	in usecase.GeminiReasonIn,
) ([]usecase.GeminiReason, usecase.GeminiAuditMeta, error) {
	start := time.Now()

	candidates := make([]usecase.ReasonCandidate, 0, len(in.Suggestions))

	for _, sug := range in.Suggestions {
		bean := findBeanByID(in.Beans, sug.BeanID)
		recipe := findRecipeByID(in.Recipes, sug.RecipeID)
		itemTitles := findRelatedItemTitles(in.Items, sug.ItemID)

		beanName := ""
		roast := ""
		origin := ""
		flavor := 0
		acidity := 0
		bitterness := 0
		body := 0
		aroma := 0

		if bean != nil {
			beanName = bean.Name
			roast = string(bean.Roast)
			origin = bean.Origin
			flavor = bean.Flavor
			acidity = bean.Acidity
			bitterness = bean.Bitterness
			body = bean.Body
			aroma = bean.Aroma
		}

		recipeName := ""
		method := ""
		if recipe != nil {
			recipeName = recipe.Name
			method = string(recipe.Method)
		}

		candidates = append(candidates, usecase.ReasonCandidate{
			SuggestionID: uint(sug.Rank),
			BeanName:     beanName,
			Roast:        roast,
			Origin:       origin,
			Flavor:       flavor,
			Acidity:      acidity,
			Bitterness:   bitterness,
			Body:         body,
			Aroma:        aroma,
			RecipeName:   recipeName,
			Method:       method,
			ItemTitles:   itemTitles,
		})
	}

	prompt := buildReasonsPrompt(usecase.BuildReasonsIn{
		Pref:       in.Pref,
		Turns:      nil,
		Candidates: candidates,
	})

	var out reasonListResponse
	if err := s.generateJSON(
		reasonSystemInstruction,
		prompt,
		reasonListSchema,
		&out,
	); err != nil {
		return nil, usecase.GeminiAuditMeta{
			Provider:   "gemini",
			Model:      s.model,
			Status:     "failed",
			DurationMS: time.Since(start).Milliseconds(),
			ErrorType:  "build_reasons_failed",
		}, err
	}

	results := make([]usecase.GeminiReason, 0, len(out.Reasons))
	for _, r := range out.Reasons {
		if strings.TrimSpace(r.Reason) == "" {
			continue
		}
		rank := int(r.SuggestionID)
		if !hasSuggestionRank(in.Suggestions, rank) {
			continue
		}
		results = append(results, usecase.GeminiReason{
			Rank:   rank,
			Reason: strings.TrimSpace(r.Reason),
		})
	}

	return results, usecase.GeminiAuditMeta{
		Provider:   "gemini",
		Model:      s.model,
		Status:     "success",
		DurationMS: time.Since(start).Milliseconds(),
		ErrorType:  "",
	}, nil
}

// 次にユーザーへ確認したい質問候補を最大3件生成
func (s *Service) BuildFollowups(
	in usecase.GeminiFollowupIn,
) ([]string, usecase.GeminiAuditMeta, error) {
	start := time.Now()

	candidates := make([]usecase.QuestionCandidate, 0, len(in.Beans))
	for _, bean := range in.Beans {
		candidates = append(candidates, usecase.QuestionCandidate{
			BeanName:   bean.Name,
			Method:     "",
			ReasonHint: bean.Desc,
		})
	}

	prompt := buildFollowupPrompt(usecase.BuildQuestionsIn{
		Pref:       in.Pref,
		Turns:      nil,
		Candidates: candidates,
		Limit:      3,
	})

	var out questionListResponse
	if err := s.generateJSON(
		followupSystemInstruction,
		prompt,
		questionListSchema,
		&out,
	); err != nil {
		return nil, usecase.GeminiAuditMeta{
			Provider:   "gemini",
			Model:      s.model,
			Status:     "failed",
			DurationMS: time.Since(start).Milliseconds(),
			ErrorType:  "build_followups_failed",
		}, err
	}

	qs := make([]string, 0, len(out.Questions))
	for _, q := range out.Questions {
		v := strings.TrimSpace(q)
		if v == "" {
			continue
		}
		qs = append(qs, v)
		if len(qs) >= 3 {
			break
		}
	}

	return qs, usecase.GeminiAuditMeta{
		Provider:   "gemini",
		Model:      s.model,
		Status:     "success",
		DurationMS: time.Since(start).Milliseconds(),
		ErrorType:  "",
	}, nil
}

// Gemini generateContentを叩いてJSONを受け取る。
func (s *Service) generateJSON(
	systemInstruction string,
	userPrompt string,
	schema json.RawMessage,
	dst interface{},
) error {
	// Gemini API の responseJsonSchema は、モデルやAPI側のJSON Schema対応差分で
	// INVALID_ARGUMENT になりやすい。MVPでは JSON mode に固定し、
	// Go側の json.Unmarshal と usecase/validator で検証する。
	_ = schema

	reqBody := generateContentRequest{
		SystemInstruction: content{
			Parts: []part{
				{Text: systemInstruction},
			},
		},
		Contents: []content{
			{
				Role: "user",
				Parts: []part{
					{Text: userPrompt},
				},
			},
		},
		GenerationConfig: generationConfig{
			ResponseMIMEType: "application/json",
		},
	}

	rawReq, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent", s.baseURL, s.model)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(rawReq))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", s.apiKey)

	res, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[GEMINI] generateContent request failed: %v", err)
		return err
	}
	defer res.Body.Close()

	rawRes, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		err := fmt.Errorf("gemini status=%d body=%s", res.StatusCode, string(rawRes))
		log.Printf("[GEMINI] generateContent failed: %v", err)
		return err
	}

	var gcRes generateContentResponse
	if err := json.Unmarshal(rawRes, &gcRes); err != nil {
		log.Printf("[GEMINI] response unmarshal failed: %v body=%s", err, string(rawRes))
		return err
	}

	text := cleanJSONText(extractCandidateText(gcRes))
	if strings.TrimSpace(text) == "" {
		err := errors.New("gemini returned empty JSON text")
		log.Printf("[GEMINI] %v body=%s", err, string(rawRes))
		return err
	}

	if err := json.Unmarshal([]byte(text), dst); err != nil {
		log.Printf("[GEMINI] invalid JSON text: %v text=%s", err, text)
		return fmt.Errorf("invalid gemini json: %w", err)
	}

	return nil
}

// cleanJSONTextは、JSON mode外の応答やモデル差分で混ざることがある
// Markdown code fenceを取り除く。通常のJSON文字列はそのまま返す。
func cleanJSONText(text string) string {
	t := strings.TrimSpace(text)
	if !strings.HasPrefix(t, "```") {
		return t
	}

	t = strings.TrimPrefix(t, "```json")
	t = strings.TrimPrefix(t, "```JSON")
	t = strings.TrimPrefix(t, "```")
	t = strings.TrimSuffix(t, "```")
	return strings.TrimSpace(t)
}

// 最初の応答のcandidateからtextを抜く。
func extractCandidateText(res generateContentResponse) string {
	if len(res.Candidates) == 0 {
		return ""
	}
	if len(res.Candidates[0].Content.Parts) == 0 {
		return ""
	}

	var b strings.Builder
	for _, p := range res.Candidates[0].Content.Parts {
		if strings.TrimSpace(p.Text) == "" {
			continue
		}
		b.WriteString(p.Text)
	}
	return b.String()
}

// 理由文生成のためにBeanの要約情報を取り出す
func findBeanByID(beans []entity.Bean, id uint) *entity.Bean {
	for i := range beans {
		if beans[i].ID == id {
			return &beans[i]
		}
	}
	return nil
}

func findRecipeByID(recipes []entity.Recipe, id *uint) *entity.Recipe {
	if id == nil {
		return nil
	}
	for i := range recipes {
		if recipes[i].ID == *id {
			return &recipes[i]
		}
	}
	return nil
}

func findRelatedItemTitles(items []entity.Item, itemID *uint) []string {
	if itemID == nil {
		return nil
	}

	out := make([]string, 0, 1)
	for _, item := range items {
		if item.ID == *itemID {
			out = append(out, item.Title)
			break
		}
	}
	return out
}

func findSuggestionRankByID(suggestions []entity.Suggestion, suggestionID uint) int {
	for _, s := range suggestions {
		if s.ID == suggestionID {
			return s.Rank
		}
	}
	return 0
}

type generateContentRequest struct {
	SystemInstruction content          `json:"systemInstruction"`
	Contents          []content        `json:"contents"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

type generationConfig struct {
	ResponseMIMEType   string          `json:"responseMimeType"`
	ResponseJSONSchema json.RawMessage `json:"responseJsonSchema,omitempty"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text,omitempty"`
}

type generateContentResponse struct {
	Candidates []candidate `json:"candidates"`
}

type candidate struct {
	Content content `json:"content"`
}

type conditionDiffResponse struct {
	Flavor     *int             `json:"flavor"`
	Acidity    *int             `json:"acidity"`
	Bitterness *int             `json:"bitterness"`
	Body       *int             `json:"body"`
	Aroma      *int             `json:"aroma"`
	Mood       *entity.Mood     `json:"mood"`
	Method     *entity.Method   `json:"method"`
	Scene      *entity.Scene    `json:"scene"`
	TempPref   *entity.TempPref `json:"temp_pref"`
	Excludes   []string         `json:"excludes"`
	Note       *string          `json:"note"`
}

type searchBundleResponse struct {
	Selections        []beanSelectionResponse `json:"selections"`
	FollowupQuestions []string                `json:"followup_questions"`
}

type beanSelectionListResponse struct {
	Selections []beanSelectionResponse `json:"selections"`
}

type beanSelectionResponse struct {
	BeanID uint   `json:"bean_id"`
	Rank   int    `json:"rank"`
	Score  int    `json:"score"`
	Reason string `json:"reason"`
}

type reasonListResponse struct {
	Reasons []reasonResponse `json:"reasons"`
}

type reasonResponse struct {
	SuggestionID uint   `json:"suggestion_id"`
	Reason       string `json:"reason"`
}

type questionListResponse struct {
	Questions []string `json:"questions"`
}

var conditionDiffSchema = json.RawMessage(`{
  "type":"object",
  "properties":{
    "flavor":{"type":["integer","null"],"minimum":1,"maximum":5},
    "acidity":{"type":["integer","null"],"minimum":1,"maximum":5},
    "bitterness":{"type":["integer","null"],"minimum":1,"maximum":5},
    "body":{"type":["integer","null"],"minimum":1,"maximum":5},
    "aroma":{"type":["integer","null"],"minimum":1,"maximum":5},
    "mood":{"type":["string","null"],"enum":["morning","work","relax","night",null]},
    "method":{"type":["string","null"],"enum":["drip","espresso","milk","iced",null]},
    "scene":{"type":["string","null"],"enum":["work","break","after_meal","relax",null]},
    "temp_pref":{"type":["string","null"],"enum":["hot","ice",null]},
    "excludes":{"type":"array","items":{"type":"string"}},
    "note":{"type":["string","null"]}
  },
  "additionalProperties":false
}`)

var searchBundleSchema = json.RawMessage(`{
  "type":"object",
  "properties":{
    "selections":{
      "type":"array",
      "items":{
        "type":"object",
        "properties":{
          "bean_id":{"type":"integer"},
          "rank":{"type":"integer","minimum":1},
          "score":{"type":"integer","minimum":0,"maximum":100},
          "reason":{"type":"string"}
        },
        "required":["bean_id","rank","score","reason"],
        "additionalProperties":false
      }
    },
    "followup_questions":{
      "type":"array",
      "items":{"type":"string"}
    }
  },
  "required":["selections","followup_questions"],
  "additionalProperties":false
}`)

var beanSelectionListSchema = json.RawMessage(`{
  "type":"object",
  "properties":{
    "selections":{
      "type":"array",
      "items":{
        "type":"object",
        "properties":{
          "bean_id":{"type":"integer"},
          "rank":{"type":"integer","minimum":1},
          "score":{"type":"integer","minimum":0,"maximum":100},
          "reason":{"type":"string"}
        },
        "required":["bean_id","rank","score","reason"],
        "additionalProperties":false
      }
    }
  },
  "required":["selections"],
  "additionalProperties":false
}`)

var reasonListSchema = json.RawMessage(`{
  "type":"object",
  "properties":{
    "reasons":{
      "type":"array",
      "items":{
        "type":"object",
        "properties":{
          "suggestion_id":{"type":"integer"},
          "reason":{"type":"string"}
        },
        "required":["suggestion_id","reason"],
        "additionalProperties":false
      }
    }
  },
  "required":["reasons"],
  "additionalProperties":false
}`)

var questionListSchema = json.RawMessage(`{
  "type":"object",
  "properties":{
    "questions":{
      "type":"array",
      "items":{"type":"string"}
    }
  },
  "required":["questions"],
  "additionalProperties":false
}`)

// 監査用にprovider/modelを返す。
func (s *Service) Info() (string, string) {
	return "gemini", s.model
}

func hasSuggestionRank(suggestions []entity.Suggestion, rank int) bool {
	if rank <= 0 {
		return false
	}

	for _, s := range suggestions {
		if s.Rank == rank {
			return true
		}
	}

	return false
}
