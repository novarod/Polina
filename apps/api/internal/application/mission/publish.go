package mission

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/application/authz"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	missiondomain "github.com/novarod/polina/apps/api/internal/domain/mission"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

// --- Publish ---

type PublishInput struct {
	UserID      uuid.UUID
	OrgID       uuid.UUID
	WorkspaceID uuid.UUID
	MissionID   uuid.UUID
}

type PublishResult struct {
	Version ports.MissionVersion
	Mission ports.Mission
}

type PublishUseCase struct{ tx ports.TxManager }

func NewPublishUseCase(tx ports.TxManager) *PublishUseCase { return &PublishUseCase{tx: tx} }

func (uc *PublishUseCase) Execute(ctx context.Context, in PublishInput) (PublishResult, error) {
	var result PublishResult
	err := uc.tx.WithinTx(ctx, func(r ports.Repositories) error {
		caller, err := authz.RequireOrgRole(ctx, r.Members(), in.UserID, in.OrgID, member.RoleDesigner)
		if err != nil {
			return err
		}

		missions := r.Missions()
		versions := r.MissionVersions()

		m, err := missions.FindByIDForUpdate(ctx, in.MissionID, in.OrgID, in.WorkspaceID)
		if err != nil {
			return err
		}

		contract, err := missiondomain.Compile(m.ID.String(), m.Graph)
		if err != nil {
			return err
		}
		h, err := contract.ContentHash()
		if err != nil {
			return err
		}

		version, err := versions.FindByHash(ctx, m.ID, in.OrgID, h)
		switch {
		case err == nil:
		case apierr.IsNotFound(err):
			contract.Hash = h
			data, mErr := json.Marshal(contract)
			if mErr != nil {
				return mErr
			}
			version, err = versions.Create(ctx, ports.MissionVersion{
				ID:             uuid.New(),
				MissionID:      m.ID,
				OrganizationID: in.OrgID,
				Hash:           h,
				Graph:          m.Graph,
				MissionData:    data,
				PublishedByID:  caller.ID,
			})
			if err != nil {
				return err
			}
		default:
			return err
		}

		updated, err := missions.SetActiveVersion(ctx, m.ID, in.OrgID, in.WorkspaceID, h, string(missiondomain.StatusApproved))
		if err != nil {
			return err
		}
		result = PublishResult{Version: version, Mission: updated}
		return nil
	})
	if err != nil {
		return PublishResult{}, err
	}
	return result, nil
}

// --- List versions ---

type ListVersionsUseCase struct {
	missions ports.MissionRepository
	versions ports.MissionVersionRepository
	members  ports.MemberRepository
}

func NewListVersionsUseCase(missions ports.MissionRepository, versions ports.MissionVersionRepository, members ports.MemberRepository) *ListVersionsUseCase {
	return &ListVersionsUseCase{missions: missions, versions: versions, members: members}
}

func (uc *ListVersionsUseCase) Execute(ctx context.Context, userID, orgID, workspaceID, missionID uuid.UUID) ([]ports.MissionVersion, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleViewer); err != nil {
		return nil, err
	}
	if _, err := uc.missions.FindByID(ctx, missionID, orgID, workspaceID); err != nil {
		return nil, err
	}
	return uc.versions.List(ctx, missionID, orgID)
}

// --- Get version (by hash) ---

type GetVersionUseCase struct {
	missions ports.MissionRepository
	versions ports.MissionVersionRepository
	members  ports.MemberRepository
}

func NewGetVersionUseCase(missions ports.MissionRepository, versions ports.MissionVersionRepository, members ports.MemberRepository) *GetVersionUseCase {
	return &GetVersionUseCase{missions: missions, versions: versions, members: members}
}

func (uc *GetVersionUseCase) Execute(ctx context.Context, userID, orgID, workspaceID, missionID uuid.UUID, hash string) (ports.MissionVersion, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleViewer); err != nil {
		return ports.MissionVersion{}, err
	}
	if _, err := uc.missions.FindByID(ctx, missionID, orgID, workspaceID); err != nil {
		return ports.MissionVersion{}, err
	}
	return uc.versions.FindByHash(ctx, missionID, orgID, hash)
}
