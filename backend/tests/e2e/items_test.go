package e2e

import (
	"net/http"
	"testing"
	"time"
)

func Test_Items_AdminCreate_Then_PublicTopAndList(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	admin := env.createAdminAndLogin(t)

	siteURL := "https://example.com"

	//source作成
	sourceBody := createSourceReq{
		Name:    "Coffee Media",
		SiteURL: &siteURL,
	}

	res, body := postJSON(
		t,
		env,
		"/sources",
		&sourceBody,
		bearerHeader(admin.AccessToken),
	)
	mustStatus(t, res, body, http.StatusCreated)

	var src sourceRes
	mustJSON(t, body, &src)

	if src.Source.ID == 0 {
		t.Fatalf("source id should not be zero")
	}

	//item作成
	publishedAt := time.Now().UTC().Format(time.RFC3339)
	summary := "Great coffee event"
	url := "https://example.com/news/1"
	imageURL := "https://example.com/image/1.png"

	itemBody := createItemReq{
		Title:       "Coffee Event News",
		Summary:     &summary,
		URL:         &url,
		ImageURL:    &imageURL,
		Kind:        "news",
		SourceID:    src.Source.ID,
		PublishedAt: publishedAt,
	}

	res, body = postJSON(
		t,
		env,
		"/items",
		&itemBody,
		bearerHeader(admin.AccessToken),
	)
	mustStatus(t, res, body, http.StatusCreated)

	var it itemRes
	mustJSON(t, body, &it)

	if it.Item.ID == 0 {
		t.Fatalf("item id should not be zero")
	}
	if it.Item.Kind != "news" {
		t.Fatalf("item kind=%q want=news", it.Item.Kind)
	}

	///items/top?limit=0 は 4キー固定で空配列
	res, body = env.doEmpty(
		t,
		http.MethodGet,
		"/items/top?limit=0",
		nil,
	)
	mustStatus(t, res, body, http.StatusOK)

	var top0 topItemsRes
	mustJSON(t, body, &top0)

	if top0.News == nil || top0.Recipe == nil || top0.Deal == nil || top0.Shop == nil {
		t.Fatalf("top groups should all exist")
	}
	if len(top0.News) != 0 || len(top0.Recipe) != 0 || len(top0.Deal) != 0 || len(top0.Shop) != 0 {
		t.Fatalf("limit=0 should return empty arrays")
	}

	///items/top?limit=3 ではnewsが返る
	res, body = env.doEmpty(
		t,
		http.MethodGet,
		"/items/top?limit=3",
		nil,
	)
	mustStatus(t, res, body, http.StatusOK)

	var top3 topItemsRes
	mustJSON(t, body, &top3)

	if len(top3.News) == 0 {
		t.Fatalf("expected news items in top")
	}

	//public listでkind / qを使って探す
	res, body = env.doEmpty(
		t,
		http.MethodGet,
		"/items?kind=news&q=Coffee&limit=20&offset=0",
		nil,
	)
	mustStatus(t, res, body, http.StatusOK)

	var list itemListRes
	mustJSON(t, body, &list)

	if len(list.Items) == 0 {
		t.Fatalf("expected items in public list")
	}
	if list.Items[0].Kind != "news" {
		t.Fatalf("list item kind=%q want=news", list.Items[0].Kind)
	}
}

// admin以外では/items作成できないことを確認。
func Test_Items_Create_Forbidden_ForNonAdmin(t *testing.T) {
	env := newTestEnv(t)
	env.resetDB(t)

	u := env.createUserDirect(t, "user-create@example.com", testUserPW, "user", true)
	st := env.loginOK(t, u.Email, testUserPW)

	publishedAt := time.Now().UTC().Format(time.RFC3339)

	itemBody := createItemReq{
		Title:       "Should Fail",
		Kind:        "news",
		SourceID:    1,
		PublishedAt: publishedAt,
	}

	res, body := postJSON(
		t,
		env,
		"/items",
		&itemBody,
		bearerHeader(st.AccessToken),
	)
	mustStatus(t, res, body, http.StatusForbidden)
	mustErrCode(t, body, "forbidden")
}
