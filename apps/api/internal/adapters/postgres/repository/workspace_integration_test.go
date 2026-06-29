//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/postgres/repository"
	"github.com/novarod/polina/apps/api/internal/ports"
)

func makeOrg(t *testing.T, repo *repository.OrganizationRepository, slug string) ports.Organization {
	t.Helper()
	o, err := repo.Create(context.Background(), ports.Organization{
		ID: uuid.New(), Name: "Org " + slug, Slug: slug, CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	return o
}

func TestWorkspaceRepository_CRUD(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	orgRepo := repository.NewOrganizationRepository(pool)
	wsRepo := repository.NewWorkspaceRepository(pool)

	org := makeOrg(t, orgRepo, "acme")

	created, err := wsRepo.Create(ctx, ports.Workspace{
		ID: uuid.New(), OrganizationID: org.ID, Name: "Team A", Description: "first", CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	assert.Equal(t, "Team A", created.Name)

	found, err := wsRepo.FindByID(ctx, created.ID, org.ID)
	require.NoError(t, err)
	assert.Equal(t, "first", found.Description)

	list, err := wsRepo.List(ctx, org.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	updated, err := wsRepo.Update(ctx, created.ID, org.ID, "Team B", "second")
	require.NoError(t, err)
	assert.Equal(t, "Team B", updated.Name)
	assert.Equal(t, "second", updated.Description)

	require.NoError(t, wsRepo.SoftDelete(ctx, created.ID, org.ID))

	_, err = wsRepo.FindByID(ctx, created.ID, org.ID)
	require.Error(t, err, "soft-deleted workspace must not be found")

	empty, err := wsRepo.List(ctx, org.ID)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

// TestWorkspaceRepository_TenantIsolation: a workspace created in org A must not
// be reachable through org B's id.
func TestWorkspaceRepository_TenantIsolation(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	orgRepo := repository.NewOrganizationRepository(pool)
	wsRepo := repository.NewWorkspaceRepository(pool)

	orgA := makeOrg(t, orgRepo, "org-a")
	orgB := makeOrg(t, orgRepo, "org-b")

	ws, err := wsRepo.Create(ctx, ports.Workspace{
		ID: uuid.New(), OrganizationID: orgA.ID, Name: "A team", CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	// Same id, wrong tenant → not found.
	_, err = wsRepo.FindByID(ctx, ws.ID, orgB.ID)
	require.Error(t, err)

	// org B sees no workspaces.
	listB, err := wsRepo.List(ctx, orgB.ID)
	require.NoError(t, err)
	assert.Empty(t, listB)
}
