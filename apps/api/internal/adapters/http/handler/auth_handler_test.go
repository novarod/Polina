package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	appauth "github.com/novarod/polina/apps/api/internal/application/auth"
	"github.com/novarod/polina/apps/api/internal/application/token"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

const meTestSecret = "me-handler-test-secret"

type fakeUserRepo struct {
	user ports.User
}

func (f *fakeUserRepo) Create(_ context.Context, u ports.User) (ports.User, error) { return u, nil }
func (f *fakeUserRepo) FindByEmail(_ context.Context, _ string) (ports.User, error) {
	return f.user, nil
}
func (f *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (ports.User, error) {
	if id != f.user.ID {
		return ports.User{}, apierr.NotFound("user")
	}
	return f.user, nil
}
func (f *fakeUserRepo) BumpTokenValidAfter(_ context.Context, _ uuid.UUID) error { return nil }

func meServer(users ports.UserRepository) *echo.Echo {
	e := echo.New()
	h := handler.NewAuthHandler(nil, nil, nil, appauth.NewMeUseCase(users), handler.CookieConfig{})
	e.GET("/auth/me", h.Me, httpmw.Auth(meTestSecret, users))
	return e
}

func signedSession(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	claims := &token.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(meTestSecret))
	require.NoError(t, err)
	return signed
}

func TestAuthHandler_Me_200_WithValidCookie(t *testing.T) {
	user := ports.User{ID: uuid.New(), Name: "Alice"}
	e := meServer(&fakeUserRepo{user: user})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(&http.Cookie{
		Name:     "session",
		Value:    signedSession(t, user.ID),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"name":"Alice"`)
	assert.Contains(t, rec.Body.String(), user.ID.String())
}

func TestAuthHandler_Me_401_WithoutCookie(t *testing.T) {
	e := meServer(&fakeUserRepo{user: ports.User{ID: uuid.New()}})

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/me", nil))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
