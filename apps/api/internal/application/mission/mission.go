// Package mission holds the application use cases for the mission module
// (quest graphs inside a workspace). Cycle 1: CRUD + structural graph editing.
package mission

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/application/authz"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	missiondomain "github.com/novarod/polina/apps/api/internal/domain/mission"
	"github.com/novarod/polina/apps/api/internal/ports"
)

// emptyGraph is the initial draft graph for a new mission.
var emptyGraph = json.RawMessage(`{"nodes":[],"edges":[]}`)

// --- Create ---

type CreateInput struct {
	UserID      uuid.UUID
	OrgID       uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Description string
}

type CreateUseCase struct {
	missions   ports.MissionRepository
	workspaces ports.WorkspaceRepository
	members    ports.MemberRepository
}

func NewCreateUseCase(missions ports.MissionRepository, workspaces ports.WorkspaceRepository, members ports.MemberRepository) *CreateUseCase {
	return &CreateUseCase{missions: missions, workspaces: workspaces, members: members}
}

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (ports.Mission, error) {
	caller, err := authz.RequireOrgRole(ctx, uc.members, in.UserID, in.OrgID, member.RoleDesigner)
	if err != nil {
		return ports.Mission{}, err
	}
	// Parent workspace must exist in this org (404 otherwise).
	if _, err := uc.workspaces.FindByID(ctx, in.WorkspaceID, in.OrgID); err != nil {
		return ports.Mission{}, err
	}
	name := strings.TrimSpace(in.Name)
	description := strings.TrimSpace(in.Description)
	if err := missiondomain.ValidateName(name); err != nil {
		return ports.Mission{}, err
	}
	if err := missiondomain.ValidateDescription(description); err != nil {
		return ports.Mission{}, err
	}
	return uc.missions.Create(ctx, ports.Mission{
		ID:             uuid.New(),
		OrganizationID: in.OrgID,
		WorkspaceID:    in.WorkspaceID,
		Name:           name,
		Description:    description,
		Status:         string(missiondomain.StatusDraft),
		Graph:          emptyGraph,
		CreatedByID:    caller.ID,
		CreatedAt:      time.Now(),
	})
}

// --- List ---

type ListUseCase struct {
	missions ports.MissionRepository
	members  ports.MemberRepository
}

func NewListUseCase(missions ports.MissionRepository, members ports.MemberRepository) *ListUseCase {
	return &ListUseCase{missions: missions, members: members}
}

func (uc *ListUseCase) Execute(ctx context.Context, userID, orgID, workspaceID uuid.UUID) ([]ports.Mission, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleViewer); err != nil {
		return nil, err
	}
	return uc.missions.List(ctx, workspaceID, orgID)
}

// --- Get ---

type GetUseCase struct {
	missions ports.MissionRepository
	members  ports.MemberRepository
}

func NewGetUseCase(missions ports.MissionRepository, members ports.MemberRepository) *GetUseCase {
	return &GetUseCase{missions: missions, members: members}
}

func (uc *GetUseCase) Execute(ctx context.Context, userID, orgID, workspaceID, missionID uuid.UUID) (ports.Mission, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleViewer); err != nil {
		return ports.Mission{}, err
	}
	return uc.missions.FindByID(ctx, missionID, orgID, workspaceID)
}

// --- UpdateGraph ---

type UpdateGraphInput struct {
	UserID      uuid.UUID
	OrgID       uuid.UUID
	WorkspaceID uuid.UUID
	MissionID   uuid.UUID
	Graph       json.RawMessage
}

type UpdateGraphUseCase struct {
	missions ports.MissionRepository
	members  ports.MemberRepository
}

func NewUpdateGraphUseCase(missions ports.MissionRepository, members ports.MemberRepository) *UpdateGraphUseCase {
	return &UpdateGraphUseCase{missions: missions, members: members}
}

func (uc *UpdateGraphUseCase) Execute(ctx context.Context, in UpdateGraphInput) (ports.Mission, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, in.UserID, in.OrgID, member.RoleDesigner); err != nil {
		return ports.Mission{}, err
	}
	if err := missiondomain.ValidateGraph(in.Graph); err != nil {
		return ports.Mission{}, err
	}
	return uc.missions.UpdateGraph(ctx, in.MissionID, in.OrgID, in.WorkspaceID, in.Graph)
}

// --- Update (name/description) ---

type UpdateInput struct {
	UserID      uuid.UUID
	OrgID       uuid.UUID
	WorkspaceID uuid.UUID
	MissionID   uuid.UUID
	Name        string
	Description string
}

type UpdateUseCase struct {
	missions ports.MissionRepository
	members  ports.MemberRepository
}

func NewUpdateUseCase(missions ports.MissionRepository, members ports.MemberRepository) *UpdateUseCase {
	return &UpdateUseCase{missions: missions, members: members}
}

func (uc *UpdateUseCase) Execute(ctx context.Context, in UpdateInput) (ports.Mission, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, in.UserID, in.OrgID, member.RoleDesigner); err != nil {
		return ports.Mission{}, err
	}
	name := strings.TrimSpace(in.Name)
	description := strings.TrimSpace(in.Description)
	if err := missiondomain.ValidateName(name); err != nil {
		return ports.Mission{}, err
	}
	if err := missiondomain.ValidateDescription(description); err != nil {
		return ports.Mission{}, err
	}
	return uc.missions.Update(ctx, in.MissionID, in.OrgID, in.WorkspaceID, name, description)
}

// --- Delete ---

type DeleteUseCase struct {
	missions ports.MissionRepository
	members  ports.MemberRepository
}

func NewDeleteUseCase(missions ports.MissionRepository, members ports.MemberRepository) *DeleteUseCase {
	return &DeleteUseCase{missions: missions, members: members}
}

func (uc *DeleteUseCase) Execute(ctx context.Context, userID, orgID, workspaceID, missionID uuid.UUID) error {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleDesigner); err != nil {
		return err
	}
	if _, err := uc.missions.FindByID(ctx, missionID, orgID, workspaceID); err != nil {
		return err
	}
	return uc.missions.SoftDelete(ctx, missionID, orgID, workspaceID)
}
