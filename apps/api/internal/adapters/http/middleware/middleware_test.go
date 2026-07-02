package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	"github.com/novarod/polina/apps/api/internal/application/token"
)

const testSecret = "test-secret-at-least-32-bytes-long-0123"

func signToken(t *testing.T, secret string, expiresAt time.Time, userID uuid.UUID) string {
	t.Helper()
	claims := &token.Claims{
		UserID:           userID,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expiresAt)},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

func echoWithAuth() *echo.Echo {
	e := echo.New()
	e.GET("/protected", func(c echo.Context) error {
		return c.String(http.StatusOK, httpmw.MustGetClaims(c).UserID.String())
	}, httpmw.Auth(testSecret))
	return e
}

func TestAuth_ValidBearer_PassesAndSetsClaims(t *testing.T) {
	uid := uuid.New()
	tok := signToken(t, testSecret, time.Now().Add(time.Hour), uid)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tok)
	rec := httptest.NewRecorder()
	echoWithAuth().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, uid.String(), rec.Body.String())
}

func TestAuth_ValidCookie_Passes(t *testing.T) {
	uid := uuid.New()
	tok := signToken(t, testSecret, time.Now().Add(time.Hour), uid)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Cookie", "session="+tok)
	rec := httptest.NewRecorder()
	echoWithAuth().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuth_MissingToken_401(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	echoWithAuth().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_WrongSecret_401(t *testing.T) {
	tok := signToken(t, "a-different-secret-of-at-least-32-bytes!", time.Now().Add(time.Hour), uuid.New())
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tok)
	rec := httptest.NewRecorder()
	echoWithAuth().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_Expired_401(t *testing.T) {
	tok := signToken(t, testSecret, time.Now().Add(-time.Hour), uuid.New())
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+tok)
	rec := httptest.NewRecorder()
	echoWithAuth().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRateLimit_ZeroLimitDoesNotPanic(t *testing.T) {
	e := echo.New()
	e.GET("/limited", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, httpmw.RateLimit(0)) // clamped to 1 req/min instead of panicking

	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.RemoteAddr = "203.0.113.8:5555"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimit_AllowsUpToLimitThenBlocks(t *testing.T) {
	e := echo.New()
	e.GET("/limited", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, httpmw.RateLimit(2)) // burst of 2

	codes := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		req.RemoteAddr = "203.0.113.7:5555"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		codes = append(codes, rec.Code)
	}

	assert.Equal(t, http.StatusOK, codes[0])
	assert.Equal(t, http.StatusOK, codes[1])
	assert.Equal(t, http.StatusTooManyRequests, codes[2])
}
