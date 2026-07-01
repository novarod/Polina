package engine

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/ports"
)

// --- Active hash (poll) ---

type GetActiveHashUseCase struct {
	missions ports.MissionRepository
}

func NewGetActiveHashUseCase(missions ports.MissionRepository) *GetActiveHashUseCase {
	return &GetActiveHashUseCase{missions: missions}
}

func (uc *GetActiveHashUseCase) Execute(ctx context.Context, orgID, missionID uuid.UUID) (string, error) {
	return uc.missions.FindActiveHash(ctx, missionID, orgID)
}

// --- Active contract ---

type GetActiveContractUseCase struct {
	versions ports.MissionVersionRepository
}

func NewGetActiveContractUseCase(versions ports.MissionVersionRepository) *GetActiveContractUseCase {
	return &GetActiveContractUseCase{versions: versions}
}

func (uc *GetActiveContractUseCase) Execute(ctx context.Context, orgID, missionID uuid.UUID) (json.RawMessage, error) {
	v, err := uc.versions.FindActive(ctx, missionID, orgID)
	if err != nil {
		return nil, err
	}
	return v.MissionData, nil
}
