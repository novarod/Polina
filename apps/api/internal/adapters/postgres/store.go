package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/novarod/polina/apps/api/internal/adapters/postgres/repository"
	"github.com/novarod/polina/apps/api/internal/ports"
)

type Store struct{ pool *pgxpool.Pool }

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Users() ports.UserRepository     { return repository.NewUserRepository(s.pool) }
func (s *Store) Members() ports.MemberRepository { return repository.NewMemberRepository(s.pool) }
func (s *Store) Organizations() ports.OrganizationRepository {
	return repository.NewOrganizationRepository(s.pool)
}
func (s *Store) Workspaces() ports.WorkspaceRepository {
	return repository.NewWorkspaceRepository(s.pool)
}
func (s *Store) Missions() ports.MissionRepository {
	return repository.NewMissionRepository(s.pool)
}
func (s *Store) MissionVersions() ports.MissionVersionRepository {
	return repository.NewMissionVersionRepository(s.pool)
}
func (s *Store) OrganizationAPIKeys() ports.OrganizationAPIKeyRepository {
	return repository.NewOrganizationAPIKeyRepository(s.pool)
}

func (s *Store) WithinTx(ctx context.Context, fn func(ports.Repositories) error) (err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	if err := fn(&txRepos{tx: tx}); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

type txRepos struct{ tx pgx.Tx }

func (t *txRepos) Users() ports.UserRepository     { return repository.NewUserRepository(t.tx) }
func (t *txRepos) Members() ports.MemberRepository { return repository.NewMemberRepository(t.tx) }
func (t *txRepos) Organizations() ports.OrganizationRepository {
	return repository.NewOrganizationRepository(t.tx)
}
func (t *txRepos) Workspaces() ports.WorkspaceRepository {
	return repository.NewWorkspaceRepository(t.tx)
}
func (t *txRepos) Missions() ports.MissionRepository {
	return repository.NewMissionRepository(t.tx)
}
func (t *txRepos) MissionVersions() ports.MissionVersionRepository {
	return repository.NewMissionVersionRepository(t.tx)
}
func (t *txRepos) OrganizationAPIKeys() ports.OrganizationAPIKeyRepository {
	return repository.NewOrganizationAPIKeyRepository(t.tx)
}
