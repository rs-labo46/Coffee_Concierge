package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strconv"
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
	Actor *entity.Actor
	Title string
}

// セッション開始結果。
type StartSessionOut struct {
	Session    entity.Session
	SessionKey string
}

// 初期条件設定入力。
type SetPrefIn struct {
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

// Search系のvalidator。
type SearchVal interface {
	StartSession(in StartSessionIn) error
	SetPref(in SetPrefIn) error
	AddTurn(in AddTurnIn) error
	PatchPref(in PatchPrefIn) error
	GetSession(in GetSessionIn) error
	ListHistory(in ListHistoryIn) error
	CloseSession(in CloseSessionIn) error
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

// 発話から条件差分を作る。
type PrefParser interface {
	ParseTurn(body string, current entity.Pref) (ConditionDiff, error)
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

// suggestion理由文を生成する。
type ExplainSvc interface {
	BuildReason(pref entity.Pref, bean entity.Bean, recipe *entity.Recipe, item *entity.Item) (string, error)
}

// 次の質問候補を生成する。
type FollowupSvc interface {
	BuildQuestions(pref entity.Pref, beans []entity.Bean) ([]string, error)
}

// searchFlowUsecase は SearchFlowUC の実装。
type searchFlowUsecase struct {
	sessions repository.SessionRepository
	beans    repository.BeanRepository
	recipes  repository.RecipeRepository
	items    repository.ItemRepository
	audits   repository.AuditRepository

	val      SearchVal
	parser   PrefParser
	ranker   Ranker
	explain  ExplainSvc
	followup FollowupSvc
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
	parser PrefParser,
	ranker Ranker,
	explain ExplainSvc,
	followup FollowupSvc,
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
		parser:   parser,
		ranker:   ranker,
		explain:  explain,
		followup: followup,
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

	result, err := u.buildResult(*pref)
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

	if u.parser != nil {
		diff, err := u.parser.ParseTurn(in.Body, *pref)
		if err == nil {
			u.applyDiff(pref, diff)
			pref.UpdatedAt = now
			if err := u.sessions.UpdatePref(pref); err != nil {
				return AddTurnOut{}, err
			}
		}
	}

	result, err := u.buildResult(*pref)
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

	result, err := u.buildResult(*pref)
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

// prefを使ってsuggestions / beans / recipes / items / followupsを組み立てる。
func (u *searchFlowUsecase) buildResult(pref entity.Pref) (SearchResult, error) {
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
		var selectedRecipe *entity.Recipe
		var selectedItem *entity.Item

		recipe, err := u.recipes.FindPrimaryByBean(bean.ID, pref.Method, pref.TempPref)
		if err == nil && recipe != nil {
			selectedRecipe = recipe
			recipeID = &recipe.ID
			recipes = append(recipes, *recipe)
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
			selectedItem = &relatedItems[0]
			itemID = &relatedItems[0].ID
			items = appendUniqueItems(items, relatedItems)
		}

		reason := rankItem.Reason
		if u.explain != nil {
			builtReason, err := u.explain.BuildReason(pref, bean, selectedRecipe, selectedItem)
			if err == nil && builtReason != "" {
				reason = builtReason
			}
		}
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

	if err := u.sessions.ReplaceSuggestions(pref.SessionID, suggestions); err != nil {
		return SearchResult{}, err
	}

	followups := make([]string, 0)
	if u.followup != nil {
		qs, err := u.followup.BuildQuestions(pref, beans)
		if err == nil {
			followups = limitStrings(qs, 3)
		}
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
	return bean.Name + " は、今の好みに近い味のバランスです。"
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

func limitStrings(xs []string, limit int) []string {
	if limit <= 0 || len(xs) == 0 {
		return []string{}
	}
	if len(xs) <= limit {
		return xs
	}
	return xs[:limit]
}
