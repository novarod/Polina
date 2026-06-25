package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

const uniqueViolationCode = "23505"

type OrganizationRepository struct{ db Querier }

func NewOrganizationRepository(db Querier) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, o ports.Organization) (ports.Organization, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO organizations (id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id, name, slug, created_at, deleted_at`,
		o.ID, o.Name, o.Slug, o.CreatedAt,
	)
	org, err := scanOrganization(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return ports.Organization{}, apierr.Validation("slug", "slug already in use")
		}
		return ports.Organization{}, fmt.Errorf("organization.Create: %w", err)
	}
	return org, nil
}

func (r *OrganizationRepository) FindByID(ctx context.Context, id uuid.UUID) (ports.Organization, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, slug, created_at, deleted_at
		FROM organizations WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	o, err := scanOrganization(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Organization{}, apierr.NotFound("organization")
		}
		return ports.Organization{}, fmt.Errorf("organization.FindByID: %w", err)
	}
	return o, nil
}

func (r *OrganizationRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]ports.OrganizationWithRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT o.id, o.name, o.slug, o.created_at, o.deleted_at, m.role
		FROM organizations o
		JOIN members m ON m.organization_id = o.id
		WHERE m.user_id = $1 AND m.deleted_at IS NULL AND o.deleted_at IS NULL
		ORDER BY o.created_at ASC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("organization.ListByUserID: %w", err)
	}
	defer rows.Close()

	out := make([]ports.OrganizationWithRole, 0)
	for rows.Next() {
		var o ports.Organization
		var deletedAt *time.Time
		var roleStr string
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &deletedAt, &roleStr); err != nil {
			return nil, fmt.Errorf("organization.ListByUserID scan: %w", err)
		}
		o.DeletedAt = deletedAt
		out = append(out, ports.OrganizationWithRole{Organization: o, Role: member.Role(roleStr)})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("organization.ListByUserID rows: %w", err)
	}
	return out, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, id uuid.UUID, name string) (ports.Organization, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE organizations SET name = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id, name, slug, created_at, deleted_at`,
		name, id,
	)
	o, err := scanOrganization(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Organization{}, apierr.NotFound("organization")
		}
		return ports.Organization{}, fmt.Errorf("organization.Update: %w", err)
	}
	return o, nil
}

func (r *OrganizationRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE organizations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

func scanOrganization(row pgx.Row) (ports.Organization, error) {
	var o ports.Organization
	var deletedAt *time.Time
	err := row.Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &deletedAt)
	o.DeletedAt = deletedAt
	return o, err
}
