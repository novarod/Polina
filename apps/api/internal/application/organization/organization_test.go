package organization_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apporg "github.com/novarod/polina/apps/api/internal/application/organization"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

// --- fakes ---

var (
	_ ports.OrganizationRepository = (*fakeOrgRepo)(nil)
	_ ports.MemberRepository       = (*fakeMemberRepo)(nil)
	_ ports.TxManager              = (*fakeTxManager)(nil)
	_ ports.Repositories           = (*fakeRepos)(nil)
)

type fakeOrgRepo struct {
	createErr   error
	createCalls int
	findByID    ports.Organization
	findByIDErr error
	listResult  []ports.OrganizationWithRole
	updated     ports.Organization
	updateErr   error
	softDeleted []uuid.UUID
}

func (f *fakeOrgRepo) Create(_ context.Context, o ports.Organization) (ports.Organization, error) {
	f.createCalls++
	if f.createErr != nil {
		return ports.Organization{}, f.createErr
	}
	return o, nil
}
func (f *fakeOrgRepo) FindByID(_ context.Context, _ uuid.UUID) (ports.Organization, error) {
	if f.findByIDErr != nil {
		return ports.Organization{}, f.findByIDErr
	}
	return f.findByID, nil
}
func (f *fakeOrgRepo) ListByUserID(_ context.Context, _ uuid.UUID) ([]ports.OrganizationWithRole, error) {
	return f.listResult, nil
}
func (f *fakeOrgRepo) Update(_ context.Context, _ uuid.UUID, name string) (ports.Organization, error) {
	if f.updateErr != nil {
		return ports.Organization{}, f.updateErr
	}
	f.updated.Name = name
	return f.updated, nil
}
func (f *fakeOrgRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	f.softDeleted = append(f.softDeleted, id)
	return nil
}

type fakeMemberRepo struct {
	found            ports.Member
	findErr          error
	created          *ports.Member
	softDeletedByOrg []uuid.UUID
}

func (f *fakeMemberRepo) Create(_ context.Context, m ports.Member) (ports.Member, error) {
	cp := m
	f.created = &cp
	return m, nil
}
func (f *fakeMemberRepo) FindByUserAndOrg(_ context.Context, _, _ uuid.UUID) (ports.Member, error) {
	if f.findErr != nil {
		return ports.Member{}, f.findErr
	}
	return f.found, nil
}
func (f *fakeMemberRepo) FindByID(_ context.Context, _ uuid.UUID) (ports.Member, error) {
	return ports.Member{}, nil
}
func (f *fakeMemberRepo) ListByOrg(_ context.Context, _ uuid.UUID, _, _ int) ([]ports.Member, int, error) {
	return nil, 0, nil
}
func (f *fakeMemberRepo) UpdateRole(_ context.Context, _ uuid.UUID, _ member.Role) (ports.Member, error) {
	return ports.Member{}, nil
}
func (f *fakeMemberRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeMemberRepo) SoftDeleteByOrg(_ context.Context, orgID uuid.UUID) error {
	f.softDeletedByOrg = append(f.softDeletedByOrg, orgID)
	return nil
}

type fakeRepos struct {
	orgs    ports.OrganizationRepository
	members ports.MemberRepository
}

func (f *fakeRepos) Users() ports.UserRepository                 { return nil }
func (f *fakeRepos) Members() ports.MemberRepository             { return f.members }
func (f *fakeRepos) Organizations() ports.OrganizationRepository { return f.orgs }

type fakeTxManager struct{ repos ports.Repositories }

func (f *fakeTxManager) WithinTx(ctx context.Context, fn func(ports.Repositories) error) error {
	return fn(f.repos)
}

func appErrCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *apierr.AppError
	require.ErrorAs(t, err, &appErr)
	return appErr.Code
}

// --- Create ---

func TestCreate_Success_MakesCreatorAdmin(t *testing.T) {
	orgs := &fakeOrgRepo{}
	members := &fakeMemberRepo{}
	tx := &fakeTxManager{repos: &fakeRepos{orgs: orgs, members: members}}
	uc := apporg.NewCreateUseCase(tx)

	userID := uuid.New()
	org, err := uc.Execute(context.Background(), apporg.CreateInput{UserID: userID, Name: "Acme Studios", Slug: "acme"})

	require.NoError(t, err)
	assert.Equal(t, "Acme Studios", org.Name)
	assert.Equal(t, "acme", org.Slug)
	assert.Equal(t, 1, orgs.createCalls)
	require.NotNil(t, members.created, "creator must be added as a member")
	assert.Equal(t, member.RoleAdmin, members.created.Role)
	assert.Equal(t, userID, members.created.UserID)
	assert.Equal(t, org.ID, members.created.OrganizationID)
}

func TestCreate_DuplicateSlug(t *testing.T) {
	orgs := &fakeOrgRepo{createErr: apierr.Validation("slug", "slug already in use")}
	members := &fakeMemberRepo{}
	tx := &fakeTxManager{repos: &fakeRepos{orgs: orgs, members: members}}
	uc := apporg.NewCreateUseCase(tx)

	_, err := uc.Execute(context.Background(), apporg.CreateInput{UserID: uuid.New(), Name: "Acme", Slug: "acme"})

	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, appErrCode(t, err))
	assert.Nil(t, members.created, "no member should be created when org creation fails")
}

