package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/ports"
)

type LogoutAllUseCase struct {
	users ports.UserRepository
}

func NewLogoutAllUseCase(users ports.UserRepository) *LogoutAllUseCase {
	return &LogoutAllUseCase{users: users}
}

func (uc *LogoutAllUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	return uc.users.BumpTokenValidAfter(ctx, userID)
}
