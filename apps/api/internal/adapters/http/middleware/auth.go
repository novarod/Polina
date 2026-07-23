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
			raw, ok := sessionToken(c)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing session")
			}
			claims, err := verifySession(c, jwtSecret, users, raw)
			if err != nil {
				return err
			}
			c.Set(claimsKey, claims)
			return next(c)
		}
	}
}

func AuthOptional(jwtSecret string, users ports.UserRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			raw, ok := sessionToken(c)
			if !ok {
				return next(c)
			}
			claims, err := verifySession(c, jwtSecret, users, raw)
			if err != nil {
				return err
			}
			c.Set(claimsKey, claims)
			return next(c)
		}
	}
}

func sessionToken(c echo.Context) (string, bool) {
	if cookie, err := c.Cookie("session"); err == nil {
		return cookie.Value, true
	}
	header := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer "), true
	}
	return "", false
}

func verifySession(c echo.Context, jwtSecret string, users ports.UserRepository, raw string) (*token.Claims, error) {
	claims := &token.Claims{}
	tok, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !tok.Valid {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired session")
	}

	user, err := users.FindByID(c.Request().Context(), claims.UserID)
	if err != nil {
		if apierr.IsNotFound(err) {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired session")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
	if revoked(claims, user.TokenValidAfter) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired session")
	}
	return claims, nil
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

func GetClaims(c echo.Context) (*token.Claims, bool) {
	claims, ok := c.Get(claimsKey).(*token.Claims)
	return claims, ok && claims != nil
}

func MustGetClaims(c echo.Context) *token.Claims {
	claims, ok := GetClaims(c)
	if !ok {
		panic("auth middleware not applied: claims missing from context")
	}
	return claims
}
