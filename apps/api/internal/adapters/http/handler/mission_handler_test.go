package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	appmission "github.com/novarod/polina/apps/api/internal/application/mission"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

const missionValidGraph = `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"END"}],"edges":[{"id":"e1","source":"n1","target":"n2"}]}`
const missionCyclicGraph = `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"END"}],"edges":[{"id":"e1","source":"n1","target":"n2"},{"id":"e2","source":"n2","target":"n1"}]}`

// fakeMissionRepo for handler tests. (fakeWsRepo, fakeMemberRepo, memberWith,
// newCtx, invoke are shared from workspace/organization handler tests.)
type fakeMissionRepo struct {
	findByID    ports.Mission
	findByIDErr error
}

func (f *fakeMissionRepo) Create(_ context.Context, m ports.Mission) (ports.Mission, error) {
	return m, nil
}
func (f *fakeMissionRepo) FindByID(_ context.Context, _, _, _ uuid.UUID) (ports.Mission, error) {
	return f.findByID, f.findByIDErr
}
func (f *fakeMissionRepo) FindByIDForUpdate(_ context.Context, _, _, _ uuid.UUID) (ports.Mission, error) {
	return f.findByID, f.findByIDErr
}
func (f *fakeMissionRepo) FindActiveHash(_ context.Context, _, _ uuid.UUID) (string, error) {
	if f.findByIDErr != nil {
		return "", f.findByIDErr
	}
	if f.findByID.ActiveHash != nil {
		return *f.findByID.ActiveHash, nil
	}
	return "", apierr.NotFound("active mission version")
}
func (f *fakeMissionRepo) List(_ context.Context, _, _ uuid.UUID) ([]ports.Mission, error) {
	return nil, nil
}
func (f *fakeMissionRepo) UpdateGraph(_ context.Context, _, _, _ uuid.UUID, graph json.RawMessage) (ports.Mission, error) {
	return ports.Mission{Graph: graph}, nil
}
func (f *fakeMissionRepo) Update(_ context.Context, _, _, _ uuid.UUID, name, description string) (ports.Mission, error) {
	return ports.Mission{Name: name, Description: description}, nil
}
func (f *fakeMissionRepo) SetActiveVersion(_ context.Context, _, _, _ uuid.UUID, hash, status string) (ports.Mission, error) {
	m := f.findByID
	m.ActiveHash = &hash
	m.Status = status
	return m, nil
}
func (f *fakeMissionRepo) SoftDelete(_ context.Context, _, _, _ uuid.UUID) error { return nil }

func newMissionHandler(mr ports.MissionRepository, members ports.MemberRepository) *handler.MissionHandler {
	wr := &fakeWsRepo{} // parent workspace "exists" (no error)
	return handler.NewMissionHandler(
		appmission.NewCreateUseCase(mr, wr, members),
		appmission.NewListUseCase(mr, members),
		appmission.NewGetUseCase(mr, members),
		appmission.NewUpdateUseCase(mr, members),
		appmission.NewUpdateGraphUseCase(mr, members),
		appmission.NewDeleteUseCase(mr, members),
	)
}

func setMissionParams(c echo.Context, org, ws, mission string) {
	c.SetParamNames("id", "workspaceID", "missionID")
	c.SetParamValues(org, ws, mission)
}

func TestMissionHandler_Create_201_Designer(t *testing.T) {
	h := newMissionHandler(&fakeMissionRepo{}, memberWith(member.RoleDesigner))
	c, rec := newCtx(t, http.MethodPost, "/x", `{"name":"Old Country"}`, uuid.New())
	c.SetParamNames("id", "workspaceID")
	c.SetParamValues(uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusCreated, invoke(t, h.Create, c, rec))
}

func TestMissionHandler_Create_403_Viewer(t *testing.T) {
	h := newMissionHandler(&fakeMissionRepo{}, memberWith(member.RoleViewer))
	c, rec := newCtx(t, http.MethodPost, "/x", `{"name":"Old Country"}`, uuid.New())
	c.SetParamNames("id", "workspaceID")
	c.SetParamValues(uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusForbidden, invoke(t, h.Create, c, rec))
}

func TestMissionHandler_UpdateGraph_200_Valid(t *testing.T) {
	h := newMissionHandler(&fakeMissionRepo{}, memberWith(member.RoleDesigner))
	c, rec := newCtx(t, http.MethodPut, "/x", missionValidGraph, uuid.New())
	setMissionParams(c, uuid.New().String(), uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusOK, invoke(t, h.UpdateGraph, c, rec))
}

func TestMissionHandler_UpdateGraph_422_Cyclic(t *testing.T) {
	h := newMissionHandler(&fakeMissionRepo{}, memberWith(member.RoleDesigner))
	c, rec := newCtx(t, http.MethodPut, "/x", missionCyclicGraph, uuid.New())
	setMissionParams(c, uuid.New().String(), uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusUnprocessableEntity, invoke(t, h.UpdateGraph, c, rec))
}

func TestMissionHandler_Get_404(t *testing.T) {
	h := newMissionHandler(&fakeMissionRepo{findByIDErr: apierr.NotFound("mission")}, memberWith(member.RoleViewer))
	c, rec := newCtx(t, http.MethodGet, "/x", "", uuid.New())
	setMissionParams(c, uuid.New().String(), uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusNotFound, invoke(t, h.Get, c, rec))
}

func TestMissionHandler_Get_400_BadMissionID(t *testing.T) {
	h := newMissionHandler(&fakeMissionRepo{}, memberWith(member.RoleViewer))
	c, rec := newCtx(t, http.MethodGet, "/x", "", uuid.New())
	setMissionParams(c, uuid.New().String(), uuid.New().String(), "not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, invoke(t, h.Get, c, rec))
}

func TestMissionHandler_Delete_204_Designer(t *testing.T) {
	mid := uuid.New()
	h := newMissionHandler(&fakeMissionRepo{findByID: ports.Mission{ID: mid}}, memberWith(member.RoleDesigner))
	c, rec := newCtx(t, http.MethodDelete, "/x", "", uuid.New())
	setMissionParams(c, uuid.New().String(), uuid.New().String(), mid.String())
	assert.Equal(t, http.StatusNoContent, invoke(t, h.Delete, c, rec))
}
