package mission_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appmission "github.com/novarod/polina/apps/api/internal/application/mission"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

var (
	_ ports.MissionRepository   = (*fakeMissionRepo)(nil)
	_ ports.WorkspaceRepository = (*fakeWorkspaceRepo)(nil)
	_ ports.MemberRepository    = (*fakeMemberRepo)(nil)
)

const validGraph = `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"END"}],"edges":[{"id":"e1","source":"n1","target":"n2"}]}`
const cyclicGraph = `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"OBJECTIVE"},{"id":"n3","type":"END"}],"edges":[{"id":"e1","source":"n1","target":"n2"},{"id":"e2","source":"n2","target":"n3"},{"id":"e3","source":"n3","target":"n2"}]}`

type fakeMissionRepo struct {
	created      ports.Mission
	createCalls  int
	findByID     ports.Mission
	findByIDErr  error
	updatedGraph json.RawMessage
	updateGCalls int
	softDeleted  []uuid.UUID
	activeHash   string
	activeCalls  int
}

func (f *fakeMissionRepo) Create(_ context.Context, m ports.Mission) (ports.Mission, error) {
	f.createCalls++
	f.created = m
	return m, nil
}
func (f *fakeMissionRepo) FindByID(_ context.Context, _, _, _ uuid.UUID) (ports.Mission, error) {
	return f.findByID, f.findByIDErr
}
func (f *fakeMissionRepo) FindByIDForUpdate(_ context.Context, _, _, _ uuid.UUID) (ports.Mission, error) {
	return f.findByID, f.findByIDErr
}
func (f *fakeMissionRepo) List(_ context.Context, _, _ uuid.UUID) ([]ports.Mission, error) {
	return nil, nil
}
func (f *fakeMissionRepo) UpdateGraph(_ context.Context, _, _, _ uuid.UUID, graph json.RawMessage) (ports.Mission, error) {
	f.updateGCalls++
	f.updatedGraph = graph
	return ports.Mission{Graph: graph}, nil
}
func (f *fakeMissionRepo) Update(_ context.Context, _, _, _ uuid.UUID, name, description string) (ports.Mission, error) {
	return ports.Mission{Name: name, Description: description}, nil
}
func (f *fakeMissionRepo) SetActiveVersion(_ context.Context, _, _, _ uuid.UUID, hash, status string) (ports.Mission, error) {
	f.activeCalls++
	f.activeHash = hash
	m := f.findByID
	m.ActiveHash = &hash
	m.Status = status
	return m, nil
}
func (f *fakeMissionRepo) SoftDelete(_ context.Context, id, _, _ uuid.UUID) error {
	f.softDeleted = append(f.softDeleted, id)
	return nil
}

type fakeWorkspaceRepo struct{ findErr error }

func (f *fakeWorkspaceRepo) Create(_ context.Context, w ports.Workspace) (ports.Workspace, error) {
	return w, nil
}
func (f *fakeWorkspaceRepo) FindByID(_ context.Context, _, _ uuid.UUID) (ports.Workspace, error) {
	return ports.Workspace{}, f.findErr
}
func (f *fakeWorkspaceRepo) List(_ context.Context, _ uuid.UUID) ([]ports.Workspace, error) {
	return nil, nil
}
func (f *fakeWorkspaceRepo) Update(_ context.Context, _, _ uuid.UUID, _, _ string) (ports.Workspace, error) {
	return ports.Workspace{}, nil
}
func (f *fakeWorkspaceRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return nil }

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

func designer(id uuid.UUID) *fakeMemberRepo {
	return &fakeMemberRepo{found: ports.Member{ID: id, Role: member.RoleDesigner}}
}
func viewer() *fakeMemberRepo { return &fakeMemberRepo{found: ports.Member{Role: member.RoleViewer}} }

// --- Create ---

