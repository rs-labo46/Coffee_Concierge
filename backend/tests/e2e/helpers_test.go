package e2e

import (
	"bytes"
	"coffee-spa/config"
	appdb "coffee-spa/db"
	"coffee-spa/entity"
	"coffee-spa/usecase"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	defaultBaseURL = "http://127.0.0.1:8080"
	testUserPW     = "CorrectPW123!"
	testAdminPW    = "AdminPW123!@"
)

// request body structs
type signupReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyEmailReq struct {
	Token string `json:"token"`
}

type forgotPwReq struct {
	Email string `json:"email"`
}

type resetPwReq struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type createSourceReq struct {
	Name    string  `json:"name"`
	SiteURL *string `json:"site_url,omitempty"`
}

type createItemReq struct {
	Title       string  `json:"title"`
	Summary     *string `json:"summary,omitempty"`
	URL         *string `json:"url,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Kind        string  `json:"kind"`
	SourceID    int64   `json:"source_id"`
	PublishedAt string  `json:"published_at"`
}

// E2Ehelperで送信可能なJSON body型だけに制限。
type jsonReq interface {
	signupReq |
		loginReq |
		verifyEmailReq |
		forgotPwReq |
		resetPwReq |
		createSourceReq |
		createItemReq
}

// response structs
type apiErrRes struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type authRes struct {
	AccessToken string `json:"access_token"`
	User        struct {
		ID            int64  `json:"id"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		TokenVer      int    `json:"token_ver"`
		EmailVerified bool   `json:"email_verified"`
	} `json:"user"`
}

