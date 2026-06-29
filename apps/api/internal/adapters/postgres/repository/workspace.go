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

type WorkspaceRepository struct{ db Querier }

func NewWorkspaceRepository(db Querier) *WorkspaceRepository {
	return &WorkspaceRepository{db: db}
}

func (r *WorkspaceRepository) Create(ctx context.Context, w ports.Workspace) (ports.Workspace, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO workspaces (id, organization_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, organization_id, name, description, created_at, deleted_at`,
		w.ID, w.OrganizationID, w.Name, w.Description, w.CreatedAt,
	)
	ws, err := scanWorkspace(row)
	if err != nil {
		return ports.Workspace{}, fmt.Errorf("workspace.Create: %w", err)
	}
	return ws, nil
}

func (r *WorkspaceRepository) FindByID(ctx context.Context, id, orgID uuid.UUID) (ports.Workspace, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, organization_id, name, description, created_at, deleted_at
		FROM workspaces WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		id, orgID,
	)
	ws, err := scanWorkspace(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Workspace{}, apierr.NotFound("workspace")
		}
		return ports.Workspace{}, fmt.Errorf("workspace.FindByID: %w", err)
	}
	return ws, nil
}

func (r *WorkspaceRepository) List(ctx context.Context, orgID uuid.UUID) ([]ports.Workspace, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, organization_id, name, description, created_at, deleted_at
		FROM workspaces WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC`, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("workspace.List: %w", err)
	}
	defer rows.Close()

	out := make([]ports.Workspace, 0)
	for rows.Next() {
		ws, err := scanWorkspace(rows)
		if err != nil {
			return nil, fmt.Errorf("workspace.List scan: %w", err)
		}
		out = append(out, ws)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("workspace.List rows: %w", err)
	}
	return out, nil
}

func (r *WorkspaceRepository) Update(ctx context.Context, id, orgID uuid.UUID, name, description string) (ports.Workspace, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE workspaces SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3 AND organization_id = $4 AND deleted_at IS NULL
		RETURNING id, organization_id, name, description, created_at, deleted_at`,
		name, description, id, orgID,
	)
	ws, err := scanWorkspace(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.Workspace{}, apierr.NotFound("workspace")
		}
		return ports.Workspace{}, fmt.Errorf("workspace.Update: %w", err)
	}
	return ws, nil
}

func (r *WorkspaceRepository) SoftDelete(ctx context.Context, id, orgID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE workspaces SET deleted_at = NOW() WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		id, orgID,
	)
	return err
}

func scanWorkspace(row pgx.Row) (ports.Workspace, error) {
	var w ports.Workspace
	var description *string
	var deletedAt *time.Time
	err := row.Scan(&w.ID, &w.OrganizationID, &w.Name, &description, &w.CreatedAt, &deletedAt)
	if description != nil {
		w.Description = *description
	}
	w.DeletedAt = deletedAt
	return w, err
}
