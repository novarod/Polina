package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/novarod/polina/apps/api/internal/application/token"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

const claimsKey = "claims"

func Auth(jwtSecret string, users ports.UserRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("session")
			if err != nil {
				header := c.Request().Header.Get("Authorization")
				if !strings.HasPrefix(header, "Bearer ") {
					return echo.NewHTTPError(http.StatusUnauthorized, "missing session")
				}
				cookie = &http.Cookie{Value: strings.TrimPrefix(header, "Bearer ")}
			}

			claims := &token.Claims{}
			tok, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid signing method")
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !tok.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired session")
			}

			user, err := users.FindByID(c.Request().Context(), claims.UserID)
			if err != nil {
				if apierr.IsNotFound(err) {
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired session")
				}
				return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
			}
			if revoked(claims, user.TokenValidAfter) {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired session")
			}

			c.Set(claimsKey, claims)
			return next(c)
		}
	}
}

func revoked(claims *token.Claims, tokenValidAfter *time.Time) bool {
	if tokenValidAfter == nil {
		return false
	}
	if claims.IssuedAt == nil {
		return true
	}
	return claims.IssuedAt.Before(tokenValidAfter.Truncate(time.Second))
}

func MustGetClaims(c echo.Context) *token.Claims {
	claims, ok := c.Get(claimsKey).(*token.Claims)
	if !ok || claims == nil {
		panic("auth middleware not applied: claims missing from context")
	}
	return claims
}
