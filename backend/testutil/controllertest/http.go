package controllertest

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
)

func JSONContext(method string, target string, body string) (*echo.Echo, echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e, e.NewContext(req, rec), rec
}

func EmptyContext(method string, target string) (*echo.Echo, echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	return e, e.NewContext(req, rec), rec
}

func WithParam(c echo.Context, name string, value string) echo.Context {
	c.SetParamNames(name)
	c.SetParamValues(value)
	return c
}

func NewCookie(name string, value string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value}
}
