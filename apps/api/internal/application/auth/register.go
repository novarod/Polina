package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterOutput struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
}

type RegisterUseCase struct {
	users        ports.UserRepository
	bcryptRounds int
}

func NewRegisterUseCase(users ports.UserRepository, bcryptRounds int) *RegisterUseCase {
	return &RegisterUseCase{users: users, bcryptRounds: bcryptRounds}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, in RegisterInput) (RegisterOutput, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	if _, err := uc.users.FindByEmail(ctx, email); err == nil {
		return RegisterOutput{}, apierr.Validation("email", "email already in use")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), uc.bcryptRounds)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("register: hash password: %w", err)
	}

	user, err := uc.users.Create(ctx, ports.User{
		ID:        uuid.New(),
		Email:     email,
		Name:      strings.TrimSpace(in.Name),
		Password:  string(hashed),
		CreatedAt: time.Now(),
	})
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("register: create user: %w", err)
	}

	return RegisterOutput{UserID: user.ID, Email: user.Email, Name: user.Name}, nil
}
