package middleware

import (
	"coffee-spa/entity"
	"errors"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// access tokenのclaims。
// sub / role / tv / expを扱う。
type accessClaims struct {
	Role string `json:"role"`
	TV   int    `json:"tv"`
	jwt.RegisteredClaims
}
type TokenVersionReader interface {
	GetByID(id uint) (*entity.User, error)
}

// JWTAuthはBearer JWTを検証し、contextにuser_id / role / tv を入れる。
// token_versionのDB整合はTokenVersionで行う。
func JWTAuth(secret string) echo.MiddlewareFunc {
	key := []byte(secret)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := strings.TrimSpace(c.Request().Header.Get(echo.HeaderAuthorization))
			if auth == "" {
				return writeUnauthorized(c)
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				return writeUnauthorized(c)
			}

			raw := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
			if raw == "" {
				return writeUnauthorized(c)
			}

			token, err := jwt.ParseWithClaims(raw, &accessClaims{}, func(token *jwt.Token) (interface{}, error) {
				if token.Method == nil {
					return nil, errors.New("missing signing method")
				}
				if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, errors.New("unexpected signing method")
				}
				return key, nil
			})
			if err != nil || !token.Valid {
				return writeUnauthorized(c)
			}

			claims, ok := token.Claims.(*accessClaims)
			if !ok || claims == nil {
				return writeUnauthorized(c)
			}
			if claims.Subject == "" || claims.Role == "" {
				return writeUnauthorized(c)
			}

			id64, err := strconv.ParseUint(claims.Subject, 10, 64)
			if err != nil || id64 == 0 {
				return writeUnauthorized(c)
			}

			userID := uint(id64)
			role := entity.Role(claims.Role)

			actor := &entity.Actor{
				UserID:   userID,
				Role:     role,
				TokenVer: claims.TV,
			}

			c.Set("user_id", userID)
			c.Set("role", claims.Role)
			c.Set("tv", claims.TV)
			c.Set("actor", actor)

			return next(c)
		}
	}
}

// DBのuser.token_verとJWTのtvが一致することを確認。
// user が存在しない場合も401。
func TokenVersion(userRepo TokenVersionReader) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get("user_id").(uint)
			if !ok || userID <= 0 {
				return writeUnauthorized(c)
			}

			tv, ok := c.Get("tv").(int)
			if !ok {
				return writeUnauthorized(c)
			}

			user, err := userRepo.GetByID(userID)
			if err != nil {
				return writeUnauthorized(c)
			}

			if user.TokenVer != tv {
				return writeUnauthorized(c)
			}

			return next(c)
		}
	}
}
