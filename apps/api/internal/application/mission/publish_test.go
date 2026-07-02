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
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

var _ ports.MissionVersionRepository = (*fakeVersionRepo)(nil)

// fakeVersionRepo records writes and can be primed with an existing version to
// exercise the idempotent (hash-already-present) path.
type fakeVersionRepo struct {
	existing    *ports.MissionVersion
	createCalls int
	created     ports.MissionVersion
	listed      []ports.MissionVersion
	findErr     error
}

func (f *fakeVersionRepo) Create(_ context.Context, v ports.MissionVersion) (ports.MissionVersion, error) {
	f.createCalls++
	v.VersionNumber = 1
	f.created = v
	return v, nil
}
func (f *fakeVersionRepo) FindByHash(_ context.Context, _, _ uuid.UUID, _ string) (ports.MissionVersion, error) {
	if f.findErr != nil {
		return ports.MissionVersion{}, f.findErr
	}
	if f.existing != nil {
		return *f.existing, nil
	}
	return ports.MissionVersion{}, apierr.NotFound("mission version")
}
func (f *fakeVersionRepo) List(_ context.Context, _, _ uuid.UUID) ([]ports.MissionVersion, error) {
	return f.listed, nil
}
func (f *fakeVersionRepo) FindActive(_ context.Context, _, _ uuid.UUID) (ports.MissionVersion, error) {
	if f.existing != nil {
		return *f.existing, nil
	}
	return ports.MissionVersion{}, apierr.NotFound("active mission version")
}

// fakeTxManager runs fn against a fixed set of repositories (no real transaction).
type fakeTxManager struct{ repos ports.Repositories }

func (t *fakeTxManager) WithinTx(ctx context.Context, fn func(ports.Repositories) error) error {
	return fn(t.repos)
}

// fakeRepos exposes only the repositories the mission publish flow touches.
type fakeRepos struct {
	members  ports.MemberRepository
	missions ports.MissionRepository
	versions ports.MissionVersionRepository
}

func (r *fakeRepos) Users() ports.UserRepository                 { return nil }
func (r *fakeRepos) Members() ports.MemberRepository             { return r.members }
func (r *fakeRepos) Organizations() ports.OrganizationRepository { return nil }
func (r *fakeRepos) Missions() ports.MissionRepository           { return r.missions }
func (r *fakeRepos) MissionVersions() ports.MissionVersionRepository {
	return r.versions
}
func (r *fakeRepos) Workspaces() ports.WorkspaceRepository { return nil }
func (r *fakeRepos) OrganizationAPIKeys() ports.OrganizationAPIKeyRepository {
	return nil
}

func publishTx(members ports.MemberRepository, missions ports.MissionRepository, versions ports.MissionVersionRepository) *fakeTxManager {
	return &fakeTxManager{repos: &fakeRepos{members: members, missions: missions, versions: versions}}
}

// --- Publish ---

func TestPublish_ValidGraph_CreatesVersionAndActivates(t *testing.T) {
	memberID := uuid.New()
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New(), Graph: json.RawMessage(validGraph)}}
	vr := &fakeVersionRepo{}
	uc := appmission.NewPublishUseCase(publishTx(designer(memberID), mr, vr))

	res, err := uc.Execute(context.Background(), appmission.PublishInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), MissionID: uuid.New(),
	})
	require.NoError(t, err)
	assert.Equal(t, 1, vr.createCalls)
	assert.Equal(t, memberID, vr.created.PublishedByID)
	assert.Len(t, vr.created.Hash, 64)
	assert.Equal(t, 1, mr.activeCalls)
	assert.Equal(t, vr.created.Hash, mr.activeHash, "mission points at the published hash")
	assert.Equal(t, "APPROVED", res.Mission.Status)
	require.NotNil(t, res.Mission.ActiveHash)
	assert.Equal(t, vr.created.Hash, *res.Mission.ActiveHash)
}

func TestPublish_Idempotent_ReusesExistingVersion(t *testing.T) {
	existing := ports.MissionVersion{ID: uuid.New(), VersionNumber: 3, Hash: "deadbeef"}
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New(), Graph: json.RawMessage(validGraph)}}
	vr := &fakeVersionRepo{existing: &existing}
	uc := appmission.NewPublishUseCase(publishTx(designer(uuid.New()), mr, vr))

	res, err := uc.Execute(context.Background(), appmission.PublishInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), MissionID: uuid.New(),
	})
	require.NoError(t, err)
	assert.Equal(t, 0, vr.createCalls, "matching hash must not create a new version")
	assert.Equal(t, 3, res.Version.VersionNumber, "reuses the existing version")
	assert.Equal(t, 1, mr.activeCalls, "still reaffirms the active version")
}

func TestPublish_InvalidGraph_422_NoWrite(t *testing.T) {
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New(), Graph: json.RawMessage(cyclicGraph)}}
	vr := &fakeVersionRepo{}
	uc := appmission.NewPublishUseCase(publishTx(designer(uuid.New()), mr, vr))

	_, err := uc.Execute(context.Background(), appmission.PublishInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), MissionID: uuid.New(),
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, appErrCode(t, err))
	assert.Equal(t, 0, vr.createCalls)
	assert.Equal(t, 0, mr.activeCalls)
}

func TestPublish_ViewerForbidden(t *testing.T) {
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New(), Graph: json.RawMessage(validGraph)}}
	vr := &fakeVersionRepo{}
	uc := appmission.NewPublishUseCase(publishTx(viewer(), mr, vr))

	_, err := uc.Execute(context.Background(), appmission.PublishInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), MissionID: uuid.New(),
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
	assert.Equal(t, 0, vr.createCalls)
}

func TestPublish_MissionNotFound_404(t *testing.T) {
	mr := &fakeMissionRepo{findByIDErr: apierr.NotFound("mission")}
	vr := &fakeVersionRepo{}
	uc := appmission.NewPublishUseCase(publishTx(designer(uuid.New()), mr, vr))

	_, err := uc.Execute(context.Background(), appmission.PublishInput{
		UserID: uuid.New(), OrgID: uuid.New(), WorkspaceID: uuid.New(), MissionID: uuid.New(),
	})
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
	assert.Equal(t, 0, vr.createCalls)
}

// --- List / Get versions ---

func TestListVersions_ViewerSucceeds(t *testing.T) {
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New()}}
	vr := &fakeVersionRepo{listed: []ports.MissionVersion{{VersionNumber: 2}, {VersionNumber: 1}}}
	uc := appmission.NewListVersionsUseCase(mr, vr, viewer())

	list, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestListVersions_MissionNotFound_404(t *testing.T) {
	mr := &fakeMissionRepo{findByIDErr: apierr.NotFound("mission")}
	uc := appmission.NewListVersionsUseCase(mr, &fakeVersionRepo{}, viewer())

	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
}

func TestGetVersion_HashNotFound_404(t *testing.T) {
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New()}}
	vr := &fakeVersionRepo{findErr: apierr.NotFound("mission version")}
	uc := appmission.NewGetVersionUseCase(mr, vr, viewer())

	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), "nope")
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
}

func TestGetVersion_NonMemberForbidden(t *testing.T) {
	mr := &fakeMissionRepo{findByID: ports.Mission{ID: uuid.New()}}
	uc := appmission.NewGetVersionUseCase(mr, &fakeVersionRepo{}, &fakeMemberRepo{findErr: apierr.NotFound("member")})

	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), "h")
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}
