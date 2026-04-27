package e2e

import (
	"net/http"
	"testing"
)

// TestE2E45_Private_SearchStartSetPrefGetSession_ReturnsOK は、認証ユーザーがsession開始、初期条件設定、詳細取得まで通せることを検証する。
func TestE2E45_Private_SearchStartSetPrefGetSession_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	access := e2e45AdminAccessToken(t)
	auth := bearer(access)

	started := e2e45StartPrivateSession(t, c, auth, "E2E45 private search")
	setRes := e2e45SetPrivatePref(t, c, auth, started.Session.ID, "認証ユーザーとして朝に飲みたい")

	if len(setRes.Result.Suggestions) == 0 {
		t.Fatalf("suggestions is empty")
	}
	if setRes.Result.Suggestions[0].ID == 0 {
		t.Fatalf("suggestion id is zero")
	}
	if setRes.Result.Suggestions[0].SessionID != started.Session.ID {
		t.Fatalf("unexpected suggestion session id: got=%d want=%d", setRes.Result.Suggestions[0].SessionID, started.Session.ID)
	}

	status, body, _ := c.getJSON(t, pathForPrivateSession(started.Session.ID), auth)
	requireStatus(t, status, http.StatusOK, body)

	var detail e2e45PrivateSessionDetailRes
	decodeJSON(t, body, &detail)
	if detail.Session.ID != started.Session.ID {
		t.Fatalf("unexpected detail session id: got=%d want=%d", detail.Session.ID, started.Session.ID)
	}
	if detail.Session.UserID == nil || *detail.Session.UserID == 0 {
		t.Fatalf("detail session user_id is empty")
	}
	if detail.Pref.SessionID != started.Session.ID {
		t.Fatalf("unexpected detail pref session id: got=%d want=%d", detail.Pref.SessionID, started.Session.ID)
	}
	if len(detail.Suggestions) == 0 {
		t.Fatalf("detail suggestions is empty")
	}
}

// TestE2E45_Private_SearchPatchPref_ReturnsOK は、認証ユーザーがsessionKeyなしで条件差分更新できることを検証する。
func TestE2E45_Private_SearchPatchPref_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	access := e2e45AdminAccessToken(t)
	auth := bearer(access)

	started := e2e45StartPrivateSession(t, c, auth, "E2E45 private patch")
	e2e45SetPrivatePref(t, c, auth, started.Session.ID, "初期条件")

	bodyScore := 2
	note := "もう少し軽めにしたい"
	status, body, _ := c.doJSON(t, http.MethodPatch, pathForSessionPref(started.Session.ID), e2e45PatchPrefReq{
		Body: &bodyScore,
		Note: &note,
	}, auth)
	requireStatus(t, status, http.StatusOK, body)

	var patchRes e2e45SetPrefRes
	decodeJSON(t, body, &patchRes)
	if patchRes.Pref.SessionID != started.Session.ID {
		t.Fatalf("unexpected patch pref session id: got=%d want=%d", patchRes.Pref.SessionID, started.Session.ID)
	}
	if patchRes.Pref.Body != bodyScore {
		t.Fatalf("unexpected patched body: got=%d want=%d", patchRes.Pref.Body, bodyScore)
	}
	if len(patchRes.Result.Suggestions) == 0 {
		t.Fatalf("patch suggestions is empty")
	}
}

