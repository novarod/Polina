package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type limiterStore struct {
	mu       sync.Mutex
	limiters map[string]*rateLimiterEntry
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newLimiterStore() *limiterStore {
	s := &limiterStore{limiters: make(map[string]*rateLimiterEntry)}
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
	for range time.Tick(2 * time.Minute) {
		s.mu.Lock()
		for k, e := range s.limiters {
			if time.Since(e.lastSeen) > 5*time.Minute {
				delete(s.limiters, k)
			}
		}
		s.mu.Unlock()
	}
}

func RateLimit(requestsPerMin int) echo.MiddlewareFunc {
	return RateLimitByKey(requestsPerMin, func(c echo.Context) string { return c.RealIP() })
}

func RateLimitByKey(requestsPerMin int, keyFn func(echo.Context) string) echo.MiddlewareFunc {
	store := newLimiterStore()
	r := rate.Every(time.Minute / time.Duration(requestsPerMin))
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !store.get(keyFn(c), r, requestsPerMin).Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}
