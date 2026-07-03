package middleware

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestLimiterStore_CloseStopsCleanupAndIsIdempotent(t *testing.T) {
	s := newLimiterStore()

	if s.get("key", rate.Every(time.Second), 1) == nil {
		t.Fatal("expected a limiter for a fresh key")
	}

	s.Close()

	select {
	case <-s.done:
	default:
		t.Fatal("cleanup goroutine still running after Close")
	}

	s.Close()
}
