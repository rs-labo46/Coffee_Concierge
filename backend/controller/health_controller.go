package controller

import (
	"net/http"

	"coffee-spa/usecase"

	"github.com/labstack/echo/v4"
)

type HealthCtl struct {
	uc usecase.HealthUsecase
}

// HealthCtlを作る
func NewHealthCtl(uc usecase.HealthUsecase) HealthCtl {
	return HealthCtl{uc}
}

// Getはhealth checkを返す
func (c HealthCtl) Get(ctx echo.Context) error {
	err := c.uc.Check()
	if err != nil {
		return ctx.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "ng",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
