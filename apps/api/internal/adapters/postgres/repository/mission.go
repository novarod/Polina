package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

type MissionRepository struct{ db Querier }

func NewMissionRepository(db Querier) *MissionRepository {
	return &MissionRepository{db: db}
}

func (r *MissionRepository) Create(ctx context.Context, m ports.Mission) (ports.Mission, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO missions (id, organization_id, workspace_id, name, description, status, graph, created_by_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id, organization_id, workspace_id, name, description, status, active_hash, graph, created_by_id, created_at, deleted_at`,
		m.ID, m.OrganizationID, m.WorkspaceID, m.Name, m.Description, m.Status, m.Graph, m.CreatedByID, m.CreatedAt,
	)
	mission, err := scanMission(row)
	if err != nil {
		return ports.Mission{}, fmt.Errorf("mission.Create: %w", err)
	}
	return mission, nil
}

func (r *MissionRepository) FindByID(ctx context.Context, id, orgID, workspaceID uuid.UUID) (ports.Mission, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, organization_id, workspace_id, name, description, status, active_hash, graph, created_by_id, created_at, deleted_at
		FROM missions
		WHERE id = $1 AND organization_id = $2 AND workspace_id = $3 AND deleted_at IS NULL`,
		id, orgID, workspaceID,
	)
	m, err := scanMission(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Mission{}, apierr.NotFound("mission")
		}
		return ports.Mission{}, fmt.Errorf("mission.FindByID: %w", err)
	}
	return m, nil
}

func (r *MissionRepository) FindByIDForUpdate(ctx context.Context, id, orgID, workspaceID uuid.UUID) (ports.Mission, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, organization_id, workspace_id, name, description, status, active_hash, graph, created_by_id, created_at, deleted_at
		FROM missions
		WHERE id = $1 AND organization_id = $2 AND workspace_id = $3 AND deleted_at IS NULL
		FOR UPDATE`,
		id, orgID, workspaceID,
	)
	m, err := scanMission(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Mission{}, apierr.NotFound("mission")
		}
		return ports.Mission{}, fmt.Errorf("mission.FindByIDForUpdate: %w", err)
	}
	return m, nil
}

func (r *MissionRepository) List(ctx context.Context, workspaceID, orgID uuid.UUID) ([]ports.Mission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, organization_id, workspace_id, name, description, status, active_hash, graph, created_by_id, created_at, deleted_at
		FROM missions
		WHERE workspace_id = $1 AND organization_id = $2 AND deleted_at IS NULL
		ORDER BY created_at ASC`,
		workspaceID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("mission.List: %w", err)
	}
	defer rows.Close()

	out := make([]ports.Mission, 0)
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			return nil, fmt.Errorf("mission.List scan: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("mission.List rows: %w", err)
	}
	return out, nil
}

func (r *MissionRepository) UpdateGraph(ctx context.Context, id, orgID, workspaceID uuid.UUID, graph json.RawMessage) (ports.Mission, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE missions SET graph = $1, updated_at = NOW()
		WHERE id = $2 AND organization_id = $3 AND workspace_id = $4 AND deleted_at IS NULL
		RETURNING id, organization_id, workspace_id, name, description, status, active_hash, graph, created_by_id, created_at, deleted_at`,
		graph, id, orgID, workspaceID,
	)
	m, err := scanMission(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Mission{}, apierr.NotFound("mission")
		}
		return ports.Mission{}, fmt.Errorf("mission.UpdateGraph: %w", err)
	}
	return m, nil
}

func (r *MissionRepository) Update(ctx context.Context, id, orgID, workspaceID uuid.UUID, name, description string) (ports.Mission, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE missions SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3 AND organization_id = $4 AND workspace_id = $5 AND deleted_at IS NULL
		RETURNING id, organization_id, workspace_id, name, description, status, active_hash, graph, created_by_id, created_at, deleted_at`,
		name, description, id, orgID, workspaceID,
	)
	m, err := scanMission(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Mission{}, apierr.NotFound("mission")
		}
		return ports.Mission{}, fmt.Errorf("mission.Update: %w", err)
	}
	return m, nil
}

func (r *MissionRepository) SetActiveVersion(ctx context.Context, id, orgID, workspaceID uuid.UUID, hash, status string) (ports.Mission, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE missions SET active_hash = $1, status = $2, updated_at = NOW()
		WHERE id = $3 AND organization_id = $4 AND workspace_id = $5 AND deleted_at IS NULL
		RETURNING id, organization_id, workspace_id, name, description, status, active_hash, graph, created_by_id, created_at, deleted_at`,
		hash, status, id, orgID, workspaceID,
	)
	m, err := scanMission(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Mission{}, apierr.NotFound("mission")
		}
		return ports.Mission{}, fmt.Errorf("mission.SetActiveVersion: %w", err)
	}
	return m, nil
}

func (r *MissionRepository) SoftDelete(ctx context.Context, id, orgID, workspaceID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE missions SET deleted_at = NOW() WHERE id = $1 AND organization_id = $2 AND workspace_id = $3 AND deleted_at IS NULL`,
		id, orgID, workspaceID,
	)
	return err
}

func scanMission(row pgx.Row) (ports.Mission, error) {
	var m ports.Mission
	var description *string
	var deletedAt *time.Time
	err := row.Scan(
		&m.ID, &m.OrganizationID, &m.WorkspaceID, &m.Name, &description,
		&m.Status, &m.ActiveHash, &m.Graph, &m.CreatedByID, &m.CreatedAt, &deletedAt,
	)
	if description != nil {
		m.Description = *description
	}
	m.DeletedAt = deletedAt
	return m, err
}
