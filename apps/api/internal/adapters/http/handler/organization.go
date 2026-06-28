package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	apporg "github.com/novarod/polina/apps/api/internal/application/organization"
	"github.com/novarod/polina/apps/api/internal/ports"
)

type OrganizationHandler struct {
	create *apporg.CreateUseCase
	list   *apporg.ListUseCase
	get    *apporg.GetUseCase
	update *apporg.UpdateUseCase
	delete *apporg.DeleteUseCase
}

func NewOrganizationHandler(
	create *apporg.CreateUseCase,
	list *apporg.ListUseCase,
	get *apporg.GetUseCase,
	update *apporg.UpdateUseCase,
	del *apporg.DeleteUseCase,
) *OrganizationHandler {
	return &OrganizationHandler{create: create, list: list, get: get, update: update, delete: del}
}

type createOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type updateOrgRequest struct {
	Name string `json:"name"`
}

type orgResponse struct {
	ID        uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

func toOrgResponse(o ports.Organization) orgResponse {
	return orgResponse{ID: o.ID, Name: o.Name, Slug: o.Slug, CreatedAt: o.CreatedAt}
}

// @Summary   Create an organization
// @Tags      organizations
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     payload  body      createOrgRequest  true  "Organization payload"
// @Success   201      {object}  orgResponse
// @Failure   401      {object}  map[string]string
// @Failure   422      {object}  map[string]string  "validation error or slug already in use"
// @Router    /organizations [post]
func (h *OrganizationHandler) Create(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	var req createOrgRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	org, err := h.create.Execute(c.Request().Context(), apporg.CreateInput{
		UserID: claims.UserID,
		Name:   req.Name,
		Slug:   req.Slug,
	})
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusCreated, toOrgResponse(org))
}

// @Summary   List my organizations
// @Tags      organizations
// @Security  BearerAuth
// @Produce   json
// @Success   200  {array}   apporg.ListItem
// @Failure   401  {object}  map[string]string
// @Router    /organizations [get]
func (h *OrganizationHandler) List(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	items, err := h.list.Execute(c.Request().Context(), claims.UserID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, items)
}

// @Summary   Get an organization
// @Tags      organizations
// @Security  BearerAuth
// @Produce   json
// @Param     id   path      string  true  "Organization ID (uuid)"
// @Success   200  {object}  orgResponse
// @Failure   403  {object}  map[string]string
// @Failure   404  {object}  map[string]string
// @Router    /organizations/{id} [get]
func (h *OrganizationHandler) Get(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid organization id")
	}
	org, err := h.get.Execute(c.Request().Context(), claims.UserID, id)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, toOrgResponse(org))
}

// @Summary   Update an organization
// @Tags      organizations
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id       path      string            true  "Organization ID (uuid)"
// @Param     payload  body      updateOrgRequest  true  "Fields to update"
// @Success   200      {object}  orgResponse
// @Failure   403      {object}  map[string]string
// @Failure   404      {object}  map[string]string
// @Router    /organizations/{id} [patch]
func (h *OrganizationHandler) Update(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid organization id")
	}
	var req updateOrgRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	org, err := h.update.Execute(c.Request().Context(), claims.UserID, id, req.Name)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, toOrgResponse(org))
}

// @Summary   Delete an organization
// @Tags      organizations
// @Security  BearerAuth
// @Param     id  path  string  true  "Organization ID (uuid)"
// @Success   204  "no content"
// @Failure   403  {object}  map[string]string
// @Failure   404  {object}  map[string]string
// @Router    /organizations/{id} [delete]
func (h *OrganizationHandler) Delete(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid organization id")
	}
	if err := h.delete.Execute(c.Request().Context(), claims.UserID, id); err != nil {
		return mapError(err)
	}
	return c.NoContent(http.StatusNoContent)
}
