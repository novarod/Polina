package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

type UserRepository struct{ db Querier }

func NewUserRepository(db Querier) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, u ports.User) (ports.User, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO users (id, email, name, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id, email, name, password, created_at, deleted_at`,
		u.ID, u.Email, u.Name, u.Password, u.CreatedAt,
	)
	return scanUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (ports.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, name, password, created_at, deleted_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`,
		strings.ToLower(email),
	)
	u, err := scanUser(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.User{}, apierr.NotFound("user")
		}
		return ports.User{}, fmt.Errorf("user.FindByEmail: %w", err)
	}
	return u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (ports.User, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, email, name, password, created_at, deleted_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	u, err := scanUser(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ports.User{}, apierr.NotFound("user")
		}
		return ports.User{}, fmt.Errorf("user.FindByID: %w", err)
	}
	return u, nil
}

func scanUser(row pgx.Row) (ports.User, error) {
	var u ports.User
	var deletedAt *time.Time
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Password, &u.CreatedAt, &deletedAt)
	u.DeletedAt = deletedAt
	return u, err
}
