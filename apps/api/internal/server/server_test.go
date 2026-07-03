package server

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestConfigureTimeouts(t *testing.T) {
	// placeholder value that configureTimeouts must overwrite
	srv := &http.Server{ReadHeaderTimeout: time.Second}
	configureTimeouts(srv)

	if srv.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout = %v, want 5s", srv.ReadHeaderTimeout)
	}
	if srv.ReadTimeout != 10*time.Second {
		t.Fatalf("ReadTimeout = %v, want 10s", srv.ReadTimeout)
	}
	if srv.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %v, want 60s", srv.IdleTimeout)
	}
}

func echoWithObservability() *echo.Echo {
	e := echo.New()
	useObservability(e, slog.New(slog.DiscardHandler))
	e.GET("/ping", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	return e
}

func TestObservability_ResponsesCarryRequestID(t *testing.T) {
	rec := httptest.NewRecorder()
	echoWithObservability().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ping", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Header().Get(echo.HeaderXRequestID) == "" {
		t.Fatal("response is missing the X-Request-ID header")
	}
}

func TestObservability_MetricsEndpoint(t *testing.T) {
	rec := httptest.NewRecorder()
	echoWithObservability().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "go_goroutines") {
		t.Fatal("metrics body is missing the go_goroutines gauge")
	}
}
