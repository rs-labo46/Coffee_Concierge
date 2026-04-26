package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestE2E45_Private_SearchHistoryWithoutToken_ReturnsUnauthorized は、履歴一覧が未認証では通らないことを検証する。
func TestE2E45_Private_SearchHistoryWithoutToken_ReturnsUnauthorized(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/search/sessions?limit=20&offset=0", nil)
	requireStatus(t, status, http.StatusUnauthorized, body)
	e2e45RequireErrorCode(t, body, "unauthorized")
}

// TestE2E45_Private_SearchHistoryWithAdminToken_ReturnsOK は、認証済みユーザーが履歴一覧を取得できることを検証する。
func TestE2E45_Private_SearchHistoryWithAdminToken_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	access := e2e45AdminAccessToken(t)

	status, body, _ := c.getJSON(t, "/search/sessions?limit=20&offset=0", bearer(access))
	requireStatus(t, status, http.StatusOK, body)
}

// TestE2E45_Saved_ListWithoutToken_ReturnsUnauthorized は、保存一覧が未認証では通らないことを検証する。
func TestE2E45_Saved_ListWithoutToken_ReturnsUnauthorized(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/saved-suggestions?limit=20&offset=0", nil)
	requireStatus(t, status, http.StatusUnauthorized, body)
	e2e45RequireErrorCode(t, body, "unauthorized")
}

// TestE2E45_Saved_ListWithAdminToken_ReturnsOK は、認証済みユーザーが保存一覧を取得できることを検証する。
func TestE2E45_Saved_ListWithAdminToken_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	access := e2e45AdminAccessToken(t)

	status, body, _ := c.getJSON(t, "/saved-suggestions?limit=20&offset=0", bearer(access))
	requireStatus(t, status, http.StatusOK, body)
}

// TestE2E45_Admin_CreateSourceWithoutToken_ReturnsUnauthorized は、admin routeが未認証を拒否することを検証する。
func TestE2E45_Admin_CreateSourceWithoutToken_ReturnsUnauthorized(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.doJSON(t, http.MethodPost, "/admin/sources", map[string]string{
		"name":     "E2E45 Source Without Token",
		"site_url": "https://example.com",
	}, nil)
	requireStatus(t, status, http.StatusUnauthorized, body)
	e2e45RequireErrorCode(t, body, "unauthorized")
}

// TestE2E45_Admin_CreateSourceWithAdminToken_ReturnsCreated は、adminがSourceを作成できることを検証する。
func TestE2E45_Admin_CreateSourceWithAdminToken_ReturnsCreated(t *testing.T) {
	c := newAPIClient(t)
	access := e2e45AdminAccessToken(t)

	status, body, _ := c.doJSON(t, http.MethodPost, "/admin/sources", map[string]string{
		"name":     fmt.Sprintf("E2E45 Source %d", time.Now().UnixNano()),
		"site_url": "https://example.com/e2e45-source",
	}, bearer(access))
	requireStatus(t, status, http.StatusCreated, body)

	var res struct {
		Source struct {
			ID uint `json:"id"`
		} `json:"source"`
	}
	decodeJSON(t, body, &res)
	if res.Source.ID == 0 {
		t.Fatalf("created source id is zero")
	}
}
