//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/postgres"
	"github.com/novarod/polina/apps/api/internal/adapters/postgres/repository"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/hash"
)

func TestWithinTx_WorkspacesAndAPIKeys(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	store := postgres.NewStore(pool)

	org := makeOrg(t, repository.NewOrganizationRepository(pool), "tx-acme")
	mbr := makeMember(t, repository.NewMemberRepository(pool), repository.NewUserRepository(pool), org.ID, "tx-admin@example.com")

	raw := "pol_tx-raw-key"
	var wsID uuid.UUID
	err := store.WithinTx(ctx, func(r ports.Repositories) error {
		ws, err := r.Workspaces().Create(ctx, ports.Workspace{
			ID: uuid.New(), OrganizationID: org.ID, Name: "Season 1", CreatedAt: time.Now(),
		})
		if err != nil {
			return err
		}
		wsID = ws.ID

		// Uncommitted writes must be visible inside the same transaction.
		if _, err := r.Workspaces().FindByID(ctx, ws.ID, org.ID); err != nil {
			return err
		}

		if _, err := r.OrganizationAPIKeys().Create(ctx, ports.OrganizationAPIKey{
			ID: uuid.New(), OrganizationID: org.ID, Name: "tx-key", KeyHash: hash.APIKey(raw),
			CreatedByID: mbr.ID, CreatedAt: time.Now(),
		}); err != nil {
			return err
		}
		_, err = r.OrganizationAPIKeys().FindActiveByHash(ctx, hash.APIKey(raw))
		return err
	})
	require.NoError(t, err)

	// After commit both rows are visible outside the transaction.
	ws, err := repository.NewWorkspaceRepository(pool).FindByID(ctx, wsID, org.ID)
	require.NoError(t, err)
	assert.Equal(t, "Season 1", ws.Name)

	key, err := repository.NewOrganizationAPIKeyRepository(pool).FindActiveByHash(ctx, hash.APIKey(raw))
	require.NoError(t, err)
	assert.Equal(t, org.ID, key.OrganizationID)
}