type signupRes struct {
	User struct {
		ID            int64  `json:"id"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		TokenVer      int    `json:"token_ver"`
		EmailVerified bool   `json:"email_verified"`
	} `json:"user"`
}

type meRes struct {
	User struct {
		ID            int64  `json:"id"`
		Email         string `json:"email"`
		Role          string `json:"role"`
		TokenVer      int    `json:"token_ver"`
		EmailVerified bool   `json:"email_verified"`
	} `json:"user"`
}

type sourceRes struct {
	Source entity.Source `json:"source"`
}

type itemRes struct {
	Item entity.Item `json:"item"`
}

type itemListRes struct {
	Items []entity.Item `json:"items"`
}

type topItemsRes struct {
	News   []entity.Item `json:"news"`
	Recipe []entity.Item `json:"recipe"`
	Deal   []entity.Item `json:"deal"`
	Shop   []entity.Item `json:"shop"`
}

// login / refresh 後に必要な認証情報をまとめる。
type authState struct {
	AccessToken string
	Rt          string
	CSRFToken   string
	UserID      int64
	Role        string
}

// E2E共通の接続先やDBをまとめる。
type testEnv struct {
	BaseURL string
	DB      *gorm.DB
	Client  *http.Client
	RDB     *redis.Client
}

func redisAddr() string {
	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr != "" {
		return addr
	}

	host := strings.TrimSpace(os.Getenv("REDIS_HOST"))
	if host == "" {
		host = "localhost"
	}

	port := strings.TrimSpace(os.Getenv("REDIS_PORT"))
	if port == "" {
		port = "6379"
	}

	return host + ":" + port
}

func (e *testEnv) resetRedis(t *testing.T) {
	t.Helper()

	if e.RDB == nil {
		t.Fatal("redis client is nil")
	}

	if err := e.RDB.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("reset redis failed: %v", err)
	}
}

// .env を複数候補から読み込み、E2E 用の接続を作る。
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	_ = godotenv.Load("../../../.env")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load error: %v", err)
	}

	d, err := appdb.Open(cfg)
	if err != nil {
		t.Fatalf("db.Open error: %v", err)
	}

	if err := appdb.Migrate(d); err != nil {
		t.Fatalf("db.Migrate error: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr(),
	})

	baseURL := strings.TrimSpace(os.Getenv("BASE_URL"))
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return &testEnv{
		BaseURL: strings.TrimRight(baseURL, "/"),
		DB:      d.G,
		Client:  &http.Client{Timeout: 10 * time.Second},
		RDB:     rdb,
	}
}

// テスト前に主要テーブルを削除。
func (e *testEnv) resetDB(t *testing.T) {
	t.Helper()

	tables := []string{
		"audit_logs",
		"items",
		"sources",
		"refresh_tokens",
		"pw_resets",
		"email_verifies",
		"users",
	}

	exists := make([]string, 0, len(tables))
	for _, table := range tables {
		if e.DB.Migrator().HasTable(table) {
			exists = append(exists, table)
		}
	}

	if len(exists) == 0 {
		t.Fatal("reset DB failed: no tables found")
	}

	sql := "TRUNCATE TABLE " + strings.Join(exists, ", ") + " RESTART IDENTITY CASCADE"
	if err := e.DB.Exec(sql).Error; err != nil {
		t.Fatalf("reset DB failed: %v", err)
	}

	if e.RDB != nil {
		if err := e.RDB.FlushDB(context.Background()).Err(); err != nil {
			t.Fatalf("reset redis failed: %v", err)
		}
	}
}

// JSONリクエスト送信の共通関数。
func doJSON[T jsonReq](
	t *testing.T,
	client *http.Client,
	baseURL string,
	method string,
	path string,
	body *T,
	headers http.Header,
	cookies ...*http.Cookie,
) (*http.Response, []byte) {
	t.Helper()

	var r io.Reader

	if body != nil {
		b, err := json.Marshal(*body)
		if err != nil {
			t.Fatalf("json.Marshal error: %v", err)
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, baseURL+path, r)
	if err != nil {
		t.Fatalf("http.NewRequest error: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, vs := range headers {
		cp := make([]string, len(vs))
		copy(cp, vs)
		req.Header[k] = cp
	}

	for _, c := range cookies {
		if c != nil {
			req.AddCookie(c)
		}
	}

	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("client.Do error: %v", err)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		_ = res.Body.Close()
		t.Fatalf("io.ReadAll error: %v", err)
	}
	_ = res.Body.Close()

	return res, b
}

// bodyなしのhelper。
func (e *testEnv) doEmpty(
	t *testing.T,
	method string,
	path string,
	headers http.Header,
	cookies ...*http.Cookie,
) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(method, e.BaseURL+path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest error: %v", err)
	}

	for k, vs := range headers {
		cp := make([]string, len(vs))
		copy(cp, vs)
		req.Header[k] = cp
	}

	for _, c := range cookies {
		if c != nil {
			req.AddCookie(c)
		}
	}

	res, err := e.Client.Do(req)
	if err != nil {
		t.Fatalf("client.Do error: %v", err)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		_ = res.Body.Close()
		t.Fatalf("io.ReadAll error: %v", err)
	}
	_ = res.Body.Close()

	return res, b
}

// POST JSON 用の helper。
func postJSON[T jsonReq](
	t *testing.T,
	env *testEnv,
	path string,
	body *T,
	headers http.Header,
	cookies ...*http.Cookie,
) (*http.Response, []byte) {
	t.Helper()

	return doJSON(
		t,
		env.Client,
		env.BaseURL,
		http.MethodPost,
		path,
		body,
		headers,
		cookies...,
	)
}

// Authorizationヘッダを組み立てる。
func bearerHeader(accessToken string) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+accessToken)
	return h
}

// csrf header を足す。
func addCSRFHeader(h http.Header, csrf string) http.Header {
	out := http.Header{}

	for k, vs := range h {
		cp := make([]string, len(vs))
		copy(cp, vs)
		out[k] = cp
	}

	out.Set("X-CSRF-Token", csrf)
	return out
}

// verify/reset tokenをDBに入れる。
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// DBに直接userを作る。
func (e *testEnv) createUserDirect(
	t *testing.T,
	email string,
	rawPW string,
	role string,
	emailVerified bool,
) entity.User {
	t.Helper()

	hasher := usecase.NewBcryptHasher()
	hash, err := hasher.Hash(rawPW)
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	u := entity.User{
		Email:         strings.ToLower(strings.TrimSpace(email)),
		PassHash:      hash,
		Role:          role,
		TokenVer:      1,
		EmailVerified: emailVerified,
	}

	if err := e.DB.Create(&u).Error; err != nil {
		t.Fatalf("create user direct error: %v", err)
	}

	return u
}

// signup後のuserをDBから引く。
func (e *testEnv) findUserByEmail(t *testing.T, email string) entity.User {
	t.Helper()

	var u entity.User
	err := e.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(email))).First(&u).Error
	if err != nil {
		t.Fatalf("find user by email error: %v", err)
	}

	return u
}

// E2E用に既知tokenをDBに直接投入。
func (e *testEnv) insertVerifyTokenDirect(t *testing.T, userID int64, rawToken string) {
	t.Helper()

	ev := entity.EmailVerify{
		UserID:    userID,
		TokenHash: sha256Hex(rawToken),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if err := e.DB.Create(&ev).Error; err != nil {
		t.Fatalf("insert verify token error: %v", err)
	}
}

// reset token をDBに直接投入。
func (e *testEnv) insertResetTokenDirect(t *testing.T, userID int64, rawToken string) {
	t.Helper()

	pw := entity.PwReset{
		UserID:    userID,
		TokenHash: sha256Hex(rawToken),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	if err := e.DB.Create(&pw).Error; err != nil {
		t.Fatalf("insert reset token error: %v", err)
	}
}

// login成功とcookie取得をまとめたhelper。
func (e *testEnv) loginOK(t *testing.T, email string, pw string) authState {
	t.Helper()

	req := loginReq{
		Email:    email,
		Password: pw,
	}

	res, body := postJSON(t, e, "/auth/login", &req, nil)
	mustStatus(t, res, body, http.StatusOK)

	var out authRes
	mustJSON(t, body, &out)

	refreshCookie := mustCookie(t, res, "refresh_token")
	csrfCookie := mustCookie(t, res, "csrf_token")

	return authState{
		AccessToken: out.AccessToken,
		Rt:          refreshCookie.Value,
		CSRFToken:   csrfCookie.Value,
		UserID:      out.User.ID,
		Role:        out.User.Role,
	}
}

// adminユーザーを直接作ってlogin
func (e *testEnv) createAdminAndLogin(t *testing.T) authState {
	t.Helper()

	email := "admin@example.com"
	_ = e.createUserDirect(t, email, testAdminPW, string(entity.RoleAdmin), true)

	return e.loginOK(t, email, testAdminPW)
}

// refresh / csrf cookieを組み立てる。
func authCookies(st authState) []*http.Cookie {
	return []*http.Cookie{
		{
			Name:  "refresh_token",
			Value: st.Rt,
			Path:  "/auth",
		},
		{
			Name:  "csrf_token",
			Value: st.CSRFToken,
			Path:  "/",
		},
	}
}

// 期待statusであることを確認。
func mustStatus(t *testing.T, res *http.Response, body []byte, want int) {
	t.Helper()

	if res.StatusCode != want {
		t.Fatalf("status=%d want=%d body=%s", res.StatusCode, want, string(body))
	}
}

// JSONをstructに流し込む。
func mustJSON[T any](t *testing.T, body []byte, out *T) {
	t.Helper()

	if err := json.Unmarshal(body, out); err != nil {
		t.Fatalf("json.Unmarshal error: %v body=%s", err, string(body))
	}
}

// Set-Cookieから指定cookieを取る。
func mustCookie(t *testing.T, res *http.Response, name string) *http.Cookie {
	t.Helper()

	for _, c := range res.Cookies() {
		if c.Name == name {
			return c
		}
	}

	t.Fatalf("cookie %q not found", name)
	return nil
}

// error JSONのerrorを確認。
func mustErrCode(t *testing.T, body []byte, want string) {
	t.Helper()

	var out apiErrRes
	mustJSON(t, body, &out)

	if out.Error != want {
		t.Fatalf("error=%q want=%q body=%s", out.Error, want, string(body))
	}
}

// /me などのno-storeを確認。
func mustNoStore(t *testing.T, res *http.Response) {
	t.Helper()

	if got := res.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control=%q want=no-store", got)
	}
}

// revoke 済みの数確認用。
func (e *testEnv) countRevokedRtsByUser(t *testing.T, userID int64) int64 {
	t.Helper()

	var n int64
	if err := e.DB.Model(&entity.Rt{}).
		Where("user_id = ? AND revoked_at IS NOT NULL", userID).
		Count(&n).Error; err != nil {
		t.Fatalf("count revoked refresh tokens error: %v", err)
	}

	return n
}
