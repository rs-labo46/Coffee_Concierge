package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"testing"
	"time"
)

const defaultBaseURL = "http://127.0.0.1:18080"

type apiClient struct {
	base string
	hc   *http.Client
}

// E2E用のHTTPクライアントを作る。
func newAPIClient(t *testing.T) *apiClient {
	t.Helper()

	if os.Getenv("RUN_E2E") != "1" {
		t.Skip("E2E is skipped. set RUN_E2E=1 and start the API server")
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}

	base := strings.TrimRight(os.Getenv("BASE_URL"), "/")
	if base == "" {
		base = defaultBaseURL
	}

	return &apiClient{
		base: base,
		hc: &http.Client{
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
	}
}

// JSONリクエストを送り、レスポンス本文を返す。
func (c *apiClient) doJSON(t *testing.T, method string, path string, body interface{}, headers map[string]string) (int, []byte, http.Header) {
	t.Helper()

	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.base+path, r)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := c.hc.Do(req)
	if err != nil {
		t.Fatalf("do request %s %s: %v", method, path, err)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return res.StatusCode, b, res.Header
}

// GETリクエストを送る。
func (c *apiClient) getJSON(t *testing.T, path string, headers map[string]string) (int, []byte, http.Header) {
	t.Helper()
	return c.doJSON(t, http.MethodGet, path, nil, headers)
}

// レスポンス本文を指定structへ。
func decodeJSON(t *testing.T, b []byte, out interface{}) {
	t.Helper()
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("decode json: %v body=%s", err, string(b))
	}
}

// HTTPステータスが期待値と一致するか検証する。
func requireStatus(t *testing.T, got int, want int, body []byte) {
	t.Helper()
	if got != want {
		t.Fatalf("unexpected status: got=%d want=%d body=%s", got, want, string(body))
	}
}

// Authorizationヘッダを作る。
func bearer(token string) map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
}

// seed管理者の認証情報(email/password)を返す。
func adminCreds() (string, string) {
	email := strings.TrimSpace(os.Getenv("E2E_ADMIN_EMAIL"))
	if email == "" {
		email = "admin@test.com"
	}

	password := strings.TrimSpace(os.Getenv("E2E_ADMIN_PASSWORD"))
	if password == "" {
		password = "AdminPass123!"
	}

	return email, password
}
