package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/ports"
)

type MeOutput struct {
	UserID uuid.UUID `json:"user_id" swaggertype:"string" format:"uuid"`
	Name   string    `json:"name"`
}

type MeUseCase struct {
	users ports.UserRepository
}

func NewMeUseCase(users ports.UserRepository) *MeUseCase {
	return &MeUseCase{users: users}
}

func (uc *MeUseCase) Execute(ctx context.Context, userID uuid.UUID) (MeOutput, error) {
	user, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return MeOutput{}, err
	}
	return MeOutput{UserID: user.ID, Name: user.Name}, nil
}
