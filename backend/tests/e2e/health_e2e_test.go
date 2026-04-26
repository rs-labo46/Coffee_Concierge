package e2e

import "testing"

// APIサーバーが起動していて、/health が正常応答するかを確認
func TestE2E_Health_ReturnsOK(t *testing.T) {
	c := newAPIClient(t)

	status, body, _ := c.getJSON(t, "/health", nil)
	requireStatus(t, status, 200, body)

	var res struct {
		Status string `json:"status"`
	}
	decodeJSON(t, body, &res)

	if res.Status != "ok" {
		t.Fatalf("unexpected health status: %s", res.Status)
	}
}
