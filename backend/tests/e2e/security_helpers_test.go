package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// e2e45ErrRes は共通エラーレスポンスの最小検証用DTO。
type e2e45ErrRes struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

var e2e45AdminTokenCache struct {
	mu    sync.Mutex
	token string
}

// e2e45RequireErrorCode は {error:string} のerror codeを検証する。
func e2e45RequireErrorCode(t *testing.T, body []byte, want string) {
	t.Helper()

	var got e2e45ErrRes
	decodeJSON(t, body, &got)
	if got.Error != want {
		t.Fatalf("unexpected error code: got=%s want=%s body=%s", got.Error, want, string(body))
	}
}

// e2e45GetCSRF は /auth/csrf を実行し、bodyのCSRF tokenを返す。
// 同時にcookie jarへcsrf_token cookieが保存される。
func e2e45GetCSRF(t *testing.T, c *apiClient) string {
	t.Helper()

	status, body, _ := c.getJSON(t, "/auth/csrf", nil)
	requireStatus(t, status, http.StatusOK, body)

	var res csrfRes
	decodeJSON(t, body, &res)
	if res.Token == "" {
		t.Fatalf("csrf token is empty")
	}

	return res.Token
}

// e2e45LoginAdmin はseed済みadminでログインし、access tokenを返す。
// refresh_token cookieはclientのcookie jarへ保存される。
func e2e45LoginAdmin(t *testing.T, c *apiClient) string {
	t.Helper()

	token, status, body := e2e45LoginAdminRaw(t, c)
	requireStatus(t, status, http.StatusOK, body)
	if token == "" {
		t.Fatalf("access token is empty")
	}

	return token
}

// e2e45LoginAdminRaw はadmin loginを実行し、status/bodyも返す。
// token cache作成時に、失敗理由をそのまま表示するために使う。
func e2e45LoginAdminRaw(t *testing.T, c *apiClient) (string, int, []byte) {
	t.Helper()

	email, password := adminCreds()
	status, body, _ := c.doJSON(t, http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, map[string]string{
		"X-Forwarded-For": e2e45UniqueIP(),
	})
	if status != http.StatusOK {
		return "", status, body
	}

	var res authRes
	decodeJSON(t, body, &res)
	return res.AccessToken, status, body
}

// e2e45AdminAccessToken はprivate/admin系E2Eで使うadmin tokenを共有する。
// 各テストで毎回loginするとlogin RateLimitに巻き込まれるため、Bearerだけ必要なテストではこの関数を使う。
func e2e45AdminAccessToken(t *testing.T) string {
	t.Helper()

	e2e45AdminTokenCache.mu.Lock()
	defer e2e45AdminTokenCache.mu.Unlock()

	if e2e45AdminTokenCache.token != "" {
		return e2e45AdminTokenCache.token
	}

	c := newAPIClient(t)
	e2e45GetCSRF(t, c)
	token, status, body := e2e45LoginAdminRaw(t, c)
	requireStatus(t, status, http.StatusOK, body)
	if token == "" {
		t.Fatalf("admin token is empty body=%s", string(body))
	}

	e2e45AdminTokenCache.token = token
	return token
}

// e2e45Cookie は /auth 配下に送信されるcookieをcookie jarから取得する。
func e2e45Cookie(t *testing.T, c *apiClient, name string) *http.Cookie {
	t.Helper()
	return e2e45CookieForPath(t, c, "/auth/refresh", name)
}

// e2e45CookieForPath は指定pathに送信されるcookieをcookie jarから取得する。
func e2e45CookieForPath(t *testing.T, c *apiClient, path string, name string) *http.Cookie {
	t.Helper()

	u, err := url.Parse(c.base + path)
	if err != nil {
		t.Fatalf("parse cookie url: %v", err)
	}

	for _, ck := range c.hc.Jar.Cookies(u) {
		if ck.Name == name {
			copied := *ck
			return &copied
		}
	}

	t.Fatalf("cookie not found: %s", name)
	return nil
}

// e2e45CookieExists は /auth 配下へ送信される指定cookieがcookie jarに残っているかを返す。
func e2e45CookieExists(t *testing.T, c *apiClient, name string) bool {
	t.Helper()

	u, err := url.Parse(c.base + "/auth/refresh")
	if err != nil {
		t.Fatalf("parse cookie url: %v", err)
	}

	for _, ck := range c.hc.Jar.Cookies(u) {
		if ck.Name == name && ck.Value != "" {
			return true
		}
	}
	return false
}

// e2e45SetCookie は別clientにcookieを移す。
// refresh reuseのように「古いcookieを再利用する」検証で使う。
func e2e45SetCookie(t *testing.T, c *apiClient, ck *http.Cookie) {
	t.Helper()

	u, err := url.Parse(c.base + "/auth/refresh")
	if err != nil {
		t.Fatalf("parse cookie url: %v", err)
	}

	copied := *ck
	c.hc.Jar.SetCookies(u, []*http.Cookie{&copied})
}

