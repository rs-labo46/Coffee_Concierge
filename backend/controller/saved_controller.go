package controller

import (
	"coffee-spa/entity"
	"coffee-spa/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPの入り口。依存先はusecase.SavedUCのみ
type SavedCtl struct {
	uc usecase.SavedUC
}

func NewSavedCtl(uc usecase.SavedUC) *SavedCtl {
	return &SavedCtl{
		uc: uc,
	}
}

type SaveSavedReq struct {
	SessionID    uint `json:"session_id"`
	SuggestionID uint `json:"suggestion_id"`
}

// 保存したものの提案を返すレスポンス
type SavedSuggestionRes struct {
	Saved entity.SavedSuggestion `json:"saved"`
}

// 保存したものの一覧を返すレスポンス
type SavedSuggestionListRes struct {
	Saved []entity.SavedSuggestion `json:"saved"`
}

// POST /saved-suggestionsを処理(認証ユーザーのみ)
func (ctl *SavedCtl) Save(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}

	var req SaveSavedReq
	if err := c.Bind(&req); err != nil {
		return writeErr(c, ErrInvalidRequest)
	}

	saved, err := ctl.uc.Save(usecase.SaveSuggestionIn{
		Actor:        *actor,
		SessionID:    req.SessionID,
		SuggestionID: req.SuggestionID,
	})
	if err != nil {
		return writeErr(c, err)
	}
	return c.JSON(http.StatusCreated, SavedSuggestionRes{
		Saved: saved,
	})
}

// GET /saved-suggestionsを処理(認証ユーザーのみ)
func (ctl *SavedCtl) List(c echo.Context) error {
	actor, err := requireActor(c)
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
	list, err := ctl.uc.List(usecase.ListSavedIn{
		Actor:  *actor,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return writeErr(c, err)
	}
	return c.JSON(http.StatusOK, SavedSuggestionListRes{
		Saved: list,
	})
}

// DELETE /saved-suggestions/:suggestionIdを処理(認証ユーザーのみ)
func (ctl *SavedCtl) Delete(c echo.Context) error {
	actor, err := requireActor(c)
	if err != nil {
		return writeErr(c, err)
	}
	SuggestionID, err := pUint(c, "suggestionId")
	if err != nil {
		return writeErr(c, err)
	}
	err = ctl.uc.Delete(usecase.DeleteSavedIn{
		Actor:        *actor,
		SuggestionID: SuggestionID,
	})
	if err != nil {
		return writeErr(c, err)
	}
	return c.JSON(http.StatusOK, MsgRes{
		Message: "saved suggestion deleted",
	})
}
