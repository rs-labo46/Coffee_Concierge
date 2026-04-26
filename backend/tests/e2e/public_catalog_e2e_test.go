package e2e

import (
	"net/http"
	"testing"
)

// 認証なしで取得できる公開カタログ系APIを確認
func TestE2E_PublicCatalog_TopAndSourcesAreReachable(t *testing.T) {
	c := newAPIClient(t)
	//news / recipe / deal / shopなどのItemが存在するか
	topStatus, topBody, _ := c.getJSON(t, "/items/top?limit=3", nil)
	requireStatus(t, topStatus, http.StatusOK, topBody)

	var top struct {
		News   []map[string]interface{} `json:"news"`
		Recipe []map[string]interface{} `json:"recipe"`
		Deal   []map[string]interface{} `json:"deal"`
		Shop   []map[string]interface{} `json:"shop"`
	}
	decodeJSON(t, topBody, &top)
	//Sourceが登録されているか
	sourceStatus, sourceBody, _ := c.getJSON(t, "/sources?limit=5&offset=0", nil)
	requireStatus(t, sourceStatus, http.StatusOK, sourceBody)

	var sources struct {
		Sources []map[string]interface{} `json:"sources"`
	}
	decodeJSON(t, sourceBody, &sources)
}
