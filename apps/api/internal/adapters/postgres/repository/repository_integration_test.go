//go:build integration

package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/postgres/repository"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

const migrationsDir = "../../../../db/migrations"

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

	_, err = pool.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	require.NoError(t, err)

	migrations, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	require.NoError(t, err)
	require.NotEmpty(t, migrations)
	sort.Strings(migrations)
	for _, path := range migrations {
		migration, err := os.ReadFile(path)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, string(migration))
		require.NoError(t, err, "applying %s", path)
	}

	t.Cleanup(pool.Close)
	return pool
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

func TestUserRepository_BumpTokenValidAfter(t *testing.T) {
	pool := setupDB(t)
	repo := repository.NewUserRepository(pool)
	ctx := context.Background()

	u := newTestUser("bump@example.com")
	_, err := repo.Create(ctx, u)
	require.NoError(t, err)

	before, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Nil(t, before.TokenValidAfter)

	require.NoError(t, repo.BumpTokenValidAfter(ctx, u.ID))

	after, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	require.NotNil(t, after.TokenValidAfter)
	assert.WithinDuration(t, time.Now(), *after.TokenValidAfter, time.Minute)
}

func TestUserRepository_BumpTokenValidAfter_NotFound(t *testing.T) {
	pool := setupDB(t)
	repo := repository.NewUserRepository(pool)

	err := repo.BumpTokenValidAfter(context.Background(), uuid.New())
	require.Error(t, err)
	assert.True(t, apierr.IsNotFound(err))
}
