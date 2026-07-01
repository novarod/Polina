package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

type MissionVersionRepository struct{ db Querier }

func NewMissionVersionRepository(db Querier) *MissionVersionRepository {
	return &MissionVersionRepository{db: db}
}

func (r *MissionVersionRepository) Create(ctx context.Context, v ports.MissionVersion) (ports.MissionVersion, error) {
	row := r.db.QueryRow(ctx, `
		WITH next AS (
			SELECT COALESCE(MAX(version_number), 0) + 1 AS n
			FROM mission_versions WHERE mission_id = $1
		)
		INSERT INTO mission_versions (id, mission_id, organization_id, version_number, hash, graph, mission_data, published_by_id, created_at)
		SELECT $2, $1, $3, next.n, $4, $5, jsonb_set($6::jsonb, '{version}', to_jsonb(next.n)), $7, NOW()
		FROM next
		RETURNING id, mission_id, organization_id, version_number, hash, graph, mission_data, published_by_id, created_at`,
		v.MissionID, v.ID, v.OrganizationID, v.Hash, v.Graph, v.MissionData, v.PublishedByID,
	)
	created, err := scanMissionVersion(row)
	if err != nil {
		return ports.MissionVersion{}, fmt.Errorf("missionVersion.Create: %w", err)
	}
	return created, nil
}

func (r *MissionVersionRepository) FindByHash(ctx context.Context, missionID, orgID uuid.UUID, hash string) (ports.MissionVersion, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, mission_id, organization_id, version_number, hash, graph, mission_data, published_by_id, created_at
		FROM mission_versions
		WHERE mission_id = $1 AND organization_id = $2 AND hash = $3`,
		missionID, orgID, hash,
	)
	v, err := scanMissionVersion(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.MissionVersion{}, apierr.NotFound("mission version")
		}
		return ports.MissionVersion{}, fmt.Errorf("missionVersion.FindByHash: %w", err)
	}
	return v, nil
}

func (r *MissionVersionRepository) List(ctx context.Context, missionID, orgID uuid.UUID) ([]ports.MissionVersion, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, mission_id, organization_id, version_number, hash, published_by_id, created_at
		FROM mission_versions
		WHERE mission_id = $1 AND organization_id = $2
		ORDER BY version_number DESC`,
		missionID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("missionVersion.List: %w", err)
	}
	defer rows.Close()

	out := make([]ports.MissionVersion, 0)
	for rows.Next() {
		var v ports.MissionVersion
		if err := rows.Scan(&v.ID, &v.MissionID, &v.OrganizationID, &v.VersionNumber, &v.Hash, &v.PublishedByID, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("missionVersion.List scan: %w", err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("missionVersion.List rows: %w", err)
	}
	return out, nil
}

func scanMissionVersion(row pgx.Row) (ports.MissionVersion, error) {
	var v ports.MissionVersion
	err := row.Scan(
		&v.ID, &v.MissionID, &v.OrganizationID, &v.VersionNumber, &v.Hash,
		&v.Graph, &v.MissionData, &v.PublishedByID, &v.CreatedAt,
	)
	return v, err
}
