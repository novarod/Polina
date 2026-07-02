package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	apporg "github.com/novarod/polina/apps/api/internal/application/organization"
	"github.com/novarod/polina/apps/api/internal/application/token"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

// --- fakes (implement the ports interfaces) ---

type fakeOrgRepo struct {
	createErr   error
	findByID    ports.Organization
	findByIDErr error
}

func (f *fakeOrgRepo) Create(_ context.Context, o ports.Organization) (ports.Organization, error) {
	if f.createErr != nil {
		return ports.Organization{}, f.createErr
	}
	return o, nil
}
func (f *fakeOrgRepo) FindByID(_ context.Context, _ uuid.UUID) (ports.Organization, error) {
	return f.findByID, f.findByIDErr
}
func (f *fakeOrgRepo) ListByUserID(_ context.Context, _ uuid.UUID) ([]ports.OrganizationWithRole, error) {
	return nil, nil
}
func (f *fakeOrgRepo) Update(_ context.Context, _ uuid.UUID, name string) (ports.Organization, error) {
	return ports.Organization{Name: name}, nil
}
func (f *fakeOrgRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return nil }

type fakeMemberRepo struct {
	found   ports.Member
	findErr error
}

func (f *fakeMemberRepo) Create(_ context.Context, m ports.Member) (ports.Member, error) {
	return m, nil
}
func (f *fakeMemberRepo) FindByUserAndOrg(_ context.Context, _, _ uuid.UUID) (ports.Member, error) {
	return f.found, f.findErr
}
func (f *fakeMemberRepo) SoftDeleteByOrg(_ context.Context, _ uuid.UUID) error { return nil }

type fakeRepos struct {
	orgs     ports.OrganizationRepository
	members  ports.MemberRepository
	missions ports.MissionRepository
	versions ports.MissionVersionRepository
}

func (f *fakeRepos) Users() ports.UserRepository                     { return nil }
func (f *fakeRepos) Members() ports.MemberRepository                 { return f.members }
func (f *fakeRepos) Organizations() ports.OrganizationRepository     { return f.orgs }
func (f *fakeRepos) Missions() ports.MissionRepository               { return f.missions }
func (f *fakeRepos) MissionVersions() ports.MissionVersionRepository { return f.versions }
func (f *fakeRepos) Workspaces() ports.WorkspaceRepository           { return nil }
func (f *fakeRepos) OrganizationAPIKeys() ports.OrganizationAPIKeyRepository {
	return nil
}

type fakeTxManager struct{ repos ports.Repositories }

func (f *fakeTxManager) WithinTx(_ context.Context, fn func(ports.Repositories) error) error {
	return fn(f.repos)
}

// noopValidator satisfies echo.Validator; the org DTOs carry no validate tags
// (validation lives in the domain), so this just needs to exist.
type noopValidator struct{}

func (noopValidator) Validate(any) error { return nil }

// newHandler wires the real use cases over the given fakes.
func newHandler(orgs ports.OrganizationRepository, members ports.MemberRepository) *handler.OrganizationHandler {
	tx := &fakeTxManager{repos: &fakeRepos{orgs: orgs, members: members}}
	return handler.NewOrganizationHandler(
		apporg.NewCreateUseCase(tx),
		apporg.NewListUseCase(orgs),
		apporg.NewGetUseCase(orgs, members),
		apporg.NewUpdateUseCase(orgs, members),
		apporg.NewDeleteUseCase(tx),
	)
}

// invoke runs h(c) and returns the resulting HTTP status, mapping a returned
// *echo.HTTPError to its code (as the real error handler would).
func invoke(t *testing.T, h func(echo.Context) error, c echo.Context, rec *httptest.ResponseRecorder) int {
	t.Helper()
	if err := h(c); err != nil {
		var he *echo.HTTPError
		require.ErrorAs(t, err, &he)
		return he.Code
	}
	return rec.Code
}

func newCtx(t *testing.T, method, target, body string, userID uuid.UUID) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	e := echo.New()
	e.Validator = noopValidator{}
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("claims", &token.Claims{UserID: userID})
	return c, rec
}

