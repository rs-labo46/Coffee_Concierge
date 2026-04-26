package e2e

import (
	"net/http"
	"testing"
)

// TestE2E45_SearchGuest_GetSessionWrongSessionKey_ReturnsNotFound は、sessionKey不一致でguest sessionを取得できないことを検証する。
func TestE2E45_SearchGuest_GetSessionWrongSessionKey_ReturnsNotFound(t *testing.T) {
	c := newAPIClient(t)
	started := e2e45StartGuestSession(t, c, "E2E45 wrong key")

	status, body, _ := c.getJSON(t, pathForGuestSession(started.Session.ID), map[string]string{
		"X-Session-Key": "wrong-session-key",
	})
	requireStatus(t, status, http.StatusNotFound, body)
	e2e45RequireErrorCode(t, body, "not_found")
}

// TestE2E45_SearchGuest_GetSessionWithoutSessionKey_ReturnsBadRequest は、X-Session-Keyなしのguest取得が400になることを検証する。
func TestE2E45_SearchGuest_GetSessionWithoutSessionKey_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)
	started := e2e45StartGuestSession(t, c, "E2E45 missing key")

	status, body, _ := c.getJSON(t, pathForGuestSession(started.Session.ID), nil)
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}

// TestE2E45_SearchGuest_SetPrefInvalidScore_ReturnsBadRequest は、score範囲外が400になることを検証する。
func TestE2E45_SearchGuest_SetPrefInvalidScore_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)
	started := e2e45StartGuestSession(t, c, "E2E45 invalid score")

	bodyReq := e2e45ValidPrefBody("invalid score")
	bodyReq["flavor"] = 9

	status, body, _ := c.doJSON(t, http.MethodPost, pathForSessionPref(started.Session.ID), bodyReq, map[string]string{
		"X-Session-Key": started.SessionKey,
	})
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}

// TestE2E45_SearchGuest_SetPrefInvalidEnum_ReturnsBadRequest は、enum不正が400になることを検証する。
func TestE2E45_SearchGuest_SetPrefInvalidEnum_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)
	started := e2e45StartGuestSession(t, c, "E2E45 invalid enum")

	bodyReq := e2e45ValidPrefBody("invalid enum")
	bodyReq["mood"] = "bad_mood"

	status, body, _ := c.doJSON(t, http.MethodPost, pathForSessionPref(started.Session.ID), bodyReq, map[string]string{
		"X-Session-Key": started.SessionKey,
	})
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}

// TestE2E45_SearchGuest_PatchPrefInvalidScore_ReturnsBadRequest は、PATCHでscore範囲外が400になることを検証する。
func TestE2E45_SearchGuest_PatchPrefInvalidScore_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)
	started := e2e45StartGuestSession(t, c, "E2E45 patch invalid score")
	e2e45SetGuestPref(t, c, started)

	status, body, _ := c.doJSON(t, http.MethodPatch, pathForSessionPref(started.Session.ID), map[string]interface{}{
		"body": 9,
	}, map[string]string{
		"X-Session-Key": started.SessionKey,
	})
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}

// TestE2E45_SearchGuest_PatchPrefEmptyBody_ReturnsBadRequest は、空PATCHが400になることを検証する。
func TestE2E45_SearchGuest_PatchPrefEmptyBody_ReturnsBadRequest(t *testing.T) {
	c := newAPIClient(t)
	started := e2e45StartGuestSession(t, c, "E2E45 empty patch")
	e2e45SetGuestPref(t, c, started)

	status, body, _ := c.doJSON(t, http.MethodPatch, pathForSessionPref(started.Session.ID), map[string]interface{}{}, map[string]string{
		"X-Session-Key": started.SessionKey,
	})
	requireStatus(t, status, http.StatusBadRequest, body)
	e2e45RequireErrorCode(t, body, "invalid_request")
}

// TestE2E45_SearchGuest_NonexistentSession_ReturnsNotFound は、存在しないguest session取得が404になることを検証する。
func TestE2E45_SearchGuest_NonexistentSession_ReturnsNotFound(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.getJSON(t, "/search/guest/sessions/99999999", map[string]string{
		"X-Session-Key": "some-session-key",
	})
	requireStatus(t, status, http.StatusNotFound, body)
	e2e45RequireErrorCode(t, body, "not_found")
}

// TestE2E45_SearchGuest_ExpiredSession_ReturnsNotFound は、期限切れguest sessionが404になることを検証する。
func TestE2E45_SearchGuest_ExpiredSession_ReturnsNotFound(t *testing.T) {
	c := newAPIClient(t)
	started := e2e45StartGuestSession(t, c, "E2E45 expired session")
	e2e45ExpireGuestSession(t, started.Session.ID)

	status, body, _ := c.getJSON(t, pathForGuestSession(started.Session.ID), map[string]string{
		"X-Session-Key": started.SessionKey,
	})
	requireStatus(t, status, http.StatusNotFound, body)
	e2e45RequireErrorCode(t, body, "not_found")
}
