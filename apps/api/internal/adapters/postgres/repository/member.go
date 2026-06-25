package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

type MemberRepository struct{ pool *pgxpool.Pool }

func NewMemberRepository(pool *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{pool: pool}
}

func (r *MemberRepository) Create(ctx context.Context, m ports.Member) (ports.Member, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO members (id, user_id, organization_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, user_id, organization_id, role, created_at, deleted_at`,
		m.ID, m.UserID, m.OrganizationID, string(m.Role), m.CreatedAt,
	)
	return scanMember(row)
}

func (r *MemberRepository) FindByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (ports.Member, error) {
	row := r.pool.QueryRow(ctx, `
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

func (r *MemberRepository) FindByID(ctx context.Context, id uuid.UUID) (ports.Member, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, organization_id, role, created_at, deleted_at
		FROM members WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	m, err := scanMember(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Member{}, apierr.NotFound("member")
		}
		return ports.Member{}, fmt.Errorf("member.FindByID: %w", err)
	}
	return m, nil
}

func (r *MemberRepository) ListByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]ports.Member, int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, organization_id, role, created_at, deleted_at
		FROM members WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		orgID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("member.ListByOrg: %w", err)
	}
	defer rows.Close()

	var members []ports.Member
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, 0, err
		}
		members = append(members, m)
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM members WHERE organization_id = $1 AND deleted_at IS NULL`, orgID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("member.ListByOrg count: %w", err)
	}

	return members, total, nil
}

func (r *MemberRepository) UpdateRole(ctx context.Context, id uuid.UUID, role member.Role) (ports.Member, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE members SET role = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id, user_id, organization_id, role, created_at, deleted_at`,
		string(role), id,
	)
	return scanMember(row)
}

func (r *MemberRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE members SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

func (r *MemberRepository) SoftDeleteByOrg(ctx context.Context, orgID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE members SET deleted_at = NOW() WHERE organization_id = $1 AND deleted_at IS NULL`, orgID)
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
