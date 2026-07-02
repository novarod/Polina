package apikey

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/novarod/polina/apps/api/internal/application/authz"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
	"github.com/novarod/polina/apps/api/pkg/hash"
)

const nameMax = 255

func generateRawKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("apikey: generate: %w", err)
	}
	return "pol_" + base64.RawURLEncoding.EncodeToString(b), nil
}

// --- Create ---

type CreateInput struct {
	UserID uuid.UUID
	OrgID  uuid.UUID
	Name   string
}

type CreateResult struct {
	Key ports.OrganizationAPIKey
	Raw string
}

type CreateUseCase struct {
	keys    ports.OrganizationAPIKeyRepository
	members ports.MemberRepository
}

func NewCreateUseCase(keys ports.OrganizationAPIKeyRepository, members ports.MemberRepository) *CreateUseCase {
	return &CreateUseCase{keys: keys, members: members}
}

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (CreateResult, error) {
	caller, err := authz.RequireOrgRole(ctx, uc.members, in.UserID, in.OrgID, member.RoleAdmin)
	if err != nil {
		return CreateResult{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" || utf8.RuneCountInString(name) > nameMax {
		return CreateResult{}, apierr.Validation("name", "name must be between 1 and 255 characters")
	}

	raw, err := generateRawKey()
	if err != nil {
		return CreateResult{}, err
	}
	created, err := uc.keys.Create(ctx, ports.OrganizationAPIKey{
		ID:             uuid.New(),
		OrganizationID: in.OrgID,
		Name:           name,
		KeyHash:        hash.APIKey(raw),
		CreatedByID:    caller.ID,
		CreatedAt:      time.Now(),
	})
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Key: created, Raw: raw}, nil
}

// --- List ---

type ListUseCase struct {
	keys    ports.OrganizationAPIKeyRepository
	members ports.MemberRepository
}

func NewListUseCase(keys ports.OrganizationAPIKeyRepository, members ports.MemberRepository) *ListUseCase {
	return &ListUseCase{keys: keys, members: members}
}

func (uc *ListUseCase) Execute(ctx context.Context, userID, orgID uuid.UUID) ([]ports.OrganizationAPIKey, error) {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleAdmin); err != nil {
		return nil, err
	}
	return uc.keys.ListByOrg(ctx, orgID)
}

// --- Revoke ---

type RevokeUseCase struct {
	keys    ports.OrganizationAPIKeyRepository
	members ports.MemberRepository
}

func NewRevokeUseCase(keys ports.OrganizationAPIKeyRepository, members ports.MemberRepository) *RevokeUseCase {
	return &RevokeUseCase{keys: keys, members: members}
}

func (uc *RevokeUseCase) Execute(ctx context.Context, userID, orgID, keyID uuid.UUID) error {
	if _, err := authz.RequireOrgRole(ctx, uc.members, userID, orgID, member.RoleAdmin); err != nil {
		return err
	}
	return uc.keys.Revoke(ctx, keyID, orgID)
}
