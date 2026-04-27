package e2e

import (
	"net/http"
	"testing"
	"time"
)

func TestE2E_AdminContent_CreateSourceItemBeanRecipe(t *testing.T) {
	c := newAPIClient(t)
	token := e2e45AdminAccessToken(t)
	auth := bearer(token)

	status, body, _ := c.doJSON(t, http.MethodPost, "/admin/sources", map[string]string{
		"name":     "E2E Source",
		"site_url": "https://example.com/e2e-source",
	}, auth)
	requireStatus(t, status, http.StatusCreated, body)
	var srcRes struct {
		Source struct {
			ID uint `json:"id"`
		} `json:"source"`
	}
	decodeJSON(t, body, &srcRes)
	if srcRes.Source.ID == 0 {
		t.Fatalf("source id is zero")
	}

	status, body, _ = c.doJSON(t, http.MethodPost, "/admin/items", map[string]interface{}{
		"title":        "E2E Item",
		"summary":      "E2E summary",
		"url":          "https://example.com/e2e-item",
		"image_url":    "https://example.com/e2e-item.png",
		"kind":         "news",
		"source_id":    srcRes.Source.ID,
		"published_at": time.Now().UTC().Format(time.RFC3339),
	}, auth)
	requireStatus(t, status, http.StatusCreated, body)

	status, body, _ = c.doJSON(t, http.MethodPost, "/admin/beans", map[string]interface{}{
		"name":       "E2E Bean",
		"roast":      "medium",
		"origin":     "Brazil",
		"flavor":     3,
		"acidity":    3,
		"bitterness": 3,
		"body":       3,
		"aroma":      3,
		"desc":       "E2E bean",
		"buy_url":    "https://example.com/e2e-bean",
		"active":     true,
	}, auth)
	requireStatus(t, status, http.StatusCreated, body)
	var beanRes struct {
		Bean struct {
			ID uint `json:"id"`
		} `json:"bean"`
	}
	decodeJSON(t, body, &beanRes)
	if beanRes.Bean.ID == 0 {
		t.Fatalf("bean id is zero")
	}

	status, body, _ = c.doJSON(t, http.MethodPost, "/admin/recipes", map[string]interface{}{
		"bean_id":   beanRes.Bean.ID,
		"name":      "E2E Recipe",
		"method":    "drip",
		"temp_pref": "hot",
		"grind":     "medium",
		"ratio":     "1:15",
		"temp":      90,
		"time_sec":  180,
		"steps":     []string{"brew"},
		"desc":      "E2E recipe",
		"active":    true,
	}, auth)
	requireStatus(t, status, http.StatusCreated, body)
}

func TestE2E_AdminContent_UserDenied(t *testing.T) {
	c := newAPIClient(t)
	status, body, _ := c.doJSON(t, http.MethodPost, "/admin/beans", map[string]interface{}{"name": "x"}, nil)
	requireStatus(t, status, http.StatusUnauthorized, body)
}
