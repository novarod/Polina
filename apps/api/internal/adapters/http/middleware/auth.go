package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/novarod/polina/apps/api/internal/domain/member"
)

type Claims struct {
	UserID   uuid.UUID   `json:"user_id"`
	MemberID uuid.UUID   `json:"member_id"`
	OrgID    uuid.UUID   `json:"org_id"`
	Role     member.Role `json:"role"`
	jwt.RegisteredClaims
}

const claimsKey = "claims"

func Auth(jwtSecret string) echo.MiddlewareFunc {
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

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid signing method")
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired session")
			}

			c.Set(claimsKey, claims)
			return next(c)
		}
	}
}

func MustGetClaims(c echo.Context) *Claims {
	claims, ok := c.Get(claimsKey).(*Claims)
	if !ok || claims == nil {
		panic("auth middleware not applied: claims missing from context")
	}
	return claims
}
