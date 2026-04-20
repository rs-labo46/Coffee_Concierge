package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

// 対話検索フロー(session開始、初期条件設定、発話追加、条件差分更新、再検索)
type SearchFlowUC interface {
	StartSession(in StartSessionIn) (StartSessionOut, error)
	SetPref(in SetPrefIn) (SetPrefOut, error)
	AddTurn(in AddTurnIn) (AddTurnOut, error)
	PatchPref(in PatchPrefIn) (PatchPrefOut, error)
}

// セッション開始入力。
type StartSessionIn struct {
	RequestID string
	Actor     *entity.Actor
	Title     string
}

// セッション開始結果。
type StartSessionOut struct {
	Session    entity.Session
	SessionKey string
}

// 初期条件設定入力。
type SetPrefIn struct {
	RequestID  string
	Actor      *entity.Actor
	SessionID  uint
	SessionKey string
	Flavor     int
	Acidity    int
	Bitterness int
	Body       int
	Aroma      int
	Mood       entity.Mood
	Method     entity.Method
	Scene      entity.Scene
	TempPref   entity.TempPref
	Excludes   []string
	Note       string
}

// 初期条件設定結果。
type SetPrefOut struct {
	Pref   entity.Pref
	Result SearchResult
}

// 発話追加入力。
type AddTurnIn struct {
	RequestID  string
	Actor      *entity.Actor
	SessionID  uint
	SessionKey string
	Body       string
}

// 発話追加結果。
type AddTurnOut struct {
	Turn   entity.Turn
	Pref   entity.Pref
	Result SearchResult
}

