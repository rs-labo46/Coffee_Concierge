package e2e

import (
	"fmt"
	"net/http"
	"testing"
)

// 未ログインのguestユーザーが、対話型検索を使えるか確認
type startSessionRes struct {
	Session struct {
		ID     uint   `json:"id"`
		Title  string `json:"title"`
		Status string `json:"status"`
	} `json:"session"`
	SessionKey string `json:"session_key"`
}

// guestユーザーがログインなしで検索sessionを開始し、初期条件を登録し、session詳細を取得できるか確認
func TestE2E_SearchGuest_StartSetPrefAndGetSession(t *testing.T) {
	c := newAPIClient(t)

	startStatus, startBody, _ := c.doJSON(t, http.MethodPost, "/search/sessions", map[string]string{
		"title": "E2E guest search",
	}, nil)
	requireStatus(t, startStatus, http.StatusCreated, startBody)

	var started startSessionRes
	decodeJSON(t, startBody, &started)
	if started.Session.ID == 0 {
		t.Fatalf("session id is zero")
	}
	if started.SessionKey == "" {
		t.Fatalf("guest session key is empty")
	}

	headers := map[string]string{
		"X-Session-Key": started.SessionKey,
	}

	setStatus, setBody, _ := c.doJSON(t, http.MethodPost, pathForSessionPref(started.Session.ID), map[string]interface{}{
		"flavor":     3,
		"acidity":    2,
		"bitterness": 2,
		"body":       3,
		"aroma":      4,
		"mood":       "morning",
		"method":     "drip",
		"scene":      "break",
		"temp_pref":  "hot",
		"excludes":   []string{},
		"note":       "朝に飲みたい",
	}, headers)
	requireStatus(t, setStatus, http.StatusOK, setBody)

	var setRes struct {
		Pref struct {
			SessionID uint   `json:"session_id"`
			Mood      string `json:"mood"`
			Method    string `json:"method"`
		} `json:"pref"`
		Result struct {
			Suggestions []map[string]interface{} `json:"suggestions"`
			Beans       []map[string]interface{} `json:"beans"`
			Recipes     []map[string]interface{} `json:"recipes"`
			Items       []map[string]interface{} `json:"items"`
			Followups   []string                 `json:"followups"`
		} `json:"result"`
	}
	decodeJSON(t, setBody, &setRes)
	if setRes.Pref.SessionID != started.Session.ID {
		t.Fatalf("unexpected pref session id: got=%d want=%d", setRes.Pref.SessionID, started.Session.ID)
	}

	getStatus, getBody, _ := c.getJSON(t, pathForGuestSession(started.Session.ID), headers)
	requireStatus(t, getStatus, http.StatusOK, getBody)

	var detail struct {
		Session struct {
			ID uint `json:"id"`
		} `json:"session"`
		Pref struct {
			SessionID uint `json:"session_id"`
		} `json:"pref"`
	}
	decodeJSON(t, getBody, &detail)
	if detail.Session.ID != started.Session.ID {
		t.Fatalf("unexpected detail session id: got=%d want=%d", detail.Session.ID, started.Session.ID)
	}
}

// guest sessionに対して、X-Session-Keyなしでpref登録しようとした場合に拒否されるか確認
func TestE2E_SearchGuest_SetPrefWithoutSessionKey_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)

	startStatus, startBody, _ := c.doJSON(t, http.MethodPost, "/search/sessions", map[string]string{
		"title": "E2E missing session key",
	}, nil)
	requireStatus(t, startStatus, http.StatusCreated, startBody)

	var started startSessionRes
	decodeJSON(t, startBody, &started)

	status, body, _ := c.doJSON(t, http.MethodPost, pathForSessionPref(started.Session.ID), map[string]interface{}{
		"flavor":     3,
		"acidity":    3,
		"bitterness": 3,
		"body":       3,
		"aroma":      3,
		"mood":       "relax",
		"method":     "drip",
		"scene":      "relax",
		"temp_pref":  "hot",
	}, nil)
	requireStatus(t, status, http.StatusBadRequest, body)
}

func pathForSessionPref(id uint) string {
	return "/search/sessions/" + uintToString(id) + "/pref"
}

func pathForGuestSession(id uint) string {
	return "/search/guest/sessions/" + uintToString(id)
}

func uintToString(v uint) string {
	return fmt.Sprintf("%d", v)
}
