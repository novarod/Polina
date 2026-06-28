package token

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/domain/member"
)

type Claims struct {
	UserID   uuid.UUID   `json:"user_id"`
	MemberID uuid.UUID   `json:"member_id"`
	OrgID    uuid.UUID   `json:"org_id"`
	Role     member.Role `json:"role"`
	jwt.RegisteredClaims
}
