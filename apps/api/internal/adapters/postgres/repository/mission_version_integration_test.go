//go:build integration

package repository_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/postgres"
	"github.com/novarod/polina/apps/api/internal/adapters/postgres/repository"
	appmission "github.com/novarod/polina/apps/api/internal/application/mission"
	"github.com/novarod/polina/apps/api/internal/ports"
)

const vGraphV1 = `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"END"}],"edges":[{"id":"e1","source":"n1","target":"n2"}]}`
const vGraphV2 = `{"nodes":[{"id":"n1","type":"START"},{"id":"n2","type":"OBJECTIVE","data":{"reward":5}},{"id":"n3","type":"END"}],"edges":[{"id":"e1","source":"n1","target":"n2"},{"id":"e2","source":"n2","target":"n3"}]}`

func seedMission(t *testing.T, pool *pgxpool.Pool, graph string) (org ports.Organization, ws ports.Workspace, mbr ports.Member, mission ports.Mission) {
	t.Helper()
	ctx := context.Background()
	orgRepo := repository.NewOrganizationRepository(pool)
	wsRepo := repository.NewWorkspaceRepository(pool)
	missionRepo := repository.NewMissionRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	org = makeOrg(t, orgRepo, "acme")
	mbr = makeMember(t, memberRepo, userRepo, org.ID, "designer@example.com")
	var err error
	ws, err = wsRepo.Create(ctx, ports.Workspace{ID: uuid.New(), OrganizationID: org.ID, Name: "Team", CreatedAt: time.Now()})
	require.NoError(t, err)
	mission, err = missionRepo.Create(ctx, ports.Mission{
		ID: uuid.New(), OrganizationID: org.ID, WorkspaceID: ws.ID, Name: "Quest",
		Status: "DRAFT", Graph: json.RawMessage(graph), CreatedByID: mbr.ID, CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	return org, ws, mbr, mission
}

func TestPublish_AtomicSequentialIdempotent(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	org, ws, mbr, mission := seedMission(t, pool, vGraphV1)

	store := postgres.NewStore(pool)
	publishUC := appmission.NewPublishUseCase(store)
	missionRepo := repository.NewMissionRepository(pool)
	versionRepo := repository.NewMissionVersionRepository(pool)

	in := appmission.PublishInput{UserID: mbr.UserID, OrgID: org.ID, WorkspaceID: ws.ID, MissionID: mission.ID}

	res, err := publishUC.Execute(ctx, in)
	require.NoError(t, err)
	assert.Equal(t, 1, res.Version.VersionNumber)
	require.Len(t, res.Version.Hash, 64)

	persisted, err := missionRepo.FindByID(ctx, mission.ID, org.ID, ws.ID)
	require.NoError(t, err)
	assert.Equal(t, "APPROVED", persisted.Status)
	require.NotNil(t, persisted.ActiveHash)
	assert.Equal(t, res.Version.Hash, *persisted.ActiveHash)

	stored, err := versionRepo.FindByHash(ctx, mission.ID, org.ID, res.Version.Hash)
	require.NoError(t, err)
	var contract struct {
		Version int `json:"version"`
	}
	require.NoError(t, json.Unmarshal(stored.MissionData, &contract))
	assert.Equal(t, 1, contract.Version, "version_number injected into mission_data")

	res2, err := publishUC.Execute(ctx, in)
	require.NoError(t, err)
	assert.Equal(t, res.Version.Hash, res2.Version.Hash)
	assert.Equal(t, 1, res2.Version.VersionNumber)

	_, err = missionRepo.UpdateGraph(ctx, mission.ID, org.ID, ws.ID, json.RawMessage(vGraphV2))
	require.NoError(t, err)
	res3, err := publishUC.Execute(ctx, in)
	require.NoError(t, err)
	assert.Equal(t, 2, res3.Version.VersionNumber)
	assert.NotEqual(t, res.Version.Hash, res3.Version.Hash)

	list, err := versionRepo.List(ctx, mission.ID, org.ID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, 2, list[0].VersionNumber)
	assert.Equal(t, 1, list[1].VersionNumber)
}

func TestPublish_VersionTenantIsolation(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	org, ws, mbr, mission := seedMission(t, pool, vGraphV1)
	otherOrg := makeOrg(t, repository.NewOrganizationRepository(pool), "other")

	store := postgres.NewStore(pool)
	publishUC := appmission.NewPublishUseCase(store)
	versionRepo := repository.NewMissionVersionRepository(pool)

	res, err := publishUC.Execute(ctx, appmission.PublishInput{
		UserID: mbr.UserID, OrgID: org.ID, WorkspaceID: ws.ID, MissionID: mission.ID,
	})
	require.NoError(t, err)

	_, err = versionRepo.FindByHash(ctx, mission.ID, otherOrg.ID, res.Version.Hash)
	require.Error(t, err)
	empty, err := versionRepo.List(ctx, mission.ID, otherOrg.ID)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestPublish_ConcurrentSameGraph_Idempotent(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	org, ws, mbr, mission := seedMission(t, pool, vGraphV1)

	store := postgres.NewStore(pool)
	publishUC := appmission.NewPublishUseCase(store)
	versionRepo := repository.NewMissionVersionRepository(pool)

	const n = 8
	in := appmission.PublishInput{UserID: mbr.UserID, OrgID: org.ID, WorkspaceID: ws.ID, MissionID: mission.ID}

	var wg sync.WaitGroup
	results := make([]appmission.PublishResult, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = publishUC.Execute(ctx, in)
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		require.NoErrorf(t, errs[i], "publish %d must not error", i)
		assert.Equal(t, 1, results[i].Version.VersionNumber, "all reuse version 1")
		assert.Equal(t, results[0].Version.Hash, results[i].Version.Hash, "same hash")
	}
	list, err := versionRepo.List(ctx, mission.ID, org.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1, "concurrent identical publishes must create exactly one version")
}

func TestPublish_ConcurrentDifferentGraphs_Sequential(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	org, ws, mbr, mission := seedMission(t, pool, vGraphV1)

	store := postgres.NewStore(pool)
	publishUC := appmission.NewPublishUseCase(store)
	missionRepo := repository.NewMissionRepository(pool)
	versionRepo := repository.NewMissionVersionRepository(pool)
	in := appmission.PublishInput{UserID: mbr.UserID, OrgID: org.ID, WorkspaceID: ws.ID, MissionID: mission.ID}

	var wg sync.WaitGroup
	var err1, err2 error
	wg.Add(2)
	go func() { defer wg.Done(); _, err1 = publishUC.Execute(ctx, in) }()
	go func() {
		defer wg.Done()
		if _, e := missionRepo.UpdateGraph(ctx, mission.ID, org.ID, ws.ID, json.RawMessage(vGraphV2)); e != nil {
			err2 = e
			return
		}
		_, err2 = publishUC.Execute(ctx, in)
	}()
	wg.Wait()
	require.NoError(t, err1)
	require.NoError(t, err2)

	list, err := versionRepo.List(ctx, mission.ID, org.ID)
	require.NoError(t, err)
	require.NotEmpty(t, list)
	seen := map[int]bool{}
	for _, v := range list {
		assert.Falsef(t, seen[v.VersionNumber], "duplicate version_number %d", v.VersionNumber)
		seen[v.VersionNumber] = true
	}
	assert.True(t, seen[1], "version 1 exists")
	if len(list) == 2 {
		assert.True(t, seen[2], "second version is 2, sequential")
	}
}
