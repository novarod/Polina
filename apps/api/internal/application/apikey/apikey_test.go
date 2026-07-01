package apikey_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appapikey "github.com/novarod/polina/apps/api/internal/application/apikey"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

var (
	_ ports.OrganizationAPIKeyRepository = (*fakeKeyRepo)(nil)
	_ ports.MemberRepository             = (*fakeMemberRepo)(nil)
)

type fakeKeyRepo struct {
	created     ports.OrganizationAPIKey
	createCalls int
	listed      []ports.OrganizationAPIKey
	revokeCalls int
	revokeErr   error
}

func (f *fakeKeyRepo) Create(_ context.Context, k ports.OrganizationAPIKey) (ports.OrganizationAPIKey, error) {
	f.createCalls++
	f.created = k
	return k, nil
}
func (f *fakeKeyRepo) FindActiveByHash(_ context.Context, _ string) (ports.OrganizationAPIKey, error) {
	return ports.OrganizationAPIKey{}, apierr.NotFound("api key")
}
func (f *fakeKeyRepo) ListByOrg(_ context.Context, _ uuid.UUID) ([]ports.OrganizationAPIKey, error) {
	return f.listed, nil
}
func (f *fakeKeyRepo) Revoke(_ context.Context, _, _ uuid.UUID) error {
	f.revokeCalls++
	return f.revokeErr
}
func (f *fakeKeyRepo) TouchLastUsed(_ context.Context, _ uuid.UUID, _ time.Duration) error {
	return nil
}

type fakeMemberRepo struct {
	found   ports.Member
	findErr error
}

func (f *fakeMemberRepo) Create(_ context.Context, m ports.Member) (ports.Member, error) {
	return m, nil
}
func (f *fakeMemberRepo) FindByUserAndOrg(_ context.Context, _, _ uuid.UUID) (ports.Member, error) {
	return f.found, f.findErr
}
func (f *fakeMemberRepo) SoftDeleteByOrg(_ context.Context, _ uuid.UUID) error { return nil }

func appErrCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *apierr.AppError
	require.ErrorAs(t, err, &appErr)
	return appErr.Code
}

func admin(id uuid.UUID) *fakeMemberRepo {
	return &fakeMemberRepo{found: ports.Member{ID: id, Role: member.RoleAdmin}}
}
func nonAdmin() *fakeMemberRepo {
	return &fakeMemberRepo{found: ports.Member{Role: member.RoleDesigner}}
}

// --- Create ---

func TestCreate_AdminReturnsRawOnce_StoresHash(t *testing.T) {
	memberID := uuid.New()
	kr := &fakeKeyRepo{}
	uc := appapikey.NewCreateUseCase(kr, admin(memberID))

	res, err := uc.Execute(context.Background(), appapikey.CreateInput{
		UserID: uuid.New(), OrgID: uuid.New(), Name: "  CI plugin  ",
	})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(res.Raw, "pol_"), "raw key has prefix")
	assert.Equal(t, "CI plugin", kr.created.Name, "name trimmed")
	assert.Equal(t, memberID, kr.created.CreatedByID)
	assert.NotEmpty(t, kr.created.KeyHash)
	assert.NotEqual(t, res.Raw, kr.created.KeyHash, "stored value is the hash, not the raw")
	assert.NotContains(t, kr.created.KeyHash, res.Raw)
}

func TestCreate_EmptyName_422(t *testing.T) {
	kr := &fakeKeyRepo{}
	uc := appapikey.NewCreateUseCase(kr, admin(uuid.New()))
	_, err := uc.Execute(context.Background(), appapikey.CreateInput{UserID: uuid.New(), OrgID: uuid.New(), Name: "   "})
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, appErrCode(t, err))
	assert.Equal(t, 0, kr.createCalls)
}

func TestCreate_NameTooLong_422(t *testing.T) {
	kr := &fakeKeyRepo{}
	uc := appapikey.NewCreateUseCase(kr, admin(uuid.New()))
	_, err := uc.Execute(context.Background(), appapikey.CreateInput{
		UserID: uuid.New(), OrgID: uuid.New(), Name: strings.Repeat("x", 256),
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, appErrCode(t, err))
	assert.Equal(t, 0, kr.createCalls)
}

func TestCreate_NonAdminForbidden(t *testing.T) {
	kr := &fakeKeyRepo{}
	uc := appapikey.NewCreateUseCase(kr, nonAdmin())
	_, err := uc.Execute(context.Background(), appapikey.CreateInput{UserID: uuid.New(), OrgID: uuid.New(), Name: "x"})
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
	assert.Equal(t, 0, kr.createCalls)
}

func TestCreate_UniqueRawEachTime(t *testing.T) {
	kr := &fakeKeyRepo{}
	uc := appapikey.NewCreateUseCase(kr, admin(uuid.New()))
	in := appapikey.CreateInput{UserID: uuid.New(), OrgID: uuid.New(), Name: "k"}
	a, err := uc.Execute(context.Background(), in)
	require.NoError(t, err)
	b, err := uc.Execute(context.Background(), in)
	require.NoError(t, err)
	assert.NotEqual(t, a.Raw, b.Raw, "each key is random")
}

// --- List ---

func TestList_AdminSucceeds(t *testing.T) {
	kr := &fakeKeyRepo{listed: []ports.OrganizationAPIKey{{Name: "a"}, {Name: "b"}}}
	uc := appapikey.NewListUseCase(kr, admin(uuid.New()))
	list, err := uc.Execute(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestList_NonAdminForbidden(t *testing.T) {
	uc := appapikey.NewListUseCase(&fakeKeyRepo{}, nonAdmin())
	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}

// --- Revoke ---

func TestRevoke_AdminSucceeds(t *testing.T) {
	kr := &fakeKeyRepo{}
	uc := appapikey.NewRevokeUseCase(kr, admin(uuid.New()))
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 1, kr.revokeCalls)
}

func TestRevoke_NonAdminForbidden(t *testing.T) {
	kr := &fakeKeyRepo{}
	uc := appapikey.NewRevokeUseCase(kr, nonAdmin())
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
	assert.Equal(t, 0, kr.revokeCalls)
}

func TestRevoke_NotFound(t *testing.T) {
	kr := &fakeKeyRepo{revokeErr: apierr.NotFound("api key")}
	uc := appapikey.NewRevokeUseCase(kr, admin(uuid.New()))
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
}
