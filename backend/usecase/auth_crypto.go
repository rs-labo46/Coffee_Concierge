package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"coffee-spa/entity"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// bcryptを使うパスワードハッシュ実装。
type BcryptHasher struct{}

// BcryptHasherを作る。
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

// 平文パスワードをbcrypt hashに。
func (h *BcryptHasher) Hash(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// hashと平文パスワードを照合。
func (h *BcryptHasher) Compare(hash string, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

// JWT/csrf/opaque tokenを作る。
// TokenSvc/RefreshSvcで必要なメソッドも持たせる。
type JWTMaker struct {
	secret []byte
}

// access tokenのclaims。
type accessClaims struct {
	Role string `json:"role"`
	TV   int    `json:"tv"`
	jwt.RegisteredClaims
}

// JWTMakerを作る。
func NewJWTMaker(secret string) *JWTMaker {
	return &JWTMaker{secret: []byte(secret)}
}

// access tokenを作る。
func (m *JWTMaker) NewAccess(userID int64, role string, tokenVer int) (string, error) {
	now := time.Now()

	claims := accessClaims{
		Role: role,
		TV:   tokenVer,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(3 * time.Minute)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	s, err := t.SignedString(m.secret)
	if err != nil {
		return "", err
	}

	return s, nil
}

// AuthUsecaseのTokenSvc interfaceを満たす。
// entity.Userを受け取り、access tokenを発行。
func (m *JWTMaker) SignAccess(user entity.User) (string, error) {
	return m.NewAccess(int64(user.ID), string(user.Role), user.TokenVer)
}

// csrf tokenを作る。
func (m *JWTMaker) NewCSRF() (string, error) {
	return newHex(32)
}

// refresh/verify/reset用のランダムtokenを作る。
func (m *JWTMaker) NewOpaque() (string, error) {
	return newHex(32)
}

// refresh token family idを作る。
func (m *JWTMaker) NewFamilyID() (string, error) {
	return newHex(16)
}

// AuthUsecaseのRefreshSvc interfaceを満たす。
// 平文refresh tokenと、そのhashをセットで返す。
func (m *JWTMaker) New() (plain string, tokenHash string, err error) {
	plain, err = m.NewOpaque()
	if err != nil {
		return "", "", err
	}
	return plain, hashOpaque(plain), nil
}

// AuthUsecaseのRefreshSvc interfaceを満たす。
// 平文tokenをsha256hexに変換。
func (m *JWTMaker) Hash(token string) string {
	return hashOpaque(token)
}

// IDGenの実装。
// verify token/reset token/familyIDなどの生成に使う。
type RandomIDGen struct{}

// NewRandomIDGen は RandomIDGen を作る。
func NewRandomIDGen() *RandomIDGen {
	return &RandomIDGen{}
}

// New はランダムな hex 文字列を返す。
// IDGen interfaceはerrorを返さないため、失敗時は時刻ベースへフォールバック。
func (g *RandomIDGen) New() string {
	s, err := newHex(32)
	if err == nil {
		return s
	}
	return strconv.FormatInt(time.Now().UnixNano(), 16)
}

// tokenをsha256hexに変換。
func hashOpaque(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// bytesをhex文字列に。
func newHex(n int) (string, error) {
	b := make([]byte, n)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
