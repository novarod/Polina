package workspace

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/application/authz"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	wsdomain "github.com/novarod/polina/apps/api/internal/domain/workspace"
	"github.com/novarod/polina/apps/api/internal/ports"
)

// --- Create ---

type CreateInput struct {
	UserID      uuid.UUID
	OrgID       uuid.UUID
	Name        string
	Description string
}

type CreateUseCase struct {
	workspaces ports.WorkspaceRepository
	members    ports.MemberRepository
}

func NewCreateUseCase(workspaces ports.WorkspaceRepository, members ports.MemberRepository) *CreateUseCase {
	return &CreateUseCase{workspaces: workspaces, members: members}
}

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (ports.Workspace, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, in.UserID, in.OrgID, member.RoleDesigner); err != nil {
		return ports.Workspace{}, err
	}
	name := strings.TrimSpace(in.Name)
	description := strings.TrimSpace(in.Description)
	if err := wsdomain.ValidateName(name); err != nil {
		return ports.Workspace{}, err
	}
	if err := wsdomain.ValidateDescription(description); err != nil {
		return ports.Workspace{}, err
	}
	return uc.workspaces.Create(ctx, ports.Workspace{
		ID:             uuid.New(),
		OrganizationID: in.OrgID,
		Name:           name,
		Description:    description,
		CreatedAt:      time.Now(),
	})
}

// --- List ---

type ListUseCase struct {
	workspaces ports.WorkspaceRepository
	members    ports.MemberRepository
}

func NewListUseCase(workspaces ports.WorkspaceRepository, members ports.MemberRepository) *ListUseCase {
	return &ListUseCase{workspaces: workspaces, members: members}
}

func (uc *ListUseCase) Execute(ctx context.Context, userID, orgID uuid.UUID) ([]ports.Workspace, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleViewer); err != nil {
		return nil, err
	}
	return uc.workspaces.List(ctx, orgID)
}

// --- Get ---

type GetUseCase struct {
	workspaces ports.WorkspaceRepository
	members    ports.MemberRepository
}

func NewGetUseCase(workspaces ports.WorkspaceRepository, members ports.MemberRepository) *GetUseCase {
	return &GetUseCase{workspaces: workspaces, members: members}
}

func (uc *GetUseCase) Execute(ctx context.Context, userID, orgID, workspaceID uuid.UUID) (ports.Workspace, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleViewer); err != nil {
		return ports.Workspace{}, err
	}
	return uc.workspaces.FindByID(ctx, workspaceID, orgID)
}

// --- Update ---

type UpdateInput struct {
	UserID      uuid.UUID
	OrgID       uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Description string
}

type UpdateUseCase struct {
	workspaces ports.WorkspaceRepository
	members    ports.MemberRepository
}

func NewUpdateUseCase(workspaces ports.WorkspaceRepository, members ports.MemberRepository) *UpdateUseCase {
	return &UpdateUseCase{workspaces: workspaces, members: members}
}

func (uc *UpdateUseCase) Execute(ctx context.Context, in UpdateInput) (ports.Workspace, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, in.UserID, in.OrgID, member.RoleDesigner); err != nil {
		return ports.Workspace{}, err
	}
	name := strings.TrimSpace(in.Name)
	description := strings.TrimSpace(in.Description)
	if err := wsdomain.ValidateName(name); err != nil {
		return ports.Workspace{}, err
	}
	if err := wsdomain.ValidateDescription(description); err != nil {
		return ports.Workspace{}, err
	}
	return uc.workspaces.Update(ctx, in.WorkspaceID, in.OrgID, name, description)
}

// --- Delete ---

type DeleteUseCase struct {
	workspaces ports.WorkspaceRepository
	members    ports.MemberRepository
}

func NewDeleteUseCase(workspaces ports.WorkspaceRepository, members ports.MemberRepository) *DeleteUseCase {
	return &DeleteUseCase{workspaces: workspaces, members: members}
}

func (uc *DeleteUseCase) Execute(ctx context.Context, userID, orgID, workspaceID uuid.UUID) error {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleDesigner); err != nil {
		return err
	}
	if _, err := uc.workspaces.FindByID(ctx, workspaceID, orgID); err != nil {
		return err
	}
	return uc.workspaces.SoftDelete(ctx, workspaceID, orgID)
}
