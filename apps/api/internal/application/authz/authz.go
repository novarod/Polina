package authz

import (
	"context"

	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

func RequireOrgRole(ctx context.Context, members ports.MemberRepository, userID, orgID uuid.UUID, minimum member.Role) (ports.Member, error) {
	m, err := members.FindByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return ports.Member{}, apierr.Forbidden("not a member of this organization")
	}
	if !m.Role.AtLeast(minimum) {
		return ports.Member{}, apierr.Forbidden("insufficient role")
	}
	return m, nil
}
