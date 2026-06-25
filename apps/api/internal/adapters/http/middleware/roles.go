package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/novarod/polina/apps/api/internal/domain/member"
)

func RequireRole(minimum member.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := MustGetClaims(c)
			if !claims.Role.AtLeast(minimum) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient role")
			}
			return next(c)
		}
	}
}
