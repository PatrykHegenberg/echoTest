package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRenderBewertungenRoute(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := renderBewertungenRoute(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestToggleWertungRoute(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/toggle/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := toggleWertungRoute(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
