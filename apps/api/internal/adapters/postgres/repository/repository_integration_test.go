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
	"github.com/novarod/polina/apps/api/internal/ports"
)

const migrationPath = "../../../../db/migrations/000001_init_schema.up.sql"

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

	migration, err := os.ReadFile(migrationPath)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(migration))
	require.NoError(t, err)

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