// e2e45StartGuestSession はguest sessionを作成する。
func e2e45StartGuestSession(t *testing.T, c *apiClient, title string) startSessionRes {
	t.Helper()

	status, body, _ := c.doJSON(t, http.MethodPost, "/search/sessions", map[string]string{
		"title": title,
	}, nil)
	requireStatus(t, status, http.StatusCreated, body)

	var res startSessionRes
	decodeJSON(t, body, &res)
	if res.Session.ID == 0 {
		t.Fatalf("session id is zero")
	}
	if res.SessionKey == "" {
		t.Fatalf("session key is empty")
	}
	return res
}

// e2e45ValidPrefBody は検索条件の正常bodyを返す。
func e2e45ValidPrefBody(note string) map[string]interface{} {
	return map[string]interface{}{
		"flavor":     3,
		"acidity":    2,
		"bitterness": 2,
		"body":       3,
		"aroma":      4,
		"mood":       "morning",
		"method":     "drip",
		"scene":      "break",
		"temp_pref":  "hot",
		"excludes":   []string{},
		"note":       note,
	}
}

// e2e45SetGuestPref はguest sessionへ初期prefを設定する。
func e2e45SetGuestPref(t *testing.T, c *apiClient, started startSessionRes) {
	t.Helper()

	status, body, _ := c.doJSON(t, http.MethodPost, pathForSessionPref(started.Session.ID), e2e45ValidPrefBody("E2E45 initial pref"), map[string]string{
		"X-Session-Key": started.SessionKey,
	})
	requireStatus(t, status, http.StatusOK, body)
}

// e2e45ExpireGuestSession はDBを直接更新し、guest sessionを期限切れ状態にする。
// 24時間待つE2Eは現実的ではないため、期限切れ検証ではDB状態を明示的に作る。
func e2e45ExpireGuestSession(t *testing.T, sessionID uint) {
	t.Helper()

	db, err := sql.Open("postgres", e2e45PostgresDSN())
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Skipf("E2E_DATABASE_DSN is not reachable: %v", err)
	}

	var tableName sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT to_regclass('public.sessions')").Scan(&tableName); err != nil {
		t.Fatalf("check sessions table: %v", err)
	}
	if !tableName.Valid || tableName.String == "" {
		t.Skip("sessions table does not exist on E2E_DATABASE_DSN; set E2E_DATABASE_DSN to the same DB used by API")
	}

	res, err := db.ExecContext(ctx, "UPDATE sessions SET guest_expires_at = NOW() - INTERVAL '1 hour' WHERE id = $1", sessionID)
	if err != nil {
		t.Fatalf("expire guest session: %v", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("rows affected: %v", err)
	}
	if affected != 1 {
		t.Fatalf("unexpected rows affected: got=%d want=1", affected)
	}
}

// e2e45PostgresDSN はE2EからDBへ直接接続するためのDSN。
func e2e45PostgresDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("E2E_DATABASE_DSN")); dsn != "" {
		return dsn
	}

	user := e2e45Env("POSTGRES_USER", "myuser")
	password := e2e45Env("POSTGRES_PASSWORD", "mypassword")
	dbName := e2e45Env("POSTGRES_DB", "mydb")
	host := e2e45Env("POSTGRES_HOST", "localhost")
	port := e2e45EnvInt("POSTGRES_PORT", 5433)

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
}

// e2e45RequireRateLimitE2E はRateLimit専用E2Eだけを明示実行に分離する。
func e2e45RequireRateLimitE2E(t *testing.T) {
	t.Helper()

	if os.Getenv("RUN_RATE_LIMIT_E2E") != "1" {
		t.Skip("RateLimit E2E is skipped. set RUN_RATE_LIMIT_E2E=1 to run it")
	}
}

// e2e45FlushRedis はRateLimit用Redisを初期化する。
// ローカルE2E専用。失敗した場合は該当テストをskipする。
func e2e45FlushRedis(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	addr := e2e45Env("E2E_REDIS_ADDR", "")
	if addr == "" {
		addr = e2e45Env("REDIS_HOST", e2e45Env("E2E_REDIS_HOST", "localhost")) + ":" + e2e45Env("REDIS_PORT", e2e45Env("E2E_REDIS_PORT", "6379"))
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: e2e45Env("REDIS_PASSWORD", ""),
		DB:       e2e45EnvInt("REDIS_DB", 0),
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis is not reachable for rate limit E2E: %v", err)
	}
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Skipf("redis FLUSHDB failed for rate limit E2E: %v", err)
	}
}

// e2e45ExpectRateLimited repeatedly sends a request until 429 is observed.
func e2e45ExpectRateLimited(t *testing.T, maxAttempts int, fn func() (int, []byte, http.Header)) {
	t.Helper()

	for i := 0; i < maxAttempts; i++ {
		status, body, header := fn()
		if status == http.StatusTooManyRequests {
			e2e45RequireErrorCode(t, body, "rate_limited")
			if header.Get("Retry-After") == "" {
				t.Fatalf("Retry-After header is empty body=%s", string(body))
			}
			return
		}
	}

	t.Fatalf("rate limit did not return 429 within %d attempts", maxAttempts)
}

func e2e45Env(key string, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func e2e45EnvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func e2e45UniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@example.com", prefix, time.Now().UnixNano())
}

func e2e45UniqueIP() string {
	n := time.Now().UnixNano()
	return fmt.Sprintf("203.0.113.%d", (n%200)+1)
}
