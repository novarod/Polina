// Package organization holds the application use cases for the organization
// (tenant) module.
package organization

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/application/authz"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	orgdomain "github.com/novarod/polina/apps/api/internal/domain/organization"
	"github.com/novarod/polina/apps/api/internal/ports"
)

// --- Create ---

type CreateInput struct {
	UserID uuid.UUID
	Name   string
	Slug   string
}

// CreateUseCase creates an organization and, in the same transaction, makes the
// creator an ADMIN member.
type CreateUseCase struct{ tx ports.TxManager }

func NewCreateUseCase(tx ports.TxManager) *CreateUseCase { return &CreateUseCase{tx: tx} }

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (ports.Organization, error) {
	name := strings.TrimSpace(in.Name)
	slug := strings.TrimSpace(in.Slug)
	if err := orgdomain.ValidateName(name); err != nil {
		return ports.Organization{}, err
	}
	if err := orgdomain.ValidateSlug(slug); err != nil {
		return ports.Organization{}, err
	}

	now := time.Now()
	org := ports.Organization{ID: uuid.New(), Name: name, Slug: slug, CreatedAt: now}

	err := uc.tx.WithinTx(ctx, func(r ports.Repositories) error {
		created, err := r.Organizations().Create(ctx, org)
		if err != nil {
			return err
		}
		org = created
		_, err = r.Members().Create(ctx, ports.Member{
			ID:             uuid.New(),
			UserID:         in.UserID,
			OrganizationID: created.ID,
			Role:           member.RoleAdmin,
			CreatedAt:      now,
		})
		return err
	})
	if err != nil {
		return ports.Organization{}, err
	}
	return org, nil
}

// --- List (caller's organizations) ---

type ListItem struct {
	ID   uuid.UUID   `json:"id" swaggertype:"string" format:"uuid"`
	Name string      `json:"name"`
	Slug string      `json:"slug"`
	Role member.Role `json:"role" swaggertype:"string"`
}

type ListUseCase struct{ orgs ports.OrganizationRepository }

func NewListUseCase(orgs ports.OrganizationRepository) *ListUseCase {
	return &ListUseCase{orgs: orgs}
}

func (uc *ListUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]ListItem, error) {
	rows, err := uc.orgs.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]ListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, ListItem{ID: r.ID, Name: r.Name, Slug: r.Slug, Role: r.Role})
	}
	return items, nil
}

// --- Get ---

type GetUseCase struct {
	orgs    ports.OrganizationRepository
	members ports.MemberRepository
}

func NewGetUseCase(orgs ports.OrganizationRepository, members ports.MemberRepository) *GetUseCase {
	return &GetUseCase{orgs: orgs, members: members}
}

func (uc *GetUseCase) Execute(ctx context.Context, userID, orgID uuid.UUID) (ports.Organization, error) {
	org, err := uc.orgs.FindByID(ctx, orgID)
	if err != nil {
		return ports.Organization{}, err
	}
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleViewer); err != nil {
		return ports.Organization{}, err
	}
	return org, nil
}

// --- Update (name only; slug is immutable) ---

type UpdateUseCase struct {
	orgs    ports.OrganizationRepository
	members ports.MemberRepository
}

func NewUpdateUseCase(orgs ports.OrganizationRepository, members ports.MemberRepository) *UpdateUseCase {
	return &UpdateUseCase{orgs: orgs, members: members}
}

func (uc *UpdateUseCase) Execute(ctx context.Context, userID, orgID uuid.UUID, name string) (ports.Organization, error) {
	if _, err := uc.orgs.FindByID(ctx, orgID); err != nil {
		return ports.Organization{}, err
	}
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleAdmin); err != nil {
		return ports.Organization{}, err
	}
	name = strings.TrimSpace(name)
	if err := orgdomain.ValidateName(name); err != nil {
		return ports.Organization{}, err
	}
	return uc.orgs.Update(ctx, orgID, name)
}

// --- Delete (soft-delete org + cascade to members, atomically) ---

type DeleteUseCase struct{ tx ports.TxManager }

func NewDeleteUseCase(tx ports.TxManager) *DeleteUseCase { return &DeleteUseCase{tx: tx} }

// Execute runs existence check, authorization and the cascade inside a single
// transaction, so a concurrent delete cannot slip between the check and the
// writes (no TOCTOU). Returns 404 if the org is gone, 403 if the caller is not
// an ADMIN member.
func (uc *DeleteUseCase) Execute(ctx context.Context, userID, orgID uuid.UUID) error {
	return uc.tx.WithinTx(ctx, func(r ports.Repositories) error {
		if _, err := r.Organizations().FindByID(ctx, orgID); err != nil {
			return err
		}
		if _, err := authz.RequireOrgRole(ctx, r.Members(), userID, orgID, member.RoleAdmin); err != nil {
			return err
		}
		if err := r.Organizations().SoftDelete(ctx, orgID); err != nil {
			return err
		}
		return r.Members().SoftDeleteByOrg(ctx, orgID)
	})
}
