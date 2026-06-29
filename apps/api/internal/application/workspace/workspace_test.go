package workspace_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"

	appws "github.com/novarod/polina/apps/api/internal/application/workspace"
)

// --- fakes ---

var (
	_ ports.WorkspaceRepository = (*fakeWorkspaceRepo)(nil)
	_ ports.MemberRepository    = (*fakeMemberRepo)(nil)
)

type fakeWorkspaceRepo struct {
	createCalls int
	created     ports.Workspace
	findByID    ports.Workspace
	findByIDErr error
	listResult  []ports.Workspace
	updated     ports.Workspace
	softDeleted []uuid.UUID
}

func (f *fakeWorkspaceRepo) Create(_ context.Context, w ports.Workspace) (ports.Workspace, error) {
	f.createCalls++
	f.created = w
	return w, nil
}
func (f *fakeWorkspaceRepo) FindByID(_ context.Context, _, _ uuid.UUID) (ports.Workspace, error) {
	return f.findByID, f.findByIDErr
}
func (f *fakeWorkspaceRepo) List(_ context.Context, _ uuid.UUID) ([]ports.Workspace, error) {
	return f.listResult, nil
}
func (f *fakeWorkspaceRepo) Update(_ context.Context, _, _ uuid.UUID, name, description string) (ports.Workspace, error) {
	f.updated = ports.Workspace{Name: name, Description: description}
	return f.updated, nil
}
func (f *fakeWorkspaceRepo) SoftDelete(_ context.Context, id, _ uuid.UUID) error {
	f.softDeleted = append(f.softDeleted, id)
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

func designer() *fakeMemberRepo {
	return &fakeMemberRepo{found: ports.Member{Role: member.RoleDesigner}}
}
func viewer() *fakeMemberRepo { return &fakeMemberRepo{found: ports.Member{Role: member.RoleViewer}} }

// --- Create ---

func TestCreate_DesignerSucceeds(t *testing.T) {
	ws := &fakeWorkspaceRepo{}
	uc := appws.NewCreateUseCase(ws, designer())
	out, err := uc.Execute(context.Background(), appws.CreateInput{
		UserID: uuid.New(), OrgID: uuid.New(), Name: "Team A", Description: "desc",
	})
	require.NoError(t, err)
	assert.Equal(t, "Team A", out.Name)
	assert.Equal(t, 1, ws.createCalls)
}

func TestCreate_ViewerForbidden(t *testing.T) {
	ws := &fakeWorkspaceRepo{}
	uc := appws.NewCreateUseCase(ws, viewer())
	_, err := uc.Execute(context.Background(), appws.CreateInput{UserID: uuid.New(), OrgID: uuid.New(), Name: "Team A"})
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
	assert.Equal(t, 0, ws.createCalls)
}

func TestCreate_NonMemberForbidden(t *testing.T) {
	ws := &fakeWorkspaceRepo{}
	uc := appws.NewCreateUseCase(ws, &fakeMemberRepo{findErr: apierr.NotFound("member")})
	_, err := uc.Execute(context.Background(), appws.CreateInput{UserID: uuid.New(), OrgID: uuid.New(), Name: "Team A"})
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}

func TestCreate_InvalidName(t *testing.T) {
	ws := &fakeWorkspaceRepo{}
	uc := appws.NewCreateUseCase(ws, designer())
	_, err := uc.Execute(context.Background(), appws.CreateInput{UserID: uuid.New(), OrgID: uuid.New(), Name: "x"})
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, appErrCode(t, err))
	assert.Equal(t, 0, ws.createCalls)
}

// --- List / Get ---

func TestList_ViewerSucceeds(t *testing.T) {
	ws := &fakeWorkspaceRepo{listResult: []ports.Workspace{{ID: uuid.New(), Name: "A"}}}
	uc := appws.NewListUseCase(ws, viewer())
	out, err := uc.Execute(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Len(t, out, 1)
}

func TestGet_NotFound(t *testing.T) {
	ws := &fakeWorkspaceRepo{findByIDErr: apierr.NotFound("workspace")}
	uc := appws.NewGetUseCase(ws, viewer())
	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
}

func TestGet_NonMemberForbidden(t *testing.T) {
	ws := &fakeWorkspaceRepo{findByID: ports.Workspace{ID: uuid.New()}}
	uc := appws.NewGetUseCase(ws, &fakeMemberRepo{findErr: apierr.NotFound("member")})
	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}

// --- Update / Delete ---

func TestUpdate_DesignerSucceeds(t *testing.T) {
	ws := &fakeWorkspaceRepo{}
	uc := appws.NewUpdateUseCase(ws, designer())
	out, err := uc.Execute(context.Background(), appws.UpdateInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), Name: "New", Description: "d",
	})
	require.NoError(t, err)
	assert.Equal(t, "New", out.Name)
}

func TestUpdate_ViewerForbidden(t *testing.T) {
	ws := &fakeWorkspaceRepo{}
	uc := appws.NewUpdateUseCase(ws, viewer())
	_, err := uc.Execute(context.Background(), appws.UpdateInput{UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), Name: "New"})
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}

func TestDelete_DesignerCascadeCallsSoftDelete(t *testing.T) {
	wid := uuid.New()
	ws := &fakeWorkspaceRepo{findByID: ports.Workspace{ID: wid}}
	uc := appws.NewDeleteUseCase(ws, designer())
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), wid)
	require.NoError(t, err)
	assert.Contains(t, ws.softDeleted, wid)
}

func TestDelete_NotFound(t *testing.T) {
	ws := &fakeWorkspaceRepo{findByIDErr: apierr.NotFound("workspace")}
	uc := appws.NewDeleteUseCase(ws, designer())
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
	assert.Empty(t, ws.softDeleted)
}

func TestDelete_ViewerForbidden(t *testing.T) {
	ws := &fakeWorkspaceRepo{findByID: ports.Workspace{ID: uuid.New()}}
	uc := appws.NewDeleteUseCase(ws, viewer())
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}
