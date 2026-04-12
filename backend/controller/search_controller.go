package controller

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

//session・pref・history・detail
//ゲストの場合はX-Session-Key
//1回ごとの対話の処理

type SearchCtl struct {
	flowUC    usecase.SearchFlowUC
	sessionUC usecase.SessionUC
}

// 生成
type StartSessionReq struct {
	Title string `json:"title"`
}

// 初回条件設定のリクエスト
type SetPrefReq struct {
	Flavor     int             `json:"flavor"`
	Acidity    int             `json:"acidity"`
	Bitterness int             `json:"bitterness"`
	Body       int             `json:"body"`
	Aroma      int             `json:"aroma"`
	Mood       entity.Mood     `json:"mood"`
	Method     entity.Method   `json:"method"`
	Scene      entity.Scene    `json:"scene"`
	TempPref   entity.TempPref `json:"temp_pref"`
	Excludes   []string        `json:"excludes"`
	Note       string          `json:"note"`
}

// 差分の条件更新のリクエスト
type PatchPrefReq struct {
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

// １回の検索結果を返すレスポンス
type SearchResultRes struct {
	Suggestions []entity.Suggestion `json:"suggestions"`
	Beans       []entity.Bean       `json:"beans"`
	Recipes     []entity.Recipe     `json:"recipes"`
	Items       []entity.Item       `json:"items"`
	Followups   []string            `json:"followups"`
}

// sessionの開始レスポンス
type StartSessionRes struct {
	Session    entity.Session `json:"session"`
	SessionKey string         `json:"session_key,omitempty"`
}

// 初回条件の設定と初回検索のレスポンス
type SetPrefRes struct {
	Pref   entity.Pref     `json:"pref"`
	Result SearchResultRes `json:"result"`
}

// 差分条件更新と再検索のレスポンス
type PatchPrefRes struct {
	Pref   entity.Pref     `json:"pref"`
	Result SearchResultRes `json:"result"`
}

// sessionの詳細レスポンス
type SessionDetailRes struct {
	Session     entity.Session      `json:"session"`
	Turns       []entity.Turn       `json:"turns"`
	Pref        entity.Pref         `json:"pref"`
	Suggestions []entity.Suggestion `json:"suggestions"`
}

// 履歴一覧レスポンス
type SessionListRes struct {
	Sessions []entity.Session `json:"sessions"`
}

// POST /search/sessionsを処理(ゲストも可能、その場合はsession_keyを返す)
func (ctl *SearchCtl) StartSession(c echo.Context) error {
	var req StartSessionReq
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}
	actor := actorFromCtx(c)
	out, err := ctl.flowUC.StartSession(usecase.StartSessionIn{
		Actor: actor,
		Title: req.Title,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusCreated, StartSessionRes{
		Session:    out.Session,
		SessionKey: out.SessionKey,
	})
}

// SetPref は POST /search/sessions/:id/pref を処理します。
// guest の場合は X-Session-Key が必要です。
func (ctl *SearchCtl) SetPref(c echo.Context) error {
	sessionID, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}

	var req SetPrefReq
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	actor := actorFromCtx(c)
	sessionKey := ""
	if actor == nil {
		sessionKey, err = requireSessionKey(c)
		if err != nil {
			return writeErr(c, err)
		}
	}

	out, err := ctl.flowUC.SetPref(usecase.SetPrefIn{
		Actor:      actor,
		SessionID:  sessionID,
		SessionKey: sessionKey,
		Flavor:     req.Flavor,
		Acidity:    req.Acidity,
		Bitterness: req.Bitterness,
		Body:       req.Body,
		Aroma:      req.Aroma,
		Mood:       req.Mood,
		Method:     req.Method,
		Scene:      req.Scene,
		TempPref:   req.TempPref,
		Excludes:   req.Excludes,
		Note:       req.Note,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, SetPrefRes{
		Pref:   out.Pref,
		Result: toSearchResultRes(out.Result),
	})
}

// PATCH /search/sessions/:id/pref を処理(ゲストの場合は、X-Session-Key)
func (ctl *SearchCtl) PatchPref(c echo.Context) error {
	sessionID, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}

	var req PatchPrefReq
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	actor := actorFromCtx(c)
	sessionKey := ""
	if actor == nil {
		sessionKey, err = requireSessionKey(c)
		if err != nil {
			return writeErr(c, err)
		}
	}

	out, err := ctl.flowUC.PatchPref(usecase.PatchPrefIn{
		Actor:      actor,
		SessionID:  sessionID,
		SessionKey: sessionKey,
		Flavor:     req.Flavor,
		Acidity:    req.Acidity,
		Bitterness: req.Bitterness,
		Body:       req.Body,
		Aroma:      req.Aroma,
		Mood:       req.Mood,
		Method:     req.Method,
		Scene:      req.Scene,
		TempPref:   req.TempPref,
		Excludes:   req.Excludes,
		Note:       req.Note,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, PatchPrefRes{
		Pref:   out.Pref,
		Result: toSearchResultRes(out.Result),
	})
}

// GET /search/sessions/:id を処理(認証ユーザー用の、sessionの詳細を取得)
func (ctl *SearchCtl) GetSession(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}
	sessionID, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}

	out, err := ctl.sessionUC.GetSession(usecase.GetSessionIn{
		Actor:     actor,
		SessionID: sessionID,
	})
	if err != nil {
		return writeErr(c, err)
	}
	return c.JSON(http.StatusOK, SessionDetailRes{
		Session:     out.Session,
		Turns:       out.Turns,
		Pref:        out.Pref,
		Suggestions: out.Suggestions,
	})
}

// GET /search/guest/sessions/:idを処理(X-session-Key)
func (ctl *SearchCtl) GetGuestSession(c echo.Context) error {
	sessionID, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}

	sessionKey, err := requireSessionKey(c)
	if err != nil {
		return writeErr(c, err)
	}
	out, err := ctl.sessionUC.GetSession(usecase.GetSessionIn{
		Actor:      nil,
		SessionID:  sessionID,
		SessionKey: sessionKey,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, SessionDetailRes{
		Session:     out.Session,
		Turns:       out.Turns,
		Pref:        out.Pref,
		Suggestions: out.Suggestions,
	})
}

// GET /search/sessionsを処理(認証ユーザーのみ)
func (ctl *SearchCtl) ListHistory(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}

	limit, err := qInt(c, "limit", 20)
	if err != nil {
		return writeErr(c, err)
	}

	offset, err := qInt(c, "offset", 0)
	if err != nil {
		return writeErr(c, err)
	}
	sessions, err := ctl.sessionUC.ListHistory(usecase.ListHistoryIn{
		Actor:  *actor,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, SessionListRes{
		Sessions: sessions,
	})
}

// POST /search/sessions/:id/closeを処理
func (ctl *SearchCtl) CloseSession(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}
	sessionID, err := pUint(c, "id")
	if err != nil {
		return writeErr(c, err)
	}
	err = ctl.sessionUC.CloseSession(usecase.CloseSessionIn{
		Actor:     actor,
		SessionID: sessionID,
	})
	if err != nil {
		return writeErr(c, err)
	}

	return c.JSON(http.StatusOK, MsgRes{
		Message: "session closed",
	})

}

// usecase.SearchResultをAPIで返す形に変換
func toSearchResultRes(in usecase.SearchResult) SearchResultRes {
	return SearchResultRes{
		Suggestions: in.Suggestions,
		Beans:       in.Beans,
		Recipes:     in.Recipes,
		Items:       in.Items,
		Followups:   in.Followups,
	}
}
