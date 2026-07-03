package server

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// A non-HTTPError (e.g. a recovered panic) must not leak its message to the client.
func TestErrorHandler_NonHTTPError_HidesInternalDetail(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	newErrorHandler(slog.New(slog.DiscardHandler))(errors.New("pgx: connection to secret-host failed: password=hunter2"), c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "secret-host")
	assert.NotContains(t, rec.Body.String(), "hunter2")
	assert.Contains(t, rec.Body.String(), "internal server error")
}

// An intentional *echo.HTTPError keeps its (safe) message and status.
func TestErrorHandler_HTTPError_PreservesMessage(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	newErrorHandler(slog.New(slog.DiscardHandler))(echo.NewHTTPError(http.StatusNotFound, "mission not found"), c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "mission not found")
}
