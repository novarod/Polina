package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	appapikey "github.com/novarod/polina/apps/api/internal/application/apikey"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

// fakeAPIKeyRepo is shared by the api-key and engine handler tests.
type fakeAPIKeyRepo struct {
	listed    []ports.OrganizationAPIKey
	active    ports.OrganizationAPIKey
	findErr   error
	revokeErr error
}

func (f *fakeAPIKeyRepo) Create(_ context.Context, k ports.OrganizationAPIKey) (ports.OrganizationAPIKey, error) {
	return k, nil
}
func (f *fakeAPIKeyRepo) FindActiveByHash(_ context.Context, _ string) (ports.OrganizationAPIKey, error) {
	return f.active, f.findErr
}
func (f *fakeAPIKeyRepo) ListByOrg(_ context.Context, _ uuid.UUID) ([]ports.OrganizationAPIKey, error) {
	return f.listed, nil
}
func (f *fakeAPIKeyRepo) Revoke(_ context.Context, _, _ uuid.UUID) error { return f.revokeErr }
func (f *fakeAPIKeyRepo) TouchLastUsed(_ context.Context, _ uuid.UUID, _ time.Duration) error {
	return nil
}

func newAPIKeyHandler(kr ports.OrganizationAPIKeyRepository, members ports.MemberRepository) *handler.APIKeyHandler {
	return handler.NewAPIKeyHandler(
		appapikey.NewCreateUseCase(kr, members),
		appapikey.NewListUseCase(kr, members),
		appapikey.NewRevokeUseCase(kr, members),
	)
}

func TestAPIKeyHandler_Create_201_Admin_ReturnsRaw(t *testing.T) {
	h := newAPIKeyHandler(&fakeAPIKeyRepo{}, memberWith(member.RoleAdmin))
	c, rec := newCtx(t, http.MethodPost, "/x", `{"name":"CI"}`, uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	assert.Equal(t, http.StatusCreated, invoke(t, h.Create, c, rec))
	assert.Contains(t, rec.Body.String(), `"key":"pol_`)
}

func TestHandler_MalformedBody_400_Generic(t *testing.T) {
	h := newAPIKeyHandler(&fakeAPIKeyRepo{}, memberWith(member.RoleAdmin))
	c, _ := newCtx(t, http.MethodPost, "/x", `{"name": `, uuid.New()) // truncated JSON
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	err := h.Create(c)
	var he *echo.HTTPError
	require.ErrorAs(t, err, &he)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "invalid request body", he.Message, "bind errors must not leak parser internals")
}

func TestAPIKeyHandler_Create_403_NonAdmin(t *testing.T) {
	h := newAPIKeyHandler(&fakeAPIKeyRepo{}, memberWith(member.RoleDesigner))
	c, rec := newCtx(t, http.MethodPost, "/x", `{"name":"CI"}`, uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	assert.Equal(t, http.StatusForbidden, invoke(t, h.Create, c, rec))
}

func TestAPIKeyHandler_List_200_NoSecret(t *testing.T) {
	kr := &fakeAPIKeyRepo{listed: []ports.OrganizationAPIKey{{ID: uuid.New(), Name: "k", KeyHash: "secret-hash"}}}
	h := newAPIKeyHandler(kr, memberWith(member.RoleAdmin))
	c, rec := newCtx(t, http.MethodGet, "/x", "", uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	assert.Equal(t, http.StatusOK, invoke(t, h.List, c, rec))
	assert.NotContains(t, rec.Body.String(), "secret-hash", "list never exposes the key hash")
}

func TestAPIKeyHandler_Revoke_204_Admin(t *testing.T) {
	h := newAPIKeyHandler(&fakeAPIKeyRepo{}, memberWith(member.RoleAdmin))
	c, rec := newCtx(t, http.MethodDelete, "/x", "", uuid.New())
	c.SetParamNames("id", "keyID")
	c.SetParamValues(uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusNoContent, invoke(t, h.Revoke, c, rec))
}

func TestAPIKeyHandler_Revoke_404_Unknown(t *testing.T) {
	h := newAPIKeyHandler(&fakeAPIKeyRepo{revokeErr: apierr.NotFound("api key")}, memberWith(member.RoleAdmin))
	c, rec := newCtx(t, http.MethodDelete, "/x", "", uuid.New())
	c.SetParamNames("id", "keyID")
	c.SetParamValues(uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusNotFound, invoke(t, h.Revoke, c, rec))
}
