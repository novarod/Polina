package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

type MemberRepository struct{ db Querier }

func NewMemberRepository(db Querier) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) Create(ctx context.Context, m ports.Member) (ports.Member, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO members (id, user_id, organization_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, user_id, organization_id, role, created_at, deleted_at`,
		m.ID, m.UserID, m.OrganizationID, string(m.Role), m.CreatedAt,
	)
	return scanMember(row)
}

func (r *MemberRepository) FindByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (ports.Member, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, organization_id, role, created_at, deleted_at
		FROM members WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		userID, orgID,
	)
	m, err := scanMember(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Member{}, apierr.NotFound("member")
		}
		return ports.Member{}, fmt.Errorf("member.FindByUserAndOrg: %w", err)
	}
	return m, nil
}

func (r *MemberRepository) SoftDeleteByOrg(ctx context.Context, orgID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE members SET deleted_at = NOW() WHERE organization_id = $1 AND deleted_at IS NULL`, orgID)
	return err
}

func scanMember(row pgx.Row) (ports.Member, error) {
	var m ports.Member
	var roleStr string
	var deletedAt *time.Time
	err := row.Scan(&m.ID, &m.UserID, &m.OrganizationID, &roleStr, &m.CreatedAt, &deletedAt)
	m.Role = member.Role(roleStr)
	m.DeletedAt = deletedAt
	return m, err
}
