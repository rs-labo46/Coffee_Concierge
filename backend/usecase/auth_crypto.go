package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// bcryptを使うパスワードハッシュ実装
type BcryptHasher struct{}

// BcryptHasherを作る
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

// 平文パスワードをbcrypt hashにする
func (h *BcryptHasher) Hash(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// hashと平文パスワードを照合する
func (h *BcryptHasher) Compare(hash string, pw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw))
}

// JWTMakerはJWT / csrf / opaque tokenを作る
type JWTMaker struct {
	secret []byte
}

// access token の独自claims
type accessClaims struct {
	Role string `json:"role"`
	TV   int    `json:"tv"`
	jwt.RegisteredClaims
}

// JWTMakerを作る
func NewJWTMaker(secret string) *JWTMaker {
	return &JWTMaker{secret: []byte(secret)}
}

// access tokenを作る
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

// csrf tokenを作る
func (m *JWTMaker) NewCSRF() (string, error) {
	return newHex(32)
}

// refresh / verify / reset 用のランダムtokenを作る
func (m *JWTMaker) NewOpaque() (string, error) {
	return newHex(32)
}

// refresh token family idを作る
func (m *JWTMaker) NewFamilyID() (string, error) {
	return newHex(16)
}

// bytesをhex文字列にする
func newHex(n int) (string, error) {
	b := make([]byte, n)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
