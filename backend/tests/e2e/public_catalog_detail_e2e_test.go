package e2e

import (
	"net/http"
	"testing"
)

// TestE2E45_Public_ItemsList_ReturnsOK は、公開Item一覧が認証なしで取得できることを検証する。
func TestE2E45_Public_ItemsList_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/items?limit=20&offset=0", nil)
	requireStatus(t, status, http.StatusOK, body)

	var res struct {
		Items []map[string]interface{} `json:"items"`
	}
	decodeJSON(t, body, &res)
}

// TestE2E45_Public_ItemsTop_ReturnsOK は、Topページ用Itemが取得できることを検証する。
func TestE2E45_Public_ItemsTop_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/items/top?limit=3", nil)
	requireStatus(t, status, http.StatusOK, body)

	var res struct {
		News   []map[string]interface{} `json:"news"`
		Recipe []map[string]interface{} `json:"recipe"`
		Deal   []map[string]interface{} `json:"deal"`
		Shop   []map[string]interface{} `json:"shop"`
	}
	decodeJSON(t, body, &res)
}

// TestE2E45_Public_SourcesList_ReturnsOK は、公開Source一覧が取得できることを検証する。
func TestE2E45_Public_SourcesList_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/sources?limit=20&offset=0", nil)
	requireStatus(t, status, http.StatusOK, body)

	var res struct {
		Sources []map[string]interface{} `json:"sources"`
	}
	decodeJSON(t, body, &res)
}

// TestE2E45_Public_SourceDetail_ReturnsOK は、一覧から取得したSourceの詳細が取得できることを検証する。
func TestE2E45_Public_SourceDetail_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/sources?limit=1&offset=0", nil)
	requireStatus(t, status, http.StatusOK, body)

	var list struct {
		Sources []struct {
			ID uint `json:"id"`
		} `json:"sources"`
	}
	decodeJSON(t, body, &list)
	if len(list.Sources) == 0 {
		t.Skip("no source seed exists")
	}

	status, body, _ = c.getJSON(t, "/sources/"+uintToString(list.Sources[0].ID), nil)
	requireStatus(t, status, http.StatusOK, body)
}

// TestE2E45_Public_ItemDetail_ReturnsOK は、一覧から取得したItemの詳細が取得できることを検証する。
func TestE2E45_Public_ItemDetail_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/items?limit=1&offset=0", nil)
	requireStatus(t, status, http.StatusOK, body)

	var list struct {
		Items []struct {
			ID uint `json:"id"`
		} `json:"items"`
	}
	decodeJSON(t, body, &list)
	if len(list.Items) == 0 {
		t.Skip("no item seed exists")
	}

	status, body, _ = c.getJSON(t, "/items/"+uintToString(list.Items[0].ID), nil)
	requireStatus(t, status, http.StatusOK, body)
}

// TestE2E45_Public_BeansList_ReturnsOK は、公開Bean一覧が取得できることを検証する。
func TestE2E45_Public_BeansList_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/beans?active=true&limit=20&offset=0", nil)
	requireStatus(t, status, http.StatusOK, body)

	var res struct {
		Beans []map[string]interface{} `json:"beans"`
	}
	decodeJSON(t, body, &res)
}

// TestE2E45_Public_BeanDetail_ReturnsOK は、一覧から取得したBeanの詳細が取得できることを検証する。
func TestE2E45_Public_BeanDetail_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/beans?active=true&limit=1&offset=0", nil)
	requireStatus(t, status, http.StatusOK, body)

	var list struct {
		Beans []struct {
			ID uint `json:"id"`
		} `json:"beans"`
	}
	decodeJSON(t, body, &list)
	if len(list.Beans) == 0 {
		t.Skip("no bean seed exists")
	}

	status, body, _ = c.getJSON(t, "/beans/"+uintToString(list.Beans[0].ID), nil)
	requireStatus(t, status, http.StatusOK, body)
}

// TestE2E45_Public_RecipesList_ReturnsOK は、公開Recipe一覧が取得できることを検証する。
func TestE2E45_Public_RecipesList_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/recipes?active=true&limit=20&offset=0", nil)
	requireStatus(t, status, http.StatusOK, body)

	var res struct {
		Recipes []map[string]interface{} `json:"recipes"`
	}
	decodeJSON(t, body, &res)
}

// TestE2E45_Public_RecipeDetail_ReturnsOK は、一覧から取得したRecipeの詳細が取得できることを検証する。
func TestE2E45_Public_RecipeDetail_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.getJSON(t, "/recipes?active=true&limit=1&offset=0", nil)
	requireStatus(t, status, http.StatusOK, body)

	var list struct {
		Recipes []struct {
			ID uint `json:"id"`
		} `json:"recipes"`
	}
	decodeJSON(t, body, &list)
	if len(list.Recipes) == 0 {
		t.Skip("no recipe seed exists")
	}

	status, body, _ = c.getJSON(t, "/recipes/"+uintToString(list.Recipes[0].ID), nil)
	requireStatus(t, status, http.StatusOK, body)
}
