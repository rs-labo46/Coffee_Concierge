package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"coffee-spa/entity"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func newTestContext(method string, path string, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, http.NoBody)
	if body != "" {
		req = httptest.NewRequest(method, path, stringsReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func stringsReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

func okHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status mismatch: want=%d got=%d body=%s", want, rec.Code, rec.Body.String())
	}
}

func makeAccessToken(t *testing.T, secret string, userID uint, role entity.Role, tv int, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": strconv.FormatUint(uint64(userID), 10),
		"role": string(role),
		"tv": tv,
		"exp": exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign jwt: %v", err)
	}
	return signed
}

type tvReaderMock struct {
	user *entity.User
	err  error
	ids  []uint
}

func (m *tvReaderMock) GetByID(id uint) (*entity.User, error) {
	m.ids = append(m.ids, id)
	if m.err != nil {
		return nil, m.err
	}
	if m.user == nil {
		return nil, errors.New("not found")
	}
	return m.user, nil
}

type allRateLimiterMock struct {
	allow bool
	retry int
	err   error

	signupIPs []string
	loginIPs  []string
	loginKeys []string
	resendIPs []string
	resendKey []string
	forgotIPs []string
	forgotKey []string
	refreshes []string
	wsKeys    []string
}

func (m *allRateLimiterMock) AllowSignup(ip string) (bool, int, error) {
	m.signupIPs = append(m.signupIPs, ip)
	return m.allow, m.retry, m.err
}
func (m *allRateLimiterMock) AllowLoginIP(ip string) (bool, int, error) {
	m.loginIPs = append(m.loginIPs, ip)
	return m.allow, m.retry, m.err
}
func (m *allRateLimiterMock) AllowLogin(emailHash string) (bool, int, error) {
	m.loginKeys = append(m.loginKeys, emailHash)
	return m.allow, m.retry, m.err
}
func (m *allRateLimiterMock) AllowResendIP(ip string) (bool, int, error) {
	m.resendIPs = append(m.resendIPs, ip)
	return m.allow, m.retry, m.err
}
func (m *allRateLimiterMock) AllowResendMail(emailHash string) (bool, int, error) {
	m.resendKey = append(m.resendKey, emailHash)
	return m.allow, m.retry, m.err
}
func (m *allRateLimiterMock) AllowForgotIP(ip string) (bool, int, error) {
	m.forgotIPs = append(m.forgotIPs, ip)
	return m.allow, m.retry, m.err
}
func (m *allRateLimiterMock) AllowForgotMail(emailHash string) (bool, int, error) {
	m.forgotKey = append(m.forgotKey, emailHash)
	return m.allow, m.retry, m.err
}
func (m *allRateLimiterMock) AllowRefreshToken(tokenHash string) (bool, int, error) {
	m.refreshes = append(m.refreshes, tokenHash)
	return m.allow, m.retry, m.err
}
func (m *allRateLimiterMock) AllowWS(key string) (bool, int, error) {
	m.wsKeys = append(m.wsKeys, key)
	return m.allow, m.retry, m.err
}
