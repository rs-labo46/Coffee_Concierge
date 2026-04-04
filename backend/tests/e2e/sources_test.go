package e2e

import (
	"net/http"
	"testing"
)

func Test_Sources_AdminCreate_Then_PublicList(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	admin := env.createAdminAndLogin(t)

	siteURL := "https://example.com"
	bodyIn := createSourceReq{
		Name:    "Coffee Media",
		SiteURL: &siteURL,
	}

	res, body := postJSON(
		t,
		env,
		"/sources",
		&bodyIn,
		bearerHeader(admin.AccessToken),
	)
	mustStatus(t, res, body, http.StatusCreated)

	var src sourceRes
	mustJSON(t, body, &src)
	if src.Source.ID == 0 {
		t.Fatal("source id should not be zero")
	}

	res, body = env.doEmpty(t, http.MethodGet, "/sources", nil)
	mustStatus(t, res, body, http.StatusOK)

	var out SourceListResCompat
	mustJSON(t, body, &out)
	if len(out.Sources) == 0 {
		t.Fatal("expected at least one source")
	}
	if out.Sources[0].Name != "Coffee Media" {
		t.Fatalf("source name = %q, want Coffee Media", out.Sources[0].Name)
	}
}

func Test_Sources_Create_Forbidden_ForNonAdmin(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	u := env.createUserDirect(t, "user-source@example.com", testUserPW, "user", true)
	st := env.loginOK(t, u.Email, testUserPW)

	siteURL := "https://example.com"
	bodyIn := createSourceReq{
		Name:    "Coffee Media",
		SiteURL: &siteURL,
	}

	res, body := postJSON(
		t,
		env,
		"/sources",
		&bodyIn,
		bearerHeader(st.AccessToken),
	)
	mustStatus(t, res, body, http.StatusForbidden)
	mustErrCode(t, body, "forbidden")
}

type SourceListResCompat struct {
	Sources []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"sources"`
}
