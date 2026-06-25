//go:build integration

package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/postgres/repository"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
)

// migrationPath is relative to this package directory.
const migrationPath = "../../../../db/migrations/000001_init_schema.up.sql"

// setupDB connects to the test database, resets it to a pristine schema and
// applies the migration. Set TEST_DATABASE_URL to override the default DSN
// (which matches docker-compose.yml).
func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://polina:polina@localhost:5432/polina?sslmode=disable"
	}
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(ctx))

	// Pristine schema per run so tests are deterministic and isolated.
	_, err = pool.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	require.NoError(t, err)

	migration, err := os.ReadFile(migrationPath)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(migration))
	require.NoError(t, err)

	t.Cleanup(pool.Close)
	return pool
}

func seedOrg(t *testing.T, pool *pgxpool.Pool, slug string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO organizations (name, slug) VALUES ($1, $2) RETURNING id`, slug, slug,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func newTestUser(email string) ports.User {
	return ports.User{
		ID:        uuid.New(),
		Email:     email,
		Name:      "Test " + email,
		Password:  "hashed-placeholder",
		CreatedAt: time.Now(),
	}
}

func TestUserRepository_CreateAndFind(t *testing.T) {
	pool := setupDB(t)
	repo := repository.NewUserRepository(pool)
	ctx := context.Background()

	u := newTestUser("create@example.com")
	created, err := repo.Create(ctx, u)
	require.NoError(t, err)
	assert.Equal(t, u.ID, created.ID)

	byEmail, err := repo.FindByEmail(ctx, "create@example.com")
	require.NoError(t, err)
	assert.Equal(t, u.ID, byEmail.ID)

	byID, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "create@example.com", byID.Email)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	pool := setupDB(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.FindByEmail(context.Background(), "missing@example.com")
	require.Error(t, err)
}

// TestMemberRepository_ListByOrg_ReturnsCorrectTotal exercises the COUNT path
// in ListByOrg, guarding against the regression where the count error was
// silently swallowed (member.go:85).
func TestMemberRepository_ListByOrg_ReturnsCorrectTotal(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)

	orgID := seedOrg(t, pool, "acme")

	emails := []string{"m1@example.com", "m2@example.com", "m3@example.com"}
	for _, email := range emails {
		u, err := userRepo.Create(ctx, newTestUser(email))
		require.NoError(t, err)
		_, err = memberRepo.Create(ctx, ports.Member{
			ID:             uuid.New(),
			UserID:         u.ID,
			OrganizationID: orgID,
			Role:           member.RoleViewer,
			CreatedAt:      time.Now(),
		})
		require.NoError(t, err)
	}

	members, total, err := memberRepo.ListByOrg(ctx, orgID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, members, len(emails))
	assert.Equal(t, len(emails), total)
}

func TestMemberRepository_SoftDeleteExcludedFromList(t *testing.T) {
	pool := setupDB(t)
	ctx := context.Background()
	userRepo := repository.NewUserRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)

	orgID := seedOrg(t, pool, "beta")
	u, err := userRepo.Create(ctx, newTestUser("solo@example.com"))
	require.NoError(t, err)
	m, err := memberRepo.Create(ctx, ports.Member{
		ID:             uuid.New(),
		UserID:         u.ID,
		OrganizationID: orgID,
		Role:           member.RoleDesigner,
		CreatedAt:      time.Now(),
	})
	require.NoError(t, err)

	require.NoError(t, memberRepo.SoftDelete(ctx, m.ID))

	members, total, err := memberRepo.ListByOrg(ctx, orgID, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, members)
	assert.Equal(t, 0, total)
}
