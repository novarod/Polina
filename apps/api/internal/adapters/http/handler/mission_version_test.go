package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	appmission "github.com/novarod/polina/apps/api/internal/application/mission"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

// fakeVersionRepo + fakeTxManager + fakeRepos scoped to handler tests.
type fakeVersionRepo struct {
	listed  []ports.MissionVersion
	findErr error
}

func (f *fakeVersionRepo) Create(_ context.Context, v ports.MissionVersion) (ports.MissionVersion, error) {
	v.VersionNumber = 1
	return v, nil
}
func (f *fakeVersionRepo) FindByHash(_ context.Context, _, _ uuid.UUID, _ string) (ports.MissionVersion, error) {
	if f.findErr != nil {
		return ports.MissionVersion{}, f.findErr
	}
	return ports.MissionVersion{}, apierr.NotFound("mission version")
}
func (f *fakeVersionRepo) List(_ context.Context, _, _ uuid.UUID) ([]ports.MissionVersion, error) {
	return f.listed, nil
}

func newVersionHandler(mr ports.MissionRepository, vr ports.MissionVersionRepository, members ports.MemberRepository) *handler.MissionVersionHandler {
	tx := &fakeTxManager{repos: &fakeRepos{members: members, missions: mr, versions: vr}}
	return handler.NewMissionVersionHandler(
		appmission.NewPublishUseCase(tx),
		appmission.NewListVersionsUseCase(mr, vr, members),
		appmission.NewGetVersionUseCase(mr, vr, members),
	)
}

func TestMissionVersionHandler_Publish_200_Designer(t *testing.T) {
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New(), Graph: []byte(missionValidGraph)}}
	h := newVersionHandler(mr, &fakeVersionRepo{}, memberWith(member.RoleDesigner))
	c, rec := newCtx(t, http.MethodPost, "/x", "", uuid.New())
	setMissionParams(c, uuid.New().String(), uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusOK, invoke(t, h.Publish, c, rec))
}

func TestMissionVersionHandler_Publish_403_Viewer(t *testing.T) {
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New(), Graph: []byte(missionValidGraph)}}
	h := newVersionHandler(mr, &fakeVersionRepo{}, memberWith(member.RoleViewer))
	c, rec := newCtx(t, http.MethodPost, "/x", "", uuid.New())
	setMissionParams(c, uuid.New().String(), uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusForbidden, invoke(t, h.Publish, c, rec))
}

func TestMissionVersionHandler_ListVersions_200(t *testing.T) {
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New()}}
	vr := &fakeVersionRepo{listed: []ports.MissionVersion{{VersionNumber: 1}}}
	h := newVersionHandler(mr, vr, memberWith(member.RoleViewer))
	c, rec := newCtx(t, http.MethodGet, "/x", "", uuid.New())
	setMissionParams(c, uuid.New().String(), uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusOK, invoke(t, h.ListVersions, c, rec))
}

func TestMissionVersionHandler_GetVersion_404_UnknownHash(t *testing.T) {
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New()}}
	vr := &fakeVersionRepo{findErr: apierr.NotFound("mission version")}
	h := newVersionHandler(mr, vr, memberWith(member.RoleViewer))
	c, rec := newCtx(t, http.MethodGet, "/x", "", uuid.New())
	c.SetParamNames("id", "workspaceID", "missionID", "hash")
	c.SetParamValues(uuid.New().String(), uuid.New().String(), uuid.New().String(), "nope")
	assert.Equal(t, http.StatusNotFound, invoke(t, h.GetVersion, c, rec))
}
