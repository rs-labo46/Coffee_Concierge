package controller

import (
	"errors"
	"net/http"

	"coffee-spa/repository"

	"github.com/labstack/echo/v4"
)

// 入力不正
var ErrInvalidRequest = errors.New("invalid request")

// APIの失敗レスポンス
type ErrRes struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// usecase・repository・controllerのエラーをHTTPstatusとAPI・error codeに変換
func mapError(err error) (int, string, string) {
	if err == nil {
		return http.StatusInternalServerError, "internal", "internal server error"
	}
	switch {
	// controller層のbind・param・query失敗
	case errors.Is(err, ErrInvalidRequest):
		return http.StatusBadRequest, "invalid_request", "request body or query is invalid"

	// 認証失敗(未ログイン、トークン不正、refresh拒否など)
	case errors.Is(err, repository.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized", "authentication failed"

	// 権限不足
	case errors.Is(err, repository.ErrForbidden):
		return http.StatusForbidden, "forbidden", "permission denied"

	// 対象なし(存在しないidや、guest sessionの期限切れなど)
	case errors.Is(err, repository.ErrNotFound):
		return http.StatusNotFound, "not_found", "resource not found"

	// 重複や二重操作(既存メール重複、二重保存、すでに使用済み)
	case errors.Is(err, repository.ErrConflict):
		return http.StatusConflict, "conflict", "resource conflict"

	// used済みtokenを再利用する状態の不整合
	case errors.Is(err, repository.ErrInvalidState):
		return http.StatusConflict, "conflict", "resource state is invalid"

	// レート制限
	case errors.Is(err, repository.ErrRateLimited):
		return http.StatusTooManyRequests, "rate_limited", "too many requests"

	default:
		return http.StatusInternalServerError, "internal", "internal server error"
	}
}

// controllerから共通で使う失敗レスポンス関数
func writeErr(c echo.Context, err error) error {
	status, code, msg := mapError(err)

	return c.JSON(status, ErrRes{
		Error:   code,
		Message: msg,
	})
}

// bind失敗やparam/query不正のときにErrInvalidRequestを毎回呼ばないようにする。
func writeInvalidRequest(c echo.Context) error {
	return writeErr(c, ErrInvalidRequest)
}