// TestE2E45_Saved_CreateWithAdminToken_ReturnsCreated は、認証ユーザーが自分の検索sessionのsuggestionを保存できることを検証する。
func TestE2E45_Saved_CreateWithAdminToken_ReturnsCreated(t *testing.T) {
	c := newAPIClient(t)
	access := e2e45AdminAccessToken(t)
	auth := bearer(access)

	started := e2e45StartPrivateSession(t, c, auth, "E2E45 saved create")
	setRes := e2e45SetPrivatePref(t, c, auth, started.Session.ID, "保存したい候補を作る")
	if len(setRes.Result.Suggestions) == 0 {
		t.Fatalf("suggestions is empty")
	}

	suggestionID := setRes.Result.Suggestions[0].ID
	status, body, _ := c.doJSON(t, http.MethodPost, "/saved-suggestions", e2e45SaveSuggestionReq{
		SessionID:    started.Session.ID,
		SuggestionID: suggestionID,
	}, auth)
	requireStatus(t, status, http.StatusCreated, body)

	var savedRes e2e45SavedSuggestionRes
	decodeJSON(t, body, &savedRes)
	if savedRes.Saved.ID == 0 {
		t.Fatalf("saved id is zero")
	}
	if savedRes.Saved.SessionID != started.Session.ID {
		t.Fatalf("unexpected saved session id: got=%d want=%d", savedRes.Saved.SessionID, started.Session.ID)
	}
	if savedRes.Saved.SuggestionID != suggestionID {
		t.Fatalf("unexpected saved suggestion id: got=%d want=%d", savedRes.Saved.SuggestionID, suggestionID)
	}
}

// TestE2E45_Saved_CreateDuplicateWithAdminToken_ReturnsConflict は、同じsuggestionの二重保存を拒否できることを検証する。
func TestE2E45_Saved_CreateDuplicateWithAdminToken_ReturnsConflict(t *testing.T) {
	c := newAPIClient(t)
	access := e2e45AdminAccessToken(t)
	auth := bearer(access)

	started := e2e45StartPrivateSession(t, c, auth, "E2E45 saved duplicate")
	setRes := e2e45SetPrivatePref(t, c, auth, started.Session.ID, "二重保存検証")
	if len(setRes.Result.Suggestions) == 0 {
		t.Fatalf("suggestions is empty")
	}

	req := e2e45SaveSuggestionReq{
		SessionID:    started.Session.ID,
		SuggestionID: setRes.Result.Suggestions[0].ID,
	}

	status, body, _ := c.doJSON(t, http.MethodPost, "/saved-suggestions", req, auth)
	requireStatus(t, status, http.StatusCreated, body)

	status, body, _ = c.doJSON(t, http.MethodPost, "/saved-suggestions", req, auth)
	requireStatus(t, status, http.StatusConflict, body)
	e2e45RequireErrorCode(t, body, "conflict")
}

func e2e45StartPrivateSession(t *testing.T, c *apiClient, headers map[string]string, title string) e2e45PrivateStartSessionRes {
	t.Helper()

	status, body, _ := c.doJSON(t, http.MethodPost, "/search/sessions", e2e45StartSessionReq{
		Title: title,
	}, headers)
	requireStatus(t, status, http.StatusCreated, body)

	var res e2e45PrivateStartSessionRes
	decodeJSON(t, body, &res)
	if res.Session.ID == 0 {
		t.Fatalf("session id is zero")
	}
	if res.Session.UserID == nil || *res.Session.UserID == 0 {
		t.Fatalf("private session user_id is empty")
	}
	if res.SessionKey != "" {
		t.Fatalf("private session must not return session_key")
	}

	return res
}

func e2e45SetPrivatePref(t *testing.T, c *apiClient, headers map[string]string, sessionID uint, note string) e2e45SetPrefRes {
	t.Helper()

	status, body, _ := c.doJSON(t, http.MethodPost, pathForSessionPref(sessionID), e2e45SetPrefReq{
		Flavor:     3,
		Acidity:    2,
		Bitterness: 2,
		Body:       3,
		Aroma:      4,
		Mood:       "morning",
		Method:     "drip",
		Scene:      "break",
		TempPref:   "hot",
		Excludes:   []string{},
		Note:       note,
	}, headers)
	requireStatus(t, status, http.StatusOK, body)

	var res e2e45SetPrefRes
	decodeJSON(t, body, &res)
	if res.Pref.SessionID != sessionID {
		t.Fatalf("unexpected pref session id: got=%d want=%d", res.Pref.SessionID, sessionID)
	}

	return res
}

func pathForPrivateSession(id uint) string {
	return "/search/sessions/" + uintToString(id)
}

type e2e45StartSessionReq struct {
	Title string `json:"title"`
}

