package controller

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

// 監査ログ。依存先はusecase.AuditUCのみ
type AuditCtl struct {
	uc usecase.AuditUC
}

func NewAuditCtl(uc usecase.AuditUC) *AuditCtl {
	return &AuditCtl{
		uc: uc,
	}
}

// 一覧レスポンス
type AuditListRes struct {
	Logs []entity.AuditLog `json:"logs"`
}

// GET /audit-logsを処理(管理者のみ)
func (ctl *AuditCtl) List(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}
	userID, err := qUint(c, "user_id")
	if err != nil {
		return writeErr(c, err)
	}
	limit, err := qInt(c, "limit", 20)
	if err != nil {
		return writeErr(c, err)
	}
	offset, err := qInt(c, "offset", 0)
	if err != nil {
		return writeErr(c, err)
	}

	logs, err := ctl.uc.List(*actor, usecase.AuditListIn{
		Type:   qStr(c, "type"),
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return writeErr(c, err)
	}
	return c.JSON(http.StatusOK, AuditListRes{
		Logs: logs,
	})
}
