package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/hash"
)

const (
	engineOrgKey   = "engineOrgID"
	engineKeyIDKey = "engineKeyID"
)

func APIKeyAuth(keys ports.OrganizationAPIKeyRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			raw := c.Request().Header.Get("x-api-key")
			if raw == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing api key")
			}
			key, err := keys.FindActiveByHash(c.Request().Context(), hash.APIKey(raw))
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid api key")
			}
			c.Set(engineOrgKey, key.OrganizationID)
			c.Set(engineKeyIDKey, key.ID)
			return next(c)
		}
	}
}

func MustGetEngineOrg(c echo.Context) uuid.UUID {
	orgID, ok := c.Get(engineOrgKey).(uuid.UUID)
	if !ok {
		panic("api key middleware not applied: org missing from context")
	}
	return orgID
}

func RateLimitByEngineKey(requestsPerMin int) echo.MiddlewareFunc {
	return RateLimitByKey(requestsPerMin, func(c echo.Context) string {
		if id, ok := c.Get(engineKeyIDKey).(uuid.UUID); ok {
			return id.String()
		}
		return c.RealIP()
	})
}

func TouchAPIKey(keys ports.OrganizationAPIKeyRepository, throttle time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if id, ok := c.Get(engineKeyIDKey).(uuid.UUID); ok {
				_ = keys.TouchLastUsed(c.Request().Context(), id, throttle)
			}
			return next(c)
		}
	}
}
