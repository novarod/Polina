package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/novarod/polina/apps/api/internal/domain/member"
)

// --- User ---

type User struct {
	ID        uuid.UUID
	Email     string
	Name      string
	Password  string
	CreatedAt time.Time
	DeletedAt *time.Time
}

type UserRepository interface {
	Create(ctx context.Context, u User) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
}

// --- Member ---

type Member struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Role           member.Role
	CreatedAt      time.Time
	DeletedAt      *time.Time
}

type MemberRepository interface {
	Create(ctx context.Context, m Member) (Member, error)
	FindByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (Member, error)
	FindByID(ctx context.Context, id uuid.UUID) (Member, error)
	ListByOrg(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]Member, int, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role member.Role) (Member, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	SoftDeleteByOrg(ctx context.Context, orgID uuid.UUID) error
}

// --- Organization ---

type Organization struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedAt time.Time
	DeletedAt *time.Time
}

type OrganizationWithRole struct {
	Organization
	Role member.Role
}

type OrganizationRepository interface {
	Create(ctx context.Context, o Organization) (Organization, error)
	FindByID(ctx context.Context, id uuid.UUID) (Organization, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]OrganizationWithRole, error)
	Update(ctx context.Context, id uuid.UUID, name string) (Organization, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// --- Transactions ---

type Repositories interface {
	Users() UserRepository
	Members() MemberRepository
	Organizations() OrganizationRepository
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(Repositories) error) error
}
