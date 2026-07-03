package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

var _ ports.OrganizationAPIKeyRepository = (*fakeKeyRepo)(nil)

type fakeKeyRepo struct {
	key        ports.OrganizationAPIKey
	findErr    error
	touchCalls int
}

func (f *fakeKeyRepo) Create(_ context.Context, k ports.OrganizationAPIKey) (ports.OrganizationAPIKey, error) {
	return k, nil
}
func (f *fakeKeyRepo) FindActiveByHash(_ context.Context, _ string) (ports.OrganizationAPIKey, error) {
	return f.key, f.findErr
}
func (f *fakeKeyRepo) ListByOrg(_ context.Context, _ uuid.UUID) ([]ports.OrganizationAPIKey, error) {
	return nil, nil
}
func (f *fakeKeyRepo) Revoke(_ context.Context, _, _ uuid.UUID) error { return nil }
func (f *fakeKeyRepo) TouchLastUsed(_ context.Context, _ uuid.UUID, _ time.Duration) error {
	f.touchCalls++
	return nil
}

func echoWithAPIKey(repo ports.OrganizationAPIKeyRepository) *echo.Echo {
	e := echo.New()
	e.GET("/engine/x", func(c echo.Context) error {
		return c.String(http.StatusOK, httpmw.MustGetEngineOrg(c).String())
	}, httpmw.APIKeyAuth(repo))
	return e
}

func TestAPIKeyAuth_Valid_SetsOrg(t *testing.T) {
	orgID := uuid.New()
	repo := &fakeKeyRepo{key: ports.OrganizationAPIKey{ID: uuid.New(), OrganizationID: orgID}}

	req := httptest.NewRequest(http.MethodGet, "/engine/x", nil)
	req.Header.Set("x-api-key", "pol_whatever")
	rec := httptest.NewRecorder()
	echoWithAPIKey(repo).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, orgID.String(), rec.Body.String())
}

func TestAPIKeyAuth_MissingHeader_401(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/engine/x", nil)
	rec := httptest.NewRecorder()
	echoWithAPIKey(&fakeKeyRepo{}).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAPIKeyAuth_UnknownOrRevoked_401(t *testing.T) {
	repo := &fakeKeyRepo{findErr: apierr.NotFound("api key")}
	req := httptest.NewRequest(http.MethodGet, "/engine/x", nil)
	req.Header.Set("x-api-key", "pol_revoked")
	rec := httptest.NewRecorder()
	echoWithAPIKey(repo).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestTouchAPIKey_CallsRepoAfterAuth(t *testing.T) {
	repo := &fakeKeyRepo{key: ports.OrganizationAPIKey{ID: uuid.New(), OrganizationID: uuid.New()}}
	e := echo.New()
	e.GET("/engine/x", func(c echo.Context) error { return c.NoContent(http.StatusOK) },
		httpmw.APIKeyAuth(repo), httpmw.TouchAPIKey(repo, time.Minute))

	req := httptest.NewRequest(http.MethodGet, "/engine/x", nil)
	req.Header.Set("x-api-key", "pol_ok")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, repo.touchCalls)
}

// seqKeyRepo returns a preset key per call, letting one request chain present
// different API keys to the same rate limiter.
type seqKeyRepo struct {
	keys []ports.OrganizationAPIKey
	n    int
}

func (f *seqKeyRepo) Create(_ context.Context, k ports.OrganizationAPIKey) (ports.OrganizationAPIKey, error) {
	return k, nil
}
func (f *seqKeyRepo) FindActiveByHash(_ context.Context, _ string) (ports.OrganizationAPIKey, error) {
	k := f.keys[f.n%len(f.keys)]
	f.n++
	return k, nil
}
func (f *seqKeyRepo) ListByOrg(_ context.Context, _ uuid.UUID) ([]ports.OrganizationAPIKey, error) {
	return nil, nil
}
func (f *seqKeyRepo) Revoke(_ context.Context, _, _ uuid.UUID) error                      { return nil }
func (f *seqKeyRepo) TouchLastUsed(_ context.Context, _ uuid.UUID, _ time.Duration) error { return nil }

// TestRateLimitByEngineKey_PerKey_NotPerIP: with one shared limiter and one IP,
// key A exhausting its budget does not block key B (keying is per key, not per IP).
func TestRateLimitByEngineKey_PerKey_NotPerIP(t *testing.T) {
	keyA := ports.OrganizationAPIKey{ID: uuid.New(), OrganizationID: uuid.New()}
	keyB := ports.OrganizationAPIKey{ID: uuid.New(), OrganizationID: uuid.New()}
	repo := &seqKeyRepo{keys: []ports.OrganizationAPIKey{keyA, keyA, keyB}} // A, A, then B

	limitMW, stop := httpmw.RateLimitByEngineKey(1)
	defer stop()
	e := echo.New()
	e.GET("/engine/x", func(c echo.Context) error { return c.NoContent(http.StatusOK) },
		httpmw.APIKeyAuth(repo), limitMW)

	codes := make([]int, 3)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/engine/x", nil)
		req.Header.Set("x-api-key", "pol_x")
		req.RemoteAddr = "203.0.113.9:1111"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		codes[i] = rec.Code
	}
	assert.Equal(t, http.StatusOK, codes[0], "key A first call")
	assert.Equal(t, http.StatusTooManyRequests, codes[1], "key A over budget")
	assert.Equal(t, http.StatusOK, codes[2], "key B has its own budget on the same IP")
}
