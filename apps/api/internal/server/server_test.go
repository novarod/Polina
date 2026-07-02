package server

import (
	"net/http"
	"testing"
	"time"
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