func TestHandler_Create_201(t *testing.T) {
	h := newHandler(&fakeOrgRepo{}, &fakeMemberRepo{})
	c, rec := newCtx(t, http.MethodPost, "/organizations", `{"name":"Acme Studios","slug":"acme"}`, uuid.New())
	assert.Equal(t, http.StatusCreated, invoke(t, h.Create, c, rec))
	assert.Contains(t, rec.Body.String(), `"slug":"acme"`)
}

func TestHandler_Create_422_DuplicateSlug(t *testing.T) {
	orgs := &fakeOrgRepo{createErr: apierr.Validation("slug", "slug already in use")}
	h := newHandler(orgs, &fakeMemberRepo{})
	c, rec := newCtx(t, http.MethodPost, "/organizations", `{"name":"Acme","slug":"acme"}`, uuid.New())
	assert.Equal(t, http.StatusUnprocessableEntity, invoke(t, h.Create, c, rec))
}

func TestHandler_Create_400_BadBody(t *testing.T) {
	h := newHandler(&fakeOrgRepo{}, &fakeMemberRepo{})
	c, rec := newCtx(t, http.MethodPost, "/organizations", `{not-json`, uuid.New())
	assert.Equal(t, http.StatusBadRequest, invoke(t, h.Create, c, rec))
}

func TestHandler_Get_200_Member(t *testing.T) {
	id := uuid.New()
	orgs := &fakeOrgRepo{findByID: ports.Organization{ID: id, Name: "Acme", Slug: "acme"}}
	members := &fakeMemberRepo{found: ports.Member{Role: member.RoleViewer}}
	h := newHandler(orgs, members)
	c, rec := newCtx(t, http.MethodGet, "/organizations/"+id.String(), "", uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	assert.Equal(t, http.StatusOK, invoke(t, h.Get, c, rec))
}

func TestHandler_Get_404_NotFound(t *testing.T) {
	orgs := &fakeOrgRepo{findByIDErr: apierr.NotFound("organization")}
	h := newHandler(orgs, &fakeMemberRepo{})
	id := uuid.New()
	c, rec := newCtx(t, http.MethodGet, "/organizations/"+id.String(), "", uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	assert.Equal(t, http.StatusNotFound, invoke(t, h.Get, c, rec))
}

func TestHandler_Get_403_NotMember(t *testing.T) {
	id := uuid.New()
	orgs := &fakeOrgRepo{findByID: ports.Organization{ID: id}}
	members := &fakeMemberRepo{findErr: apierr.NotFound("member")}
	h := newHandler(orgs, members)
	c, rec := newCtx(t, http.MethodGet, "/organizations/"+id.String(), "", uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	assert.Equal(t, http.StatusForbidden, invoke(t, h.Get, c, rec))
}

func TestHandler_Get_400_BadID(t *testing.T) {
	h := newHandler(&fakeOrgRepo{}, &fakeMemberRepo{})
	c, rec := newCtx(t, http.MethodGet, "/organizations/not-a-uuid", "", uuid.New())
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	assert.Equal(t, http.StatusBadRequest, invoke(t, h.Get, c, rec))
}

func TestHandler_Delete_204_Admin(t *testing.T) {
	id := uuid.New()
	orgs := &fakeOrgRepo{findByID: ports.Organization{ID: id}}
	members := &fakeMemberRepo{found: ports.Member{Role: member.RoleAdmin}}
	h := newHandler(orgs, members)
	c, rec := newCtx(t, http.MethodDelete, "/organizations/"+id.String(), "", uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	assert.Equal(t, http.StatusNoContent, invoke(t, h.Delete, c, rec))
}

func TestHandler_Update_403_NonAdmin(t *testing.T) {
	id := uuid.New()
	orgs := &fakeOrgRepo{findByID: ports.Organization{ID: id}}
	members := &fakeMemberRepo{found: ports.Member{Role: member.RoleViewer}}
	h := newHandler(orgs, members)
	c, rec := newCtx(t, http.MethodPatch, "/organizations/"+id.String(), `{"name":"New"}`, uuid.New())
	c.SetParamNames("id")
	c.SetParamValues(id.String())
	assert.Equal(t, http.StatusForbidden, invoke(t, h.Update, c, rec))
}
