package authz_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/application/authz"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

var _ ports.MemberRepository = (*fakeMemberRepo)(nil)

type fakeMemberRepo struct {
	found   ports.Member
	findErr error
}

func (f *fakeMemberRepo) FindByUserAndOrg(_ context.Context, _, _ uuid.UUID) (ports.Member, error) {
	if f.findErr != nil {
		return ports.Member{}, f.findErr
	}
	return f.found, nil
}

func (f *fakeMemberRepo) Create(_ context.Context, m ports.Member) (ports.Member, error) {
	return m, nil
}
func (f *fakeMemberRepo) SoftDeleteByOrg(_ context.Context, _ uuid.UUID) error { return nil }

func forbiddenCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *apierr.AppError
	require.ErrorAs(t, err, &appErr)
	return appErr.Code
}

func TestRequireOrgRole_AdminMeetsViewerMinimum(t *testing.T) {
	repo := &fakeMemberRepo{found: ports.Member{ID: uuid.New(), Role: member.RoleAdmin}}
	m, err := authz.RequireOrgRole(context.Background(), repo, uuid.New(), uuid.New(), member.RoleViewer)
	require.NoError(t, err)
	assert.Equal(t, member.RoleAdmin, m.Role)
}

func TestRequireOrgRole_ViewerInsufficientForAdmin(t *testing.T) {
	repo := &fakeMemberRepo{found: ports.Member{ID: uuid.New(), Role: member.RoleViewer}}
	_, err := authz.RequireOrgRole(context.Background(), repo, uuid.New(), uuid.New(), member.RoleAdmin)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, forbiddenCode(t, err))
}

func TestRequireOrgRole_NotAMember(t *testing.T) {
	repo := &fakeMemberRepo{findErr: apierr.NotFound("member")}
	_, err := authz.RequireOrgRole(context.Background(), repo, uuid.New(), uuid.New(), member.RoleViewer)
	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, forbiddenCode(t, err))
}