func TestCreate_DesignerSucceeds_SetsCreatorAndDraft(t *testing.T) {
	memberID := uuid.New()
	mr := &fakeMissionRepo{}
	uc := appmission.NewCreateUseCase(mr, &fakeWorkspaceRepo{}, designer(memberID))
	m, err := uc.Execute(context.Background(), appmission.CreateInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), Name: "Old Country",
	})
	require.NoError(t, err)
	assert.Equal(t, "Old Country", m.Name)
	assert.Equal(t, memberID, mr.created.CreatedByID)
	assert.Equal(t, "DRAFT", mr.created.Status)
	assert.JSONEq(t, `{"nodes":[],"edges":[]}`, string(mr.created.Graph))
}

func TestCreate_ParentWorkspaceMissing_404(t *testing.T) {
	mr := &fakeMissionRepo{}
	uc := appmission.NewCreateUseCase(mr, &fakeWorkspaceRepo{findErr: apierr.NotFound("workspace")}, designer(uuid.New()))
	_, err := uc.Execute(context.Background(), appmission.CreateInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), Name: "X",
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
	assert.Equal(t, 0, mr.createCalls)
}

func TestCreate_ViewerForbidden(t *testing.T) {
	mr := &fakeMissionRepo{}
	uc := appmission.NewCreateUseCase(mr, &fakeWorkspaceRepo{}, viewer())
	_, err := uc.Execute(context.Background(), appmission.CreateInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), Name: "X",
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
	assert.Equal(t, 0, mr.createCalls)
}

func TestCreate_NonMemberForbidden(t *testing.T) {
	uc := appmission.NewCreateUseCase(&fakeMissionRepo{}, &fakeWorkspaceRepo{}, &fakeMemberRepo{findErr: apierr.NotFound("member")})
	_, err := uc.Execute(context.Background(), appmission.CreateInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), Name: "X",
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}

// --- UpdateGraph ---

func TestUpdateGraph_ValidSaves(t *testing.T) {
	mr := &fakeMissionRepo{}
	uc := appmission.NewUpdateGraphUseCase(mr, designer(uuid.New()))
	_, err := uc.Execute(context.Background(), appmission.UpdateGraphInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), MissionID: uuid.New(),
		Graph: json.RawMessage(validGraph),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, mr.updateGCalls)
}

func TestUpdateGraph_CyclicRejected422(t *testing.T) {
	mr := &fakeMissionRepo{}
	uc := appmission.NewUpdateGraphUseCase(mr, designer(uuid.New()))
	_, err := uc.Execute(context.Background(), appmission.UpdateGraphInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), MissionID: uuid.New(),
		Graph: json.RawMessage(cyclicGraph),
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, appErrCode(t, err))
	assert.Equal(t, 0, mr.updateGCalls, "invalid graph must not be persisted")
}

func TestUpdateGraph_ViewerForbidden(t *testing.T) {
	mr := &fakeMissionRepo{}
	uc := appmission.NewUpdateGraphUseCase(mr, viewer())
	_, err := uc.Execute(context.Background(), appmission.UpdateGraphInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), MissionID: uuid.New(),
		Graph: json.RawMessage(validGraph),
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
	assert.Equal(t, 0, mr.updateGCalls)
}

// --- Get / Delete ---

func TestGet_NotFound(t *testing.T) {
	uc := appmission.NewGetUseCase(&fakeMissionRepo{findByIDErr: apierr.NotFound("mission")}, viewer())
	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
}

func TestDelete_DesignerSoftDeletes(t *testing.T) {
	mid := uuid.New()
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: mid}}
	uc := appmission.NewDeleteUseCase(mr, designer(uuid.New()))
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New(), mid)
	require.NoError(t, err)
	assert.Contains(t, mr.softDeleted, mid)
}

func TestDelete_NotFound(t *testing.T) {
	mr := &fakeMissionRepo{findByIDErr: apierr.NotFound("mission")}
	uc := appmission.NewDeleteUseCase(mr, designer(uuid.New()))
	err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
	assert.Empty(t, mr.softDeleted)
}