type e2e45SetPrefReq struct {
	Flavor     int      `json:"flavor"`
	Acidity    int      `json:"acidity"`
	Bitterness int      `json:"bitterness"`
	Body       int      `json:"body"`
	Aroma      int      `json:"aroma"`
	Mood       string   `json:"mood"`
	Method     string   `json:"method"`
	Scene      string   `json:"scene"`
	TempPref   string   `json:"temp_pref"`
	Excludes   []string `json:"excludes"`
	Note       string   `json:"note"`
}

type e2e45PatchPrefReq struct {
	Flavor     *int     `json:"flavor,omitempty"`
	Acidity    *int     `json:"acidity,omitempty"`
	Bitterness *int     `json:"bitterness,omitempty"`
	Body       *int     `json:"body,omitempty"`
	Aroma      *int     `json:"aroma,omitempty"`
	Mood       *string  `json:"mood,omitempty"`
	Method     *string  `json:"method,omitempty"`
	Scene      *string  `json:"scene,omitempty"`
	TempPref   *string  `json:"temp_pref,omitempty"`
	Excludes   []string `json:"excludes,omitempty"`
	Note       *string  `json:"note,omitempty"`
}

type e2e45SaveSuggestionReq struct {
	SessionID    uint `json:"session_id"`
	SuggestionID uint `json:"suggestion_id"`
}

type e2e45PrivateStartSessionRes struct {
	Session    e2e45SessionRes `json:"session"`
	SessionKey string          `json:"session_key"`
}

type e2e45SetPrefRes struct {
	Pref   e2e45PrefRes         `json:"pref"`
	Result e2e45SearchResultRes `json:"result"`
}

type e2e45PrivateSessionDetailRes struct {
	Session     e2e45SessionRes      `json:"session"`
	Turns       []e2e45TurnRes       `json:"turns"`
	Pref        e2e45PrefRes         `json:"pref"`
	Suggestions []e2e45SuggestionRes `json:"suggestions"`
}

type e2e45SavedSuggestionRes struct {
	Saved e2e45SavedRes `json:"saved"`
}

type e2e45SearchResultRes struct {
	Suggestions []e2e45SuggestionRes `json:"suggestions"`
	Beans       []e2e45BeanRes       `json:"beans"`
	Recipes     []e2e45RecipeRes     `json:"recipes"`
	Items       []e2e45ItemRes       `json:"items"`
	Followups   []string             `json:"followups"`
}

type e2e45SessionRes struct {
	ID     uint   `json:"id"`
	UserID *uint  `json:"user_id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

type e2e45PrefRes struct {
	ID        uint   `json:"id"`
	SessionID uint   `json:"session_id"`
	Flavor    int    `json:"flavor"`
	Acidity   int    `json:"acidity"`
	Body      int    `json:"body"`
	Mood      string `json:"mood"`
	Method    string `json:"method"`
	Scene     string `json:"scene"`
	TempPref  string `json:"temp_pref"`
	Note      string `json:"note"`
}

type e2e45SuggestionRes struct {
	ID        uint   `json:"id"`
	SessionID uint   `json:"session_id"`
	BeanID    uint   `json:"bean_id"`
	RecipeID  *uint  `json:"recipe_id"`
	ItemID    *uint  `json:"item_id"`
	Score     int    `json:"score"`
	Reason    string `json:"reason"`
	Rank      int    `json:"rank"`
}

type e2e45SavedRes struct {
	ID           uint `json:"id"`
	UserID       uint `json:"user_id"`
	SessionID    uint `json:"session_id"`
	SuggestionID uint `json:"suggestion_id"`
}

type e2e45TurnRes struct {
	ID        uint   `json:"id"`
	SessionID uint   `json:"session_id"`
	Role      string `json:"role"`
	Kind      string `json:"kind"`
	Body      string `json:"body"`
}

type e2e45BeanRes struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type e2e45RecipeRes struct {
	ID     uint   `json:"id"`
	BeanID uint   `json:"bean_id"`
	Name   string `json:"name"`
}

type e2e45ItemRes struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
}
