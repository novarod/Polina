package middleware_test

import (
	"context"
	"errors"
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
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

const testSecret = "test-secret-at-least-32-bytes-long-0123"

var _ ports.UserRepository = (*fakeUserRepo)(nil)

type fakeUserRepo struct {
	users   map[uuid.UUID]ports.User
	findErr error
}

func newFakeUserRepo(users ...ports.User) *fakeUserRepo {
	f := &fakeUserRepo{users: make(map[uuid.UUID]ports.User, len(users))}
	for _, u := range users {
		f.users[u.ID] = u
	}
	return f
}

func (f *fakeUserRepo) Create(_ context.Context, u ports.User) (ports.User, error) { return u, nil }

func (f *fakeUserRepo) FindByEmail(_ context.Context, _ string) (ports.User, error) {
	return ports.User{}, apierr.NotFound("user")
}

func (f *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (ports.User, error) {
	if f.findErr != nil {
		return ports.User{}, f.findErr
	}
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return ports.User{}, apierr.NotFound("user")
}

func (f *fakeUserRepo) BumpTokenValidAfter(_ context.Context, _ uuid.UUID) error { return nil }

func signToken(t *testing.T, secret string, issuedAt, expiresAt time.Time, userID uuid.UUID) string {
	t.Helper()
	claims := &token.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

func echoWithAuth(users ports.UserRepository) *echo.Echo {
	e := echo.New()
	e.GET("/protected", func(c echo.Context) error {
		return c.String(http.StatusOK, httpmw.MustGetClaims(c).UserID.String())
	}, httpmw.Auth(testSecret, users))
	return e
}

func getProtected(e *echo.Echo, tok string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if tok != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestAuth_ValidBearer_PassesAndSetsClaims(t *testing.T) {
	uid := uuid.New()
	tok := signToken(t, testSecret, time.Now(), time.Now().Add(time.Hour), uid)

	rec := getProtected(echoWithAuth(newFakeUserRepo(ports.User{ID: uid})), tok)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, uid.String(), rec.Body.String())
}

func TestAuth_ValidCookie_Passes(t *testing.T) {
	uid := uuid.New()
	tok := signToken(t, testSecret, time.Now(), time.Now().Add(time.Hour), uid)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Cookie", "session="+tok)
	rec := httptest.NewRecorder()
	echoWithAuth(newFakeUserRepo(ports.User{ID: uid})).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuth_MissingToken_401(t *testing.T) {
	rec := getProtected(echoWithAuth(newFakeUserRepo()), "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_WrongSecret_401(t *testing.T) {
	uid := uuid.New()
	tok := signToken(t, "a-different-secret-of-at-least-32-bytes!", time.Now(), time.Now().Add(time.Hour), uid)
	rec := getProtected(echoWithAuth(newFakeUserRepo(ports.User{ID: uid})), tok)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_Expired_401(t *testing.T) {
	uid := uuid.New()
	tok := signToken(t, testSecret, time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour), uid)
	rec := getProtected(echoWithAuth(newFakeUserRepo(ports.User{ID: uid})), tok)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_UserNotFound_401(t *testing.T) {
	tok := signToken(t, testSecret, time.Now(), time.Now().Add(time.Hour), uuid.New())
	rec := getProtected(echoWithAuth(newFakeUserRepo()), tok)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_UserLookupInfraError_500(t *testing.T) {
	uid := uuid.New()
	tok := signToken(t, testSecret, time.Now(), time.Now().Add(time.Hour), uid)
	users := newFakeUserRepo(ports.User{ID: uid})
	users.findErr = errors.New("pgx: connection refused")

	rec := getProtected(echoWithAuth(users), tok)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAuth_TokenIssuedBeforeCutoff_401(t *testing.T) {
	uid := uuid.New()
	cutoff := time.Now()
	tok := signToken(t, testSecret, cutoff.Add(-time.Hour), cutoff.Add(time.Hour), uid)

	rec := getProtected(echoWithAuth(newFakeUserRepo(ports.User{ID: uid, TokenValidAfter: &cutoff})), tok)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// The cutoff is truncated to seconds so a token issued right after a logout-all,
// within the same second, must stay valid.
func TestAuth_TokenIssuedSameSecondAsCutoff_Passes(t *testing.T) {
	uid := uuid.New()
	issuedAt := time.Now().Truncate(time.Second)
	cutoff := issuedAt.Add(500 * time.Millisecond)
	tok := signToken(t, testSecret, issuedAt, issuedAt.Add(time.Hour), uid)

	rec := getProtected(echoWithAuth(newFakeUserRepo(ports.User{ID: uid, TokenValidAfter: &cutoff})), tok)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuth_TokenWithoutIatAndCutoffSet_401(t *testing.T) {
	uid := uuid.New()
	cutoff := time.Now()
	claims := &token.Claims{
		UserID:           uid,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	require.NoError(t, err)

	rec := getProtected(echoWithAuth(newFakeUserRepo(ports.User{ID: uid, TokenValidAfter: &cutoff})), tok)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRateLimit_ZeroLimitDoesNotPanic(t *testing.T) {
	mw, stop := httpmw.RateLimit(0) // clamped to 1 req/min instead of panicking
	defer stop()
	e := echo.New()
	e.GET("/limited", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, mw)

	req := httptest.NewRequest(http.MethodGet, "/limited", nil)
	req.RemoteAddr = "203.0.113.8:5555"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimit_AllowsUpToLimitThenBlocks(t *testing.T) {
	mw, stop := httpmw.RateLimit(2) // burst of 2
	defer stop()
	e := echo.New()
	e.GET("/limited", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, mw)

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
