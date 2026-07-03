package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type limiterStore struct {
	mu        sync.Mutex
	limiters  map[string]*rateLimiterEntry
	stop      chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newLimiterStore() *limiterStore {
	s := &limiterStore{
		limiters: make(map[string]*rateLimiterEntry),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go s.cleanup()
	return s
}

func (s *limiterStore) get(key string, r rate.Limit, b int) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.limiters[key]
	if !ok {
		entry = &rateLimiterEntry{limiter: rate.NewLimiter(r, b)}
		s.limiters[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter
}

func (s *limiterStore) cleanup() {
	defer close(s.done)
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			for k, e := range s.limiters {
				if time.Since(e.lastSeen) > 5*time.Minute {
					delete(s.limiters, k)
				}
			}
			s.mu.Unlock()
		case <-s.stop:
			return
		}
	}
}

func (s *limiterStore) Close() {
	s.closeOnce.Do(func() { close(s.stop) })
	<-s.done
}

func RateLimit(requestsPerMin int) (echo.MiddlewareFunc, func()) {
	return RateLimitByKey(requestsPerMin, func(c echo.Context) string { return c.RealIP() })
}

func RateLimitByKey(requestsPerMin int, keyFn func(echo.Context) string) (echo.MiddlewareFunc, func()) {
	if requestsPerMin < 1 {
		requestsPerMin = 1
	}
	store := newLimiterStore()
	r := rate.Every(time.Minute / time.Duration(requestsPerMin))
	mw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !store.get(keyFn(c), r, requestsPerMin).Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
	return mw, store.Close
}
