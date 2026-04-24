package controller

import (
	"net/http"
	"testing"

	"coffee-spa/usecase"
)

func TestHealthCtlGetServiceUnavailable(t *testing.T) {
	ctl := NewHealthCtl(&healthUCMock{checkFn: func() error { return usecase.ErrInternal }})
	c, rec, err := newJSONContext(http.MethodGet, "/health", nil)
	if err != nil { t.Fatal(err) }
	if err := ctl.Get(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusServiceUnavailable { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}
