package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	appws "github.com/novarod/polina/apps/api/internal/application/workspace"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

// fakeWsRepo implements ports.WorkspaceRepository for handler tests.
// (fakeMemberRepo, newCtx and invoke are shared from organization_handler_test.go.)
type fakeWsRepo struct {
	findByID    ports.Workspace
	findByIDErr error
}

func (f *fakeWsRepo) Create(_ context.Context, w ports.Workspace) (ports.Workspace, error) {
	return w, nil
}
func (f *fakeWsRepo) FindByID(_ context.Context, _, _ uuid.UUID) (ports.Workspace, error) {
	return f.findByID, f.findByIDErr
}
func (f *fakeWsRepo) List(_ context.Context, _ uuid.UUID) ([]ports.Workspace, error) {
	return nil, nil
}
func (f *fakeWsRepo) Update(_ context.Context, _, _ uuid.UUID, name, description string) (ports.Workspace, error) {
	return ports.Workspace{Name: name, Description: description}, nil
}
func (f *fakeWsRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return nil }

func newWsHandler(ws ports.WorkspaceRepository, members ports.MemberRepository) *handler.WorkspaceHandler {
	return handler.NewWorkspaceHandler(
		appws.NewCreateUseCase(ws, members),
		appws.NewListUseCase(ws, members),
		appws.NewGetUseCase(ws, members),
		appws.NewUpdateUseCase(ws, members),
		appws.NewDeleteUseCase(ws, members),
	)
}

func memberWith(role member.Role) *fakeMemberRepo {
	return &fakeMemberRepo{found: ports.Member{Role: role}}
}

func TestWsHandler_Create_201_Designer(t *testing.T) {
	h := newWsHandler(&fakeWsRepo{}, memberWith(member.RoleDesigner))
	c, rec := newCtx(t, http.MethodPost, "/x", `{"name":"Team A","description":"d"}`, uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	assert.Equal(t, http.StatusCreated, invoke(t, h.Create, c, rec))
}

func TestWsHandler_Create_403_Viewer(t *testing.T) {
	h := newWsHandler(&fakeWsRepo{}, memberWith(member.RoleViewer))
	c, rec := newCtx(t, http.MethodPost, "/x", `{"name":"Team A"}`, uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	assert.Equal(t, http.StatusForbidden, invoke(t, h.Create, c, rec))
}

func TestWsHandler_Create_400_BadOrgID(t *testing.T) {
	h := newWsHandler(&fakeWsRepo{}, memberWith(member.RoleDesigner))
	c, rec := newCtx(t, http.MethodPost, "/x", `{"name":"Team A"}`, uuid.New())
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, invoke(t, h.Create, c, rec))
}

func TestWsHandler_Get_404(t *testing.T) {
	h := newWsHandler(&fakeWsRepo{findByIDErr: apierr.NotFound("workspace")}, memberWith(member.RoleViewer))
	c, rec := newCtx(t, http.MethodGet, "/x", "", uuid.New())
	c.SetParamNames("id", "workspaceID")
	c.SetParamValues(uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusNotFound, invoke(t, h.Get, c, rec))
}

func TestWsHandler_Delete_204_Designer(t *testing.T) {
	wid := uuid.New()
	h := newWsHandler(&fakeWsRepo{findByID: ports.Workspace{ID: wid}}, memberWith(member.RoleDesigner))
	c, rec := newCtx(t, http.MethodDelete, "/x", "", uuid.New())
	c.SetParamNames("id", "workspaceID")
	c.SetParamValues(uuid.New().String(), wid.String())
	assert.Equal(t, http.StatusNoContent, invoke(t, h.Delete, c, rec))
}

func TestWsHandler_Update_403_Viewer(t *testing.T) {
	h := newWsHandler(&fakeWsRepo{}, memberWith(member.RoleViewer))
	c, rec := newCtx(t, http.MethodPatch, "/x", `{"name":"New"}`, uuid.New())
	c.SetParamNames("id", "workspaceID")
	c.SetParamValues(uuid.New().String(), uuid.New().String())
	assert.Equal(t, http.StatusForbidden, invoke(t, h.Update, c, rec))
}
