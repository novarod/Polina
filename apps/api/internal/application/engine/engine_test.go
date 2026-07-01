package engine_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appengine "github.com/novarod/polina/apps/api/internal/application/engine"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

// The engine use cases delegate to the repositories with the org from the caller;
// these fakes assert the org is passed through and surface the repo's error.
type fakeMissionRepo struct {
	activeHash string
	err        error
	gotOrg     uuid.UUID
}

func (f *fakeMissionRepo) FindActiveHash(_ context.Context, _, orgID uuid.UUID) (string, error) {
	f.gotOrg = orgID
	return f.activeHash, f.err
}

// Unused MissionRepository methods (only FindActiveHash is exercised here).
func (f *fakeMissionRepo) Create(context.Context, ports.Mission) (ports.Mission, error) {
	return ports.Mission{}, nil
}
func (f *fakeMissionRepo) FindByID(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (ports.Mission, error) {
	return ports.Mission{}, nil
}
func (f *fakeMissionRepo) FindByIDForUpdate(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (ports.Mission, error) {
	return ports.Mission{}, nil
}
func (f *fakeMissionRepo) List(context.Context, uuid.UUID, uuid.UUID) ([]ports.Mission, error) {
	return nil, nil
}
func (f *fakeMissionRepo) UpdateGraph(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, json.RawMessage) (ports.Mission, error) {
	return ports.Mission{}, nil
}
func (f *fakeMissionRepo) Update(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, string) (ports.Mission, error) {
	return ports.Mission{}, nil
}
func (f *fakeMissionRepo) SetActiveVersion(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, string) (ports.Mission, error) {
	return ports.Mission{}, nil
}
func (f *fakeMissionRepo) SoftDelete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}

type fakeVersionRepo struct {
	active ports.MissionVersion
	err    error
}

func (f *fakeVersionRepo) FindActive(_ context.Context, _, _ uuid.UUID) (ports.MissionVersion, error) {
	return f.active, f.err
}
func (f *fakeVersionRepo) Create(context.Context, ports.MissionVersion) (ports.MissionVersion, error) {
	return ports.MissionVersion{}, nil
}
func (f *fakeVersionRepo) FindByHash(context.Context, uuid.UUID, uuid.UUID, string) (ports.MissionVersion, error) {
	return ports.MissionVersion{}, nil
}
func (f *fakeVersionRepo) List(context.Context, uuid.UUID, uuid.UUID) ([]ports.MissionVersion, error) {
	return nil, nil
}

func codeOf(t *testing.T, err error) int {
	t.Helper()
	var appErr *apierr.AppError
	require.ErrorAs(t, err, &appErr)
	return appErr.Code
}

func TestGetActiveHash_ReturnsHash_ScopedByOrg(t *testing.T) {
	org := uuid.New()
	mr := &fakeMissionRepo{activeHash: "abc"}
	h, err := appengine.NewGetActiveHashUseCase(mr).Execute(context.Background(), org, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "abc", h)
	assert.Equal(t, org, mr.gotOrg, "org from the caller scopes the query")
}

func TestGetActiveHash_404WhenUnpublished(t *testing.T) {
	mr := &fakeMissionRepo{err: apierr.NotFound("active mission version")}
	_, err := appengine.NewGetActiveHashUseCase(mr).Execute(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, codeOf(t, err))
}

func TestGetActiveContract_ReturnsMissionData(t *testing.T) {
	vr := &fakeVersionRepo{active: ports.MissionVersion{MissionData: json.RawMessage(`{"nodes":{}}`)}}
	data, err := appengine.NewGetActiveContractUseCase(vr).Execute(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.JSONEq(t, `{"nodes":{}}`, string(data))
}

func TestGetActiveContract_404WhenNoActive(t *testing.T) {
	vr := &fakeVersionRepo{err: apierr.NotFound("active mission version")}
	_, err := appengine.NewGetActiveContractUseCase(vr).Execute(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, codeOf(t, err))
}