func TestCreate_InvalidSlug_RejectedBeforeTx(t *testing.T) {
	orgs := &fakeOrgRepo{}
	tx := &fakeTxManager{repos: &fakeRepos{orgs: orgs, members: &fakeMemberRepo{}}}
	uc := apporg.NewCreateUseCase(tx)

	_, err := uc.Execute(context.Background(), apporg.CreateInput{UserID: uuid.New(), Name: "Acme", Slug: "Invalid Slug!"})

	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, appErrCode(t, err))
	assert.Equal(t, 0, orgs.createCalls, "invalid slug must be rejected before touching the repo")
}

func TestCreate_InvalidName(t *testing.T) {
	tx := &fakeTxManager{repos: &fakeRepos{orgs: &fakeOrgRepo{}, members: &fakeMemberRepo{}}}
	uc := apporg.NewCreateUseCase(tx)

	_, err := uc.Execute(context.Background(), apporg.CreateInput{UserID: uuid.New(), Name: "A", Slug: "acme"})

	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, appErrCode(t, err))
}

// --- List ---

func TestList_MapsRole(t *testing.T) {
	orgID := uuid.New()
	orgs := &fakeOrgRepo{listResult: []ports.OrganizationWithRole{
		{Organization: ports.Organization{ID: orgID, Name: "Acme", Slug: "acme"}, Role: member.RoleAdmin},
	}}
	uc := apporg.NewListUseCase(orgs)

	items, err := uc.Execute(context.Background(), uuid.New())

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, orgID, items[0].ID)
	assert.Equal(t, member.RoleAdmin, items[0].Role)
}

// --- Get ---

func TestGet_MemberSucceeds(t *testing.T) {
	orgID := uuid.New()
	orgs := &fakeOrgRepo{findByID: ports.Organization{ID: orgID, Name: "Acme", Slug: "acme"}}
	members := &fakeMemberRepo{found: ports.Member{Role: member.RoleViewer}}
	uc := apporg.NewGetUseCase(orgs, members)

	org, err := uc.Execute(context.Background(), uuid.New(), orgID)

	require.NoError(t, err)
	assert.Equal(t, orgID, org.ID)
}

func TestGet_NotFound(t *testing.T) {
	orgs := &fakeOrgRepo{findByIDErr: apierr.NotFound("organization")}
	uc := apporg.NewGetUseCase(orgs, &fakeMemberRepo{})

	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New())

	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
}

func TestGet_NotAMember(t *testing.T) {
	orgs := &fakeOrgRepo{findByID: ports.Organization{ID: uuid.New()}}
	members := &fakeMemberRepo{findErr: apierr.NotFound("member")}
	uc := apporg.NewGetUseCase(orgs, members)

	_, err := uc.Execute(context.Background(), uuid.New(), uuid.New())

	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}

// --- Update ---

func TestUpdate_AdminChangesName(t *testing.T) {
	orgID := uuid.New()
	orgs := &fakeOrgRepo{
		findByID: ports.Organization{ID: orgID, Name: "Old", Slug: "acme"},
		updated:  ports.Organization{ID: orgID, Slug: "acme"},
	}
	members := &fakeMemberRepo{found: ports.Member{Role: member.RoleAdmin}}
	uc := apporg.NewUpdateUseCase(orgs, members)

	org, err := uc.Execute(context.Background(), uuid.New(), orgID, "New Name")

	require.NoError(t, err)
	assert.Equal(t, "New Name", org.Name)
}

func TestUpdate_NonAdminForbidden(t *testing.T) {
	orgID := uuid.New()
	orgs := &fakeOrgRepo{findByID: ports.Organization{ID: orgID}}
	members := &fakeMemberRepo{found: ports.Member{Role: member.RoleViewer}}
	uc := apporg.NewUpdateUseCase(orgs, members)

	_, err := uc.Execute(context.Background(), uuid.New(), orgID, "New Name")

	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}

// --- Delete ---

func TestDelete_AdminCascades(t *testing.T) {
	orgID := uuid.New()
	orgs := &fakeOrgRepo{findByID: ports.Organization{ID: orgID}}
	members := &fakeMemberRepo{found: ports.Member{Role: member.RoleAdmin}}
	tx := &fakeTxManager{repos: &fakeRepos{orgs: orgs, members: members}}
	uc := apporg.NewDeleteUseCase(tx)

	err := uc.Execute(context.Background(), uuid.New(), orgID)

	require.NoError(t, err)
	assert.Contains(t, orgs.softDeleted, orgID)
	assert.Contains(t, members.softDeletedByOrg, orgID)
}

func TestDelete_NonAdminForbidden(t *testing.T) {
	orgID := uuid.New()
	orgs := &fakeOrgRepo{findByID: ports.Organization{ID: orgID}}
	members := &fakeMemberRepo{found: ports.Member{Role: member.RoleDesigner}}
	tx := &fakeTxManager{repos: &fakeRepos{orgs: orgs, members: members}}
	uc := apporg.NewDeleteUseCase(tx)

	err := uc.Execute(context.Background(), uuid.New(), orgID)

	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
	assert.Empty(t, orgs.softDeleted, "nothing should be deleted when authorization fails")
}

func TestDelete_NotFound(t *testing.T) {
	orgs := &fakeOrgRepo{findByIDErr: apierr.NotFound("organization")}
	tx := &fakeTxManager{repos: &fakeRepos{orgs: orgs, members: &fakeMemberRepo{}}}
	uc := apporg.NewDeleteUseCase(tx)

	err := uc.Execute(context.Background(), uuid.New(), uuid.New())

	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, appErrCode(t, err))
}
