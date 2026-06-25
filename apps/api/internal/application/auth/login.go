package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
	"golang.org/x/crypto/bcrypt"
)

type LoginInput struct {
	Email          string `json:"email"    validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	OrganizationID string `json:"organization_id" validate:"omitempty,uuid"`
}

type LoginOutput struct {
	Token  string    `json:"token"`
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}

type LoginUseCase struct {
	users     ports.UserRepository
	members   ports.MemberRepository
	jwtSecret string
	expiryH   int
}

func NewLoginUseCase(users ports.UserRepository, members ports.MemberRepository, jwtSecret string, expiryH int) *LoginUseCase {
	return &LoginUseCase{users: users, members: members, jwtSecret: jwtSecret, expiryH: expiryH}
}

func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (LoginOutput, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))

	user, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		return LoginOutput{}, apierr.ErrBadLogin
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return LoginOutput{}, apierr.ErrBadLogin
	}

	claims := &middleware.Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(uc.expiryH) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	if in.OrganizationID != "" {
		orgID, _ := uuid.Parse(in.OrganizationID)
		m, err := uc.members.FindByUserAndOrg(ctx, user.ID, orgID)
		if err != nil {
			return LoginOutput{}, apierr.Forbidden("not a member of this organization")
		}
		claims.MemberID = m.ID
		claims.OrgID = orgID
		claims.Role = m.Role
	} else {
		claims.Role = member.Role("")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(uc.jwtSecret))
	if err != nil {
		return LoginOutput{}, fmt.Errorf("login: sign token: %w", err)
	}

	return LoginOutput{Token: signed, UserID: user.ID, Name: user.Name}, nil
}