// 条件差分更新入力。
type PatchPrefIn struct {
	RequestID  string
	Actor      *entity.Actor
	SessionID  uint
	SessionKey string
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

// 条件差分更新結果。
type PatchPrefOut struct {
	Pref   entity.Pref
	Result SearchResult
}

// 1回の検索結果。
type SearchResult struct {
	Suggestions []entity.Suggestion
	Beans       []entity.Bean
	Recipes     []entity.Recipe
	Items       []entity.Item
	Followups   []string
}

// 発話解析の結果。更新可能キーだけを持つ。
type ConditionDiff struct {
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

// ランカーが返す中間結果。
type RankItem struct {
	Bean   entity.Bean
	Score  int
	Reason string
}

// RBean候補を再ランキングする。
type Ranker interface {
	Rank(pref entity.Pref, beans []entity.Bean) ([]RankItem, error)
}

// searchFlowUsecase は SearchFlowUC の実装。
type searchFlowUsecase struct {
	sessions repository.SessionRepository
	beans    repository.BeanRepository
	recipes  repository.RecipeRepository
	items    repository.ItemRepository
	audits   repository.AuditRepository
	// 入力検証はusecase側。
	val SearchVal
	// 候補のランキング。
	ranker Ranker
	// 発話解釈・理由文・追加質問生成の補助。
	gemini GeminiClient
	// 現在時刻取得とguest session key生成。
	clock    Clock
	idGen    IDGen
	guestTTL time.Duration
}

func NewSearchFlowUsecase(
	sessions repository.SessionRepository,
	beans repository.BeanRepository,
	recipes repository.RecipeRepository,
	items repository.ItemRepository,
	audits repository.AuditRepository,
	val SearchVal,
	ranker Ranker,
	gemini GeminiClient,
	clock Clock,
	idGen IDGen,
	guestTTL time.Duration,
) SearchFlowUC {
	return &searchFlowUsecase{
		sessions: sessions,
		beans:    beans,
		recipes:  recipes,
		items:    items,
		audits:   audits,
		val:      val,
		ranker:   ranker,
		gemini:   gemini,
		clock:    clock,
		idGen:    idGen,
		guestTTL: guestTTL,
	}
}

// 対話検索セッションを開始する。
// guestの場合だけsessionKeyを返す。
func (u *searchFlowUsecase) StartSession(in StartSessionIn) (StartSessionOut, error) {
	if err := u.val.StartSession(in); err != nil {
		return StartSessionOut{}, err
	}

	now := u.clock.Now()

	session := &entity.Session{
		Title:     in.Title,
		Status:    entity.SessionActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	sessionKey := ""

	if in.Actor != nil && in.Actor.UserID > 0 {
		uid := in.Actor.UserID
		session.UserID = &uid
	} else {
		sessionKey = u.idGen.New()
		expiresAt := now.Add(u.guestTTL)
		session.SessionKeyHash = hashText(sessionKey)
		session.GuestExpiresAt = &expiresAt
	}

	if err := u.sessions.CreateSession(session); err != nil {
		return StartSessionOut{}, err
	}

	u.writeAudit(
		"ai.session.start",
		userIDPtr(in.Actor),
		map[string]string{
			"session_id": uintToStr(session.ID),
			"mode":       actorMode(in.Actor),
		},
	)

	return StartSessionOut{
		Session:    *session,
		SessionKey: sessionKey,
	}, nil
}

// 初期条件を保存し、初回検索を返す。
func (u *searchFlowUsecase) SetPref(in SetPrefIn) (SetPrefOut, error) {
	if err := u.val.SetPref(in); err != nil {
		return SetPrefOut{}, err
	}

	session, err := u.resolveWritableSession(in.Actor, in.SessionID, in.SessionKey)
	if err != nil {
		return SetPrefOut{}, err
	}

	if _, err := u.sessions.GetPrefBySessionID(session.ID); err == nil {
		return SetPrefOut{}, repository.ErrConflict
	}
	if err != nil && !errorsIs(err, repository.ErrNotFound) {
		return SetPrefOut{}, err
	}

	now := u.clock.Now()

	pref := &entity.Pref{
		SessionID:  session.ID,
		Flavor:     in.Flavor,
		Acidity:    in.Acidity,
		Bitterness: in.Bitterness,
		Body:       in.Body,
		Aroma:      in.Aroma,
		Mood:       in.Mood,
		Method:     in.Method,
		Scene:      in.Scene,
		TempPref:   in.TempPref,
		Excludes:   in.Excludes,
		Note:       in.Note,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := u.sessions.CreatePref(pref); err != nil {
		return SetPrefOut{}, err
	}

	result, err := u.buildResult(*pref, in.RequestID, userIDPtr(in.Actor))
	if err != nil {
		return SetPrefOut{}, err
	}

	u.writeAudit(
		"ai.pref.set",
		userIDPtr(in.Actor),
		map[string]string{
			"session_id": uintToStr(session.ID),
		},
	)

	return SetPrefOut{
		Pref:   *pref,
		Result: result,
	}, nil
}

// 発話を追加し、条件差分を反映して再検索する。
func (u *searchFlowUsecase) AddTurn(in AddTurnIn) (AddTurnOut, error) {
	if err := u.val.AddTurn(in); err != nil {
		return AddTurnOut{}, err
	}

	session, err := u.resolveWritableSession(in.Actor, in.SessionID, in.SessionKey)
	if err != nil {
		return AddTurnOut{}, err
	}

	now := u.clock.Now()

	turn := &entity.Turn{
		SessionID: session.ID,
		Role:      entity.TurnRoleUser,
		Kind:      entity.TurnKindMessage,
		Body:      in.Body,
		CreatedAt: now,
	}

	if err := u.sessions.CreateTurn(turn); err != nil {
		return AddTurnOut{}, err
	}

	pref, err := u.sessions.GetPrefBySessionID(session.ID)
	if err != nil {
		return AddTurnOut{}, err
	}
	turns, err := u.sessions.ListTurns(session.ID)
	if err != nil {
		return AddTurnOut{}, err
	}

	updated := false

	if u.gemini != nil {
		u.writeAudit(
			"ai.request",
			userIDPtr(in.Actor),
			map[string]string{
				"session_id": uintToStr(session.ID),
				"mode":       "condition_diff",
			},
		)

		diffOut, meta, err := u.gemini.BuildConditionDiff(GeminiConditionDiffIn{
			InputText: in.Body,
			Pref:      *pref,
			Turns:     turns,
		})
		if err == nil {
			diff := toConditionDiff(diffOut)
			u.applyDiff(pref, diff)
			updated = true

			u.writeAudit(
				"ai.success",
				userIDPtr(in.Actor),
				map[string]string{
					"request_id":  in.RequestID,
					"session_id":  uintToStr(session.ID),
					"provider":    meta.Provider,
					"model":       meta.Model,
					"mode":        "condition_diff",
					"status":      meta.Status,
					"duration_ms": strconv.FormatInt(meta.DurationMS, 10),
				},
			)
		} else {
			u.writeAudit(
				"ai.failed",
				userIDPtr(in.Actor),
				map[string]string{
					"request_id":  in.RequestID,
					"session_id":  uintToStr(session.ID),
					"provider":    meta.Provider,
					"model":       meta.Model,
					"mode":        "condition_diff",
					"duration_ms": strconv.FormatInt(meta.DurationMS, 10),
					"error_type":  meta.ErrorType,
				},
			)
		}
	}
	if updated {
		pref.UpdatedAt = now
		if err := u.sessions.UpdatePref(pref); err != nil {
			return AddTurnOut{}, err
		}
	}

	result, err := u.buildResult(*pref, in.RequestID, userIDPtr(in.Actor))
	if err != nil {
		return AddTurnOut{}, err
	}

	u.writeAudit(
		"ai.turn.add",
		userIDPtr(in.Actor),
		map[string]string{
			"session_id": uintToStr(session.ID),
			"turn_id":    uintToStr(turn.ID),
		},
	)

	return AddTurnOut{
		Turn:   *turn,
		Pref:   *pref,
		Result: result,
	}, nil
}

// 条件差分を反映して再検索する。
func (u *searchFlowUsecase) PatchPref(in PatchPrefIn) (PatchPrefOut, error) {
	if err := u.val.PatchPref(in); err != nil {
		return PatchPrefOut{}, err
	}

	session, err := u.resolveWritableSession(in.Actor, in.SessionID, in.SessionKey)
	if err != nil {
		return PatchPrefOut{}, err
	}

	pref, err := u.sessions.GetPrefBySessionID(session.ID)
	if err != nil {
		return PatchPrefOut{}, err
	}

	now := u.clock.Now()

	if in.Flavor != nil {
		pref.Flavor = *in.Flavor
	}
	if in.Acidity != nil {
		pref.Acidity = *in.Acidity
	}
	if in.Bitterness != nil {
		pref.Bitterness = *in.Bitterness
	}
	if in.Body != nil {
		pref.Body = *in.Body
	}
	if in.Aroma != nil {
		pref.Aroma = *in.Aroma
	}
	if in.Mood != nil {
		pref.Mood = *in.Mood
	}
	if in.Method != nil {
		pref.Method = *in.Method
	}
	if in.Scene != nil {
		pref.Scene = *in.Scene
	}
	if in.TempPref != nil {
		pref.TempPref = *in.TempPref
	}
	if in.Excludes != nil {
		pref.Excludes = in.Excludes
	}
	if in.Note != nil {
		pref.Note = *in.Note
	}

	pref.UpdatedAt = now

	if err := u.sessions.UpdatePref(pref); err != nil {
		return PatchPrefOut{}, err
	}

	result, err := u.buildResult(*pref, in.RequestID, userIDPtr(in.Actor))
	if err != nil {
		return PatchPrefOut{}, err
	}

	u.writeAudit(
		"ai.pref.patch",
		userIDPtr(in.Actor),
		map[string]string{
			"session_id": uintToStr(session.ID),
		},
	)

	return PatchPrefOut{
		Pref:   *pref,
		Result: result,
	}, nil
}

// suggestion群に理由文を埋める。Geminiが使えない、または失敗した場合は既存のテンプレ理由文を維持。
func (u *searchFlowUsecase) fillSuggestionReasons(
	pref entity.Pref,
	suggestions *[]entity.Suggestion,
	beans []entity.Bean,
	recipes []entity.Recipe,
	items []entity.Item,
	requestID string,
	userID *uint,
) {
	if suggestions == nil || len(*suggestions) == 0 {
		return
	}

	provider, model := u.aiInfo()

	if u.gemini == nil {
		u.writeAudit(
			"ai.fallback",
			userID,
			map[string]string{
				"request_id":    requestID,
				"session_id":    uintToStr(pref.SessionID),
				"provider":      provider,
				"model":         model,
				"mode":          "reason",
				"fallback_type": "reason_template",
				"reason":        "gemini_client_nil",
			},
		)
		return
	}

	u.writeAudit(
		"ai.request",
		userID,
		map[string]string{
			"request_id": requestID,
			"session_id": uintToStr(pref.SessionID),
			"provider":   provider,
			"model":      model,
			"mode":       "reason",
		},
	)

	reasons, meta, err := u.gemini.BuildReasons(GeminiReasonIn{
		InputText:    pref.Note,
		Pref:         pref,
		Suggestions:  *suggestions,
		Beans:        beans,
		Recipes:      recipes,
		Items:        items,
		ExplainLevel: "normal",
	})
	if err != nil {
		u.writeAudit(
			"ai.failed",
			userID,
			map[string]string{
				"request_id":  requestID,
				"session_id":  uintToStr(pref.SessionID),
				"provider":    meta.Provider,
				"model":       meta.Model,
				"mode":        "reason",
				"duration_ms": strconv.FormatInt(meta.DurationMS, 10),
				"error_type":  meta.ErrorType,
			},
		)

		u.writeAudit(
			"ai.fallback",
			userID,
			map[string]string{
				"request_id":    requestID,
				"session_id":    uintToStr(pref.SessionID),
				"provider":      meta.Provider,
				"model":         meta.Model,
				"mode":          "reason",
				"fallback_type": "reason_template",
				"reason":        "gemini_build_reasons_failed",
			},
		)
		return
	}

	applied := 0
	for _, r := range reasons {
		for i := range *suggestions {
			if (*suggestions)[i].Rank == r.Rank && r.Reason != "" {
				(*suggestions)[i].Reason = r.Reason
				applied++
				break
			}
		}
	}

	if applied == 0 {
		u.writeAudit(
			"ai.fallback",
			userID,
			map[string]string{
				"request_id":    requestID,
				"session_id":    uintToStr(pref.SessionID),
				"provider":      meta.Provider,
				"model":         meta.Model,
				"mode":          "reason",
				"fallback_type": "reason_template",
				"reason":        "empty_reason_result",
			},
		)
		return
	}

	u.writeAudit(
		"ai.success",
		userID,
		map[string]string{
			"request_id":  requestID,
			"session_id":  uintToStr(pref.SessionID),
			"provider":    meta.Provider,
			"model":       meta.Model,
			"mode":        "reason",
			"status":      meta.Status,
			"duration_ms": strconv.FormatInt(meta.DurationMS, 10),
		},
	)
}

// 次の質問候補を組み立てる。Geminiが失敗した場合は空配列を返し、検索結果本体は止めない。
func (u *searchFlowUsecase) buildFollowups(
	pref entity.Pref,
	beans []entity.Bean,
	requestID string,
	userID *uint,
) []string {
	provider, model := u.aiInfo()

	if u.gemini == nil || len(beans) == 0 {
		u.writeAudit(
			"ai.fallback",
			userID,
			map[string]string{
				"request_id":    requestID,
				"session_id":    uintToStr(pref.SessionID),
				"provider":      provider,
				"model":         model,
				"mode":          "followup",
				"fallback_type": "empty_followups",
				"reason":        "gemini_client_nil_or_no_beans",
			},
		)
		return []string{}
	}

	u.writeAudit(
		"ai.request",
		userID,
		map[string]string{
			"request_id": requestID,
			"session_id": uintToStr(pref.SessionID),
			"provider":   provider,
			"model":      model,
			"mode":       "followup",
		},
	)

	qs, meta, err := u.gemini.BuildFollowups(GeminiFollowupIn{
		InputText: pref.Note,
		Pref:      pref,
		Beans:     beans,
	})
	if err != nil {
		u.writeAudit(
			"ai.failed",
			userID,
			map[string]string{
				"request_id":  requestID,
				"session_id":  uintToStr(pref.SessionID),
				"provider":    meta.Provider,
				"model":       meta.Model,
				"mode":        "followup",
				"duration_ms": strconv.FormatInt(meta.DurationMS, 10),
				"error_type":  meta.ErrorType,
			},
		)

		u.writeAudit(
			"ai.fallback",
			userID,
			map[string]string{
				"request_id":    requestID,
				"session_id":    uintToStr(pref.SessionID),
				"provider":      meta.Provider,
				"model":         meta.Model,
				"mode":          "followup",
				"fallback_type": "empty_followups",
				"reason":        "gemini_build_followups_failed",
			},
		)
		return []string{}
	}

	qs = limitStrings(qs, 3)
	if len(qs) == 0 {
		u.writeAudit(
			"ai.fallback",
			userID,
			map[string]string{
				"request_id":    requestID,
				"session_id":    uintToStr(pref.SessionID),
				"provider":      meta.Provider,
				"model":         meta.Model,
				"mode":          "followup",
				"fallback_type": "empty_followups",
				"reason":        "empty_followup_result",
			},
		)
		return []string{}
	}

	u.writeAudit(
		"ai.success",
		userID,
		map[string]string{
			"request_id":  requestID,
			"session_id":  uintToStr(pref.SessionID),
			"provider":    meta.Provider,
			"model":       meta.Model,
			"mode":        "followup",
			"status":      meta.Status,
			"duration_ms": strconv.FormatInt(meta.DurationMS, 10),
		},
	)

	return qs
}

// prefを使ってsuggestions / beans / recipes / items / followupsを組み立てる。
func (u *searchFlowUsecase) buildResult(pref entity.Pref, requestID string, userID *uint) (SearchResult, error) {
	beanList, err := u.beans.SearchByPref(pref, 10)
	if err != nil {
		return SearchResult{}, err
	}

	ranked, err := u.rankBeans(pref, beanList)
	if err != nil {
		return SearchResult{}, err
	}

	now := u.clock.Now()

	suggestions := make([]entity.Suggestion, 0, len(ranked))
	beans := make([]entity.Bean, 0, len(ranked))
	recipes := make([]entity.Recipe, 0, len(ranked))
	items := make([]entity.Item, 0, len(ranked))

	for i, rankItem := range ranked {
		bean := rankItem.Bean
		beans = append(beans, bean)

		var recipeID *uint
		var itemID *uint

		recipe, err := u.recipes.FindPrimaryByBean(bean.ID, pref.Method, pref.TempPref)
		if err == nil && recipe != nil {
			recipeID = &recipe.ID
			recipes = appendUniqueRecipes(recipes, []entity.Recipe{*recipe})
		}

		relatedItems, err := u.items.SearchRelated(
			bean.Name,
			bean.Roast,
			bean.Origin,
			pref.Mood,
			pref.Method,
			3,
			now,
		)
		if err == nil && len(relatedItems) > 0 {
			itemID = &relatedItems[0].ID
			items = appendUniqueItems(items, relatedItems)
		}

		reason := rankItem.Reason
		if reason == "" {
			reason = defaultReason(bean, pref)
		}

		suggestions = append(suggestions, entity.Suggestion{
			SessionID: pref.SessionID,
			BeanID:    bean.ID,
			RecipeID:  recipeID,
			ItemID:    itemID,
			Score:     rankItem.Score,
			Reason:    reason,
			Rank:      i + 1,
		})
	}

	u.fillSuggestionReasons(pref, &suggestions, beans, recipes, items, requestID, userID)
	followups := u.buildFollowups(pref, beans, requestID, userID)
	if err := u.sessions.ReplaceSuggestions(pref.SessionID, suggestions); err != nil {
		return SearchResult{}, err
	}

	u.writeAudit(
		"ai.suggest.build",
		nil,
		map[string]string{
			"session_id":       uintToStr(pref.SessionID),
			"suggestion_count": strconv.Itoa(len(suggestions)),
		},
	)

	u.writeAudit(
		"ai.explain.build",
		nil,
		map[string]string{
			"session_id":       uintToStr(pref.SessionID),
			"suggestion_count": strconv.Itoa(len(suggestions)),
		},
	)

	return SearchResult{
		Suggestions: suggestions,
		Beans:       beans,
		Recipes:     recipes,
		Items:       items,
		Followups:   followups,
	}, nil
}

// actorまたはsessionKeyで書き込み可能なsessionを解決する。
func (u *searchFlowUsecase) resolveWritableSession(
	actor *entity.Actor,
	sessionID uint,
	sessionKey string,
) (*entity.Session, error) {
	if actor != nil && actor.UserID > 0 {
		session, err := u.sessions.GetSessionByID(sessionID)
		if err != nil {
			return nil, err
		}
		if session.UserID == nil {
			return nil, repository.ErrForbidden
		}
		if *session.UserID != actor.UserID {
			return nil, repository.ErrForbidden
		}
		if session.Status != entity.SessionActive {
			return nil, repository.ErrConflict
		}
		return session, nil
	}

	if sessionKey == "" {
		return nil, repository.ErrUnauthorized
	}

	session, err := u.sessions.GetGuestSessionByID(
		sessionID,
		hashText(sessionKey),
		u.clock.Now(),
	)
	if err != nil {
		return nil, err
	}
	if session.Status != entity.SessionActive {
		return nil, repository.ErrConflict
	}
	return session, nil
}

// ランカーがあれば使い、なければ簡易スコアで並べる。
func (u *searchFlowUsecase) rankBeans(pref entity.Pref, beans []entity.Bean) ([]RankItem, error) {
	if u.ranker != nil {
		return u.ranker.Rank(pref, beans)
	}

	out := make([]RankItem, 0, len(beans))
	for _, bean := range beans {
		score := simpleScore(pref, bean)
		out = append(out, RankItem{
			Bean:   bean,
			Score:  score,
			Reason: "",
		})
	}

	sort.Slice(out, func(i int, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Bean.ID < out[j].Bean.ID
		}
		return out[i].Score > out[j].Score
	})

	return out, nil
}

// parser結果をprefに反映する。
func (u *searchFlowUsecase) applyDiff(pref *entity.Pref, diff ConditionDiff) {
	if diff.Flavor != nil {
		pref.Flavor = *diff.Flavor
	}
	if diff.Acidity != nil {
		pref.Acidity = *diff.Acidity
	}
	if diff.Bitterness != nil {
		pref.Bitterness = *diff.Bitterness
	}
	if diff.Body != nil {
		pref.Body = *diff.Body
	}
	if diff.Aroma != nil {
		pref.Aroma = *diff.Aroma
	}
	if diff.Mood != nil {
		pref.Mood = *diff.Mood
	}
	if diff.Method != nil {
		pref.Method = *diff.Method
	}
	if diff.Scene != nil {
		pref.Scene = *diff.Scene
	}
	if diff.TempPref != nil {
		pref.TempPref = *diff.TempPref
	}
	if diff.Excludes != nil {
		pref.Excludes = diff.Excludes
	}
	if diff.Note != nil {
		pref.Note = *diff.Note
	}
}

func (u *searchFlowUsecase) writeAudit(
	typ string,
	userID *uint,
	meta map[string]string,
) {
	if u.audits == nil {
		return
	}

	raw, err := json.Marshal(meta)
	if err != nil {
		raw = []byte(`{}`)
	}

	_ = u.audits.Create(&entity.AuditLog{
		Type:   typ,
		UserID: userID,
		IP:     "",
		UA:     "",
		Meta:   raw,
	})
}

func userIDPtr(actor *entity.Actor) *uint {
	if actor == nil || actor.UserID == 0 {
		return nil
	}
	id := actor.UserID
	return &id
}

func actorMode(actor *entity.Actor) string {
	if actor == nil || actor.UserID == 0 {
		return "guest"
	}
	return "user"
}

func hashText(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

func errorsIs(err error, target error) bool {
	return err != nil && target != nil && err.Error() == target.Error()
}

func simpleScore(pref entity.Pref, bean entity.Bean) int {
	score := 0
	score += 10 - abs(pref.Flavor-bean.Flavor)
	score += 10 - abs(pref.Acidity-bean.Acidity)
	score += 10 - abs(pref.Bitterness-bean.Bitterness)
	score += 10 - abs(pref.Body-bean.Body)
	score += 10 - abs(pref.Aroma-bean.Aroma)
	return score
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func defaultReason(bean entity.Bean, pref entity.Pref) string {
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
		return bean.Name + " は、今の条件に近い候補です。"
	}

	return bean.Name + " は、" + strings.Join(parts, "で、今の好みに合いやすい候補です。")
}

func appendUniqueItems(base []entity.Item, extra []entity.Item) []entity.Item {
	seen := make(map[uint]struct{}, len(base))
	for _, item := range base {
		seen[item.ID] = struct{}{}
	}

	out := make([]entity.Item, 0, len(base)+len(extra))
	out = append(out, base...)

	for _, item := range extra {
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		out = append(out, item)
	}

	return out
}

func toConditionDiff(in GeminiConditionDiffOut) ConditionDiff {
	return ConditionDiff{
		Flavor:     in.Flavor,
		Acidity:    in.Acidity,
		Bitterness: in.Bitterness,
		Body:       in.Body,
		Aroma:      in.Aroma,
		Mood:       in.Mood,
		Method:     in.Method,
		Scene:      in.Scene,
		TempPref:   in.TempPref,
		Excludes:   in.Excludes,
		Note:       in.Note,
	}
}

func limitStrings(xs []string, limit int) []string {
	if limit <= 0 || len(xs) == 0 {
		return []string{}
	}
	if len(xs) <= limit {
		return xs
	}
	return xs[:limit]
}

// 同じrecipe.IDを2回以上入れないためのhelper
func appendUniqueRecipes(base []entity.Recipe, extra []entity.Recipe) []entity.Recipe {
	seen := make(map[uint]struct{}, len(base))
	for _, recipe := range base {
		seen[recipe.ID] = struct{}{}
	}

	out := make([]entity.Recipe, 0, len(base)+len(extra))
	out = append(out, base...)

	for _, recipe := range extra {
		if _, ok := seen[recipe.ID]; ok {
			continue
		}
		seen[recipe.ID] = struct{}{}
		out = append(out, recipe)
	}

	return out
}

// 監査ログ用にuintを文字列化。
func uintToStr(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

// GeminiClientの具体的なmodel名が分からない時に監査用補助値を返す。
// 成功/失敗時はGeminiAuditMetaのmodelを優先し、request時の暫定値としてだけ使う。
func (u *searchFlowUsecase) aiInfo() (string, string) {
	if u.gemini == nil {
		return "unknown", "unknown"
	}
	return u.gemini.Info()
}
