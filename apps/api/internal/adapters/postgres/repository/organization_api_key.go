package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

type OrganizationAPIKeyRepository struct{ db Querier }

func NewOrganizationAPIKeyRepository(db Querier) *OrganizationAPIKeyRepository {
	return &OrganizationAPIKeyRepository{db: db}
}

func (r *OrganizationAPIKeyRepository) Create(ctx context.Context, k ports.OrganizationAPIKey) (ports.OrganizationAPIKey, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO organization_api_keys (id, organization_id, name, key_hash, created_by_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, organization_id, name, key_hash, last_used_at, created_by_id, created_at, revoked_at`,
		k.ID, k.OrganizationID, k.Name, k.KeyHash, k.CreatedByID, k.CreatedAt,
	)
	created, err := scanAPIKey(row)
	if err != nil {
		return ports.OrganizationAPIKey{}, fmt.Errorf("apiKey.Create: %w", err)
	}
	return created, nil
}

func (r *OrganizationAPIKeyRepository) FindActiveByHash(ctx context.Context, keyHash string) (ports.OrganizationAPIKey, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, organization_id, name, key_hash, last_used_at, created_by_id, created_at, revoked_at
		FROM organization_api_keys
		WHERE key_hash = $1 AND revoked_at IS NULL`,
		keyHash,
	)
	k, err := scanAPIKey(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.OrganizationAPIKey{}, apierr.NotFound("api key")
		}
		return ports.OrganizationAPIKey{}, fmt.Errorf("apiKey.FindActiveByHash: %w", err)
	}
	return k, nil
}

func (r *OrganizationAPIKeyRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]ports.OrganizationAPIKey, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, organization_id, name, last_used_at, created_by_id, created_at, revoked_at
		FROM organization_api_keys
		WHERE organization_id = $1
		ORDER BY created_at DESC`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("apiKey.ListByOrg: %w", err)
	}
	defer rows.Close()

	out := make([]ports.OrganizationAPIKey, 0)
	for rows.Next() {
		var k ports.OrganizationAPIKey
		if err := rows.Scan(&k.ID, &k.OrganizationID, &k.Name, &k.LastUsedAt, &k.CreatedByID, &k.CreatedAt, &k.RevokedAt); err != nil {
			return nil, fmt.Errorf("apiKey.ListByOrg scan: %w", err)
		}
		out = append(out, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("apiKey.ListByOrg rows: %w", err)
	}
	return out, nil
}

func (r *OrganizationAPIKeyRepository) Revoke(ctx context.Context, id, orgID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE organization_api_keys SET revoked_at = NOW()
		WHERE id = $1 AND organization_id = $2 AND revoked_at IS NULL`,
		id, orgID,
	)
	if err != nil {
		return fmt.Errorf("apiKey.Revoke: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apierr.NotFound("api key")
	}
	return nil
}

func (r *OrganizationAPIKeyRepository) TouchLastUsed(ctx context.Context, id uuid.UUID, throttle time.Duration) error {
	_, err := r.db.Exec(ctx, `
		UPDATE organization_api_keys SET last_used_at = NOW()
		WHERE id = $1 AND (last_used_at IS NULL OR last_used_at < NOW() - make_interval(secs => $2))`,
		id, throttle.Seconds(),
	)
	if err != nil {
		return fmt.Errorf("apiKey.TouchLastUsed: %w", err)
	}
	return nil
}

func scanAPIKey(row pgx.Row) (ports.OrganizationAPIKey, error) {
	var k ports.OrganizationAPIKey
	err := row.Scan(
		&k.ID, &k.OrganizationID, &k.Name, &k.KeyHash,
		&k.LastUsedAt, &k.CreatedByID, &k.CreatedAt, &k.RevokedAt,
	)
	return k, err
}
