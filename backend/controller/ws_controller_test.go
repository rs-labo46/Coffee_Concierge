package controller

import (
	"net/http"
	"testing"

	"coffee-spa/testutil/controllertest"
)

func TestSessionKeyFromWs_PrefersQueryThenHeader(t *testing.T) {
	_, c, _ := controllertest.EmptyContext(http.MethodGet, "/ws/guest/search/sessions/1?session_key=query-key")
	c.Request().Header.Set(HeaderSessionKey, "header-key")
	if got := sessionKeyFromWs(c); got != "query-key" {
		t.Fatalf("sessionKeyFromWs() = %q, want query-key", got)
	}

	_, c, _ = controllertest.EmptyContext(http.MethodGet, "/ws/guest/search/sessions/1")
	c.Request().Header.Set(HeaderSessionKey, "header-key")
	if got := sessionKeyFromWs(c); got != "header-key" {
		t.Fatalf("sessionKeyFromWs() = %q, want header-key", got)
	}
}
