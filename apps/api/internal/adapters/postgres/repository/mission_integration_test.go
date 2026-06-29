//go:build integration

package repository_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/postgres/repository"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
)

func makeMember(t *testing.T, mr *repository.MemberRepository, ur *repository.UserRepository, orgID uuid.UUID, email string) ports.Member {
	t.Helper()
	u, err := ur.Create(context.Background(), newTestUser(email))
	require.NoError(t, err)
	m, err := mr.Create(context.Background(), ports.Member{
		ID: uuid.New(), UserID: u.ID, OrganizationID: orgID, Role: member.RoleAdmin, CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	return m
}

func TestMissionRepository_CRUD(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	orgRepo := repository.NewOrganizationRepository(pool)
	wsRepo := repository.NewWorkspaceRepository(pool)
	missionRepo := repository.NewMissionRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	org := makeOrg(t, orgRepo, "acme")
	mbr := makeMember(t, memberRepo, userRepo, org.ID, "designer@example.com")
	ws, err := wsRepo.Create(ctx, ports.Workspace{ID: uuid.New(), OrganizationID: org.ID, Name: "Team A", CreatedAt: time.Now()})
	require.NoError(t, err)

	created, err := missionRepo.Create(ctx, ports.Mission{
		ID: uuid.New(), OrganizationID: org.ID, WorkspaceID: ws.ID, Name: "Old Country",
		Status: "DRAFT", Graph: json.RawMessage(`{"nodes":[],"edges":[]}`), CreatedByID: mbr.ID, CreatedAt: time.Now(),
	})
	require.NoError(t, err)
	assert.Equal(t, "DRAFT", created.Status)

	found, err := missionRepo.FindByID(ctx, created.ID, org.ID, ws.ID)
	require.NoError(t, err)
	assert.Equal(t, "Old Country", found.Name)

	list, err := missionRepo.List(ctx, ws.ID, org.ID)
	require.NoError(t, err)
	assert.Len(t, list, 1)

	graph := json.RawMessage(`{"nodes":[{"id":"n1","type":"START"}],"edges":[]}`)
	updated, err := missionRepo.UpdateGraph(ctx, created.ID, org.ID, ws.ID, graph)
	require.NoError(t, err)
	assert.JSONEq(t, string(graph), string(updated.Graph))

	renamed, err := missionRepo.Update(ctx, created.ID, org.ID, ws.ID, "New Name", "desc")
	require.NoError(t, err)
	assert.Equal(t, "New Name", renamed.Name)

	require.NoError(t, missionRepo.SoftDelete(ctx, created.ID, org.ID, ws.ID))
	_, err = missionRepo.FindByID(ctx, created.ID, org.ID, ws.ID)
	require.Error(t, err)
}

// TestMissionRepository_TenantAndWorkspaceIsolation: a mission in (orgA, wsA) is
// not reachable via another workspace (same org) nor another org.
func TestMissionRepository_TenantAndWorkspaceIsolation(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	orgRepo := repository.NewOrganizationRepository(pool)
	wsRepo := repository.NewWorkspaceRepository(pool)
	missionRepo := repository.NewMissionRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)
	userRepo := repository.NewUserRepository(pool)

	orgA := makeOrg(t, orgRepo, "org-a")
	orgB := makeOrg(t, orgRepo, "org-b")
	mbrA := makeMember(t, memberRepo, userRepo, orgA.ID, "a@example.com")

	wsA, err := wsRepo.Create(ctx, ports.Workspace{ID: uuid.New(), OrganizationID: orgA.ID, Name: "WS A", CreatedAt: time.Now()})
	require.NoError(t, err)
	wsA2, err := wsRepo.Create(ctx, ports.Workspace{ID: uuid.New(), OrganizationID: orgA.ID, Name: "WS A2", CreatedAt: time.Now()})
	require.NoError(t, err)

	m, err := missionRepo.Create(ctx, ports.Mission{
		ID: uuid.New(), OrganizationID: orgA.ID, WorkspaceID: wsA.ID, Name: "Quest",
		Status: "DRAFT", Graph: json.RawMessage(`{"nodes":[],"edges":[]}`), CreatedByID: mbrA.ID, CreatedAt: time.Now(),
	})
	require.NoError(t, err)

	// Wrong workspace (same org) → not found.
	_, err = missionRepo.FindByID(ctx, m.ID, orgA.ID, wsA2.ID)
	require.Error(t, err)
	// Wrong tenant → not found.
	_, err = missionRepo.FindByID(ctx, m.ID, orgB.ID, wsA.ID)
	require.Error(t, err)
	// Listing another workspace sees nothing.
	other, err := missionRepo.List(ctx, wsA2.ID, orgA.ID)
	require.NoError(t, err)
	assert.Empty(t, other)
}
