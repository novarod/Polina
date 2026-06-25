package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/novarod/polina/apps/api/pkg/hash"
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
	for range time.Tick(5 * time.Minute) {
		s.mu.Lock()
		for k, e := range s.limiters {
			if time.Since(e.lastSeen) > 10*time.Minute {
				delete(s.limiters, k)
			}
		}
		s.mu.Unlock()
	}
}

func RateLimit(requestsPerMin int) echo.MiddlewareFunc {
	store := newLimiterStore()
	r := rate.Every(time.Minute / time.Duration(requestsPerMin))
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := c.RealIP()
			if !store.get(key, r, requestsPerMin).Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}

func EngineRateLimit(requestsPerMin int) echo.MiddlewareFunc {
	store := newLimiterStore()
	r := rate.Every(time.Minute / time.Duration(requestsPerMin))
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.Request().Header.Get("x-api-key")
			if apiKey == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing x-api-key")
			}
			key := hash.APIKey(apiKey)
			if !store.get(key, r, requestsPerMin).Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "engine rate limit exceeded")
			}
			return next(c)
		}
	}
}
