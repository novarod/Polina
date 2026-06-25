//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/postgres"
	"github.com/novarod/polina/apps/api/internal/adapters/postgres/repository"
	apporg "github.com/novarod/polina/apps/api/internal/application/organization"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

func newOrg(slug string) ports.Organization {
	return ports.Organization{ID: uuid.New(), Name: "Org " + slug, Slug: slug, CreatedAt: time.Now()}
}

func TestOrganizationRepository_CreateAndFind(t *testing.T) {
	pool := setupDB(t)
	repo := repository.NewOrganizationRepository(pool)
	ctx := context.Background()

	o := newOrg("acme")
	created, err := repo.Create(ctx, o)
	require.NoError(t, err)
	assert.Equal(t, o.ID, created.ID)

	found, err := repo.FindByID(ctx, o.ID)
	require.NoError(t, err)
	assert.Equal(t, "acme", found.Slug)
}

func TestOrganizationRepository_DuplicateSlugMaps422(t *testing.T) {
	pool := setupDB(t)
	repo := repository.NewOrganizationRepository(pool)
	ctx := context.Background()

	_, err := repo.Create(ctx, newOrg("dup"))
	require.NoError(t, err)

	_, err = repo.Create(ctx, newOrg("dup"))
	require.Error(t, err)
	var appErr *apierr.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, 422, appErr.Code)
	assert.Equal(t, "slug", appErr.Field)
}

func TestOrganizationRepository_ListByUserID(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	orgRepo := repository.NewOrganizationRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)

	org, err := orgRepo.Create(ctx, newOrg("acme"))
	require.NoError(t, err)
	u, err := userRepo.Create(ctx, newTestUser("member@example.com"))
	require.NoError(t, err)
	_, err = memberRepo.Create(ctx, ports.Member{
		ID: uuid.New(), UserID: u.ID, OrganizationID: org.ID, Role: member.RoleAdmin, CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	list, err := orgRepo.ListByUserID(ctx, u.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, org.ID, list[0].ID)
	assert.Equal(t, member.RoleAdmin, list[0].Role)
}

// TestStore_CreateOrganization_PersistsOrgAndAdminMember exercises the real
// transactional create through the use case: both the org row and the ADMIN
// member row must be present after commit.
func TestStore_CreateOrganization_PersistsOrgAndAdminMember(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	store := postgres.NewStore(pool)
	userRepo := repository.NewUserRepository(pool)

	owner, err := userRepo.Create(ctx, newTestUser("owner@example.com"))
	require.NoError(t, err)

	uc := apporg.NewCreateUseCase(store)
	org, err := uc.Execute(ctx, apporg.CreateInput{UserID: owner.ID, Name: "Acme Studios", Slug: "acme"})
	require.NoError(t, err)

	var orgCount int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM organizations WHERE id = $1 AND deleted_at IS NULL`, org.ID,
	).Scan(&orgCount))
	assert.Equal(t, 1, orgCount)

	var role string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT role FROM members WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		owner.ID, org.ID,
	).Scan(&role))
	assert.Equal(t, string(member.RoleAdmin), role)
}

// TestStore_DeleteOrganization_CascadesToMembers verifies the org and its
// members are soft-deleted atomically.
func TestStore_DeleteOrganization_CascadesToMembers(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	store := postgres.NewStore(pool)
	userRepo := repository.NewUserRepository(pool)

	owner, err := userRepo.Create(ctx, newTestUser("owner@example.com"))
	require.NoError(t, err)

	org, err := apporg.NewCreateUseCase(store).Execute(ctx, apporg.CreateInput{
		UserID: owner.ID, Name: "Acme", Slug: "acme",
	})
	require.NoError(t, err)

	delUC := apporg.NewDeleteUseCase(store)
	require.NoError(t, delUC.Execute(ctx, owner.ID, org.ID))

	var deletedOrgs int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM organizations WHERE id = $1 AND deleted_at IS NOT NULL`, org.ID,
	).Scan(&deletedOrgs))
	assert.Equal(t, 1, deletedOrgs)

	var deletedMembers int
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT count(*) FROM members WHERE organization_id = $1 AND deleted_at IS NOT NULL`, org.ID,
	).Scan(&deletedMembers))
	assert.Equal(t, 1, deletedMembers)
}
