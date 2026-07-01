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
	appengine "github.com/novarod/polina/apps/api/internal/application/engine"
	appmission "github.com/novarod/polina/apps/api/internal/application/mission"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/hash"
)

func TestOrganizationAPIKeyRepository_CRUDAndTouch(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	orgRepo := repository.NewOrganizationRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	keyRepo := repository.NewOrganizationAPIKeyRepository(pool)

	org := makeOrg(t, orgRepo, "acme")
	mbr := makeMember(t, memberRepo, userRepo, org.ID, "admin@example.com")

	raw := "pol_integration-raw-key"
	created, err := keyRepo.Create(ctx, ports.OrganizationAPIKey{
		ID: uuid.New(), OrganizationID: org.ID, Name: "CI", KeyHash: hash.APIKey(raw),
		CreatedByID: mbr.ID, CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	assert.Nil(t, created.LastUsedAt)

	// FindActiveByHash resolves the org.
	found, err := keyRepo.FindActiveByHash(ctx, hash.APIKey(raw))
	require.NoError(t, err)
	assert.Equal(t, org.ID, found.OrganizationID)

	// List returns the key.
	list, err := keyRepo.ListByOrg(ctx, org.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "CI", list[0].Name)

	// Touch sets last_used_at (throttle 0 → always writes).
	require.NoError(t, keyRepo.TouchLastUsed(ctx, created.ID, 0))
	afterTouch, err := keyRepo.ListByOrg(ctx, org.ID)
	require.NoError(t, err)
	require.NotNil(t, afterTouch[0].LastUsedAt)
	firstUse := *afterTouch[0].LastUsedAt

	// A second touch under a large throttle is a no-op (last_used_at unchanged).
	require.NoError(t, keyRepo.TouchLastUsed(ctx, created.ID, time.Hour))
	afterNoop, err := keyRepo.ListByOrg(ctx, org.ID)
	require.NoError(t, err)
	require.NotNil(t, afterNoop[0].LastUsedAt)
	assert.WithinDuration(t, firstUse, *afterNoop[0].LastUsedAt, time.Millisecond, "throttled touch is a no-op")

	// Revoke → the key no longer authenticates.
	require.NoError(t, keyRepo.Revoke(ctx, created.ID, org.ID))
	_, err = keyRepo.FindActiveByHash(ctx, hash.APIKey(raw))
	require.Error(t, err, "revoked key must not resolve")

	// Revoking again → NotFound.
	require.Error(t, keyRepo.Revoke(ctx, created.ID, org.ID))
}

func TestEngine_ActiveFlow_And_Isolation(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	org, ws, mbr, mission := seedMission(t, pool, vGraphV1)
	otherOrg := makeOrg(t, repository.NewOrganizationRepository(pool), "other")

	store := postgres.NewStore(pool)
	publishUC := appmission.NewPublishUseCase(store)
	hashUC := appengine.NewGetActiveHashUseCase(repository.NewMissionRepository(pool))
	contractUC := appengine.NewGetActiveContractUseCase(repository.NewMissionVersionRepository(pool))

	// Before publish: engine sees no active version.
	_, err := hashUC.Execute(ctx, org.ID, mission.ID)
	require.Error(t, err)

	res, err := publishUC.Execute(ctx, appmission.PublishInput{
		UserID: mbr.UserID, OrgID: org.ID, WorkspaceID: ws.ID, MissionID: mission.ID,
	})
	require.NoError(t, err)

	// Active hash matches the published version.
	h, err := hashUC.Execute(ctx, org.ID, mission.ID)
	require.NoError(t, err)
	assert.Equal(t, res.Version.Hash, h)

	// Active contract is the published mission_data.
	contract, err := contractUC.Execute(ctx, org.ID, mission.ID)
	require.NoError(t, err)
	assert.JSONEq(t, string(res.Version.MissionData), string(contract))

	// Isolation: another org's key cannot read this mission.
	_, err = hashUC.Execute(ctx, otherOrg.ID, mission.ID)
	require.Error(t, err)
	_, err = contractUC.Execute(ctx, otherOrg.ID, mission.ID)
	require.Error(t, err)
}
