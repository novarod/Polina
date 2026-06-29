package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	appws "github.com/novarod/polina/apps/api/internal/application/workspace"
	"github.com/novarod/polina/apps/api/internal/ports"
)

type WorkspaceHandler struct {
	create *appws.CreateUseCase
	list   *appws.ListUseCase
	get    *appws.GetUseCase
	update *appws.UpdateUseCase
	delete *appws.DeleteUseCase
}

func NewWorkspaceHandler(
	create *appws.CreateUseCase,
	list *appws.ListUseCase,
	get *appws.GetUseCase,
	update *appws.UpdateUseCase,
	del *appws.DeleteUseCase,
) *WorkspaceHandler {
	return &WorkspaceHandler{create: create, list: list, get: get, update: update, delete: del}
}

type createWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type workspaceResponse struct {
	ID             uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	OrganizationID uuid.UUID `json:"organization_id" swaggertype:"string" format:"uuid"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
}

func toWorkspaceResponse(w ports.Workspace) workspaceResponse {
	return workspaceResponse{
		ID:             w.ID,
		OrganizationID: w.OrganizationID,
		Name:           w.Name,
		Description:    w.Description,
		CreatedAt:      w.CreatedAt,
	}
}

func orgParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "invalid organization id")
	}
	return id, nil
}

func workspaceParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("workspaceID"))
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "invalid workspace id")
	}
	return id, nil
}

// @Summary   Create a workspace
// @Tags      workspaces
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id       path      string                  true  "Organization ID (uuid)"
// @Param     payload  body      createWorkspaceRequest  true  "Workspace payload"
// @Success   201      {object}  workspaceResponse
// @Failure   403      {object}  map[string]string
// @Failure   422      {object}  map[string]string
// @Router    /organizations/{id}/workspaces [post]
func (h *WorkspaceHandler) Create(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	var req createWorkspaceRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ws, err := h.create.Execute(c.Request().Context(), appws.CreateInput{
		UserID: claims.UserID, OrgID: orgID, Name: req.Name, Description: req.Description,
	})
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusCreated, toWorkspaceResponse(ws))
}

// @Summary   List workspaces
// @Tags      workspaces
// @Security  BearerAuth
// @Produce   json
// @Param     id   path      string  true  "Organization ID (uuid)"
// @Success   200  {array}   workspaceResponse
// @Failure   403  {object}  map[string]string
// @Router    /organizations/{id}/workspaces [get]
func (h *WorkspaceHandler) List(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	list, err := h.list.Execute(c.Request().Context(), claims.UserID, orgID)
	if err != nil {
		return mapError(err)
	}
	out := make([]workspaceResponse, 0, len(list))
	for _, w := range list {
		out = append(out, toWorkspaceResponse(w))
	}
	return c.JSON(http.StatusOK, out)
}

// @Summary   Get a workspace
// @Tags      workspaces
// @Security  BearerAuth
// @Produce   json
// @Param     id           path      string  true  "Organization ID (uuid)"
// @Param     workspaceID  path      string  true  "Workspace ID (uuid)"
// @Success   200          {object}  workspaceResponse
// @Failure   403          {object}  map[string]string
// @Failure   404          {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID} [get]
func (h *WorkspaceHandler) Get(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	wid, err := workspaceParam(c)
	if err != nil {
		return err
	}
	ws, err := h.get.Execute(c.Request().Context(), claims.UserID, orgID, wid)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, toWorkspaceResponse(ws))
}

// @Summary   Update a workspace
// @Tags      workspaces
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id           path      string                  true  "Organization ID (uuid)"
// @Param     workspaceID  path      string                  true  "Workspace ID (uuid)"
// @Param     payload      body      updateWorkspaceRequest  true  "Fields to update"
// @Success   200          {object}  workspaceResponse
// @Failure   403          {object}  map[string]string
// @Failure   404          {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID} [patch]
func (h *WorkspaceHandler) Update(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	wid, err := workspaceParam(c)
	if err != nil {
		return err
	}
	var req updateWorkspaceRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ws, err := h.update.Execute(c.Request().Context(), appws.UpdateInput{
		UserID: claims.UserID, OrgID: orgID, WorkspaceID: wid, Name: req.Name, Description: req.Description,
	})
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, toWorkspaceResponse(ws))
}

// @Summary   Delete a workspace
// @Tags      workspaces
// @Security  BearerAuth
// @Param     id           path  string  true  "Organization ID (uuid)"
// @Param     workspaceID  path  string  true  "Workspace ID (uuid)"
// @Success   204  "no content"
// @Failure   403  {object}  map[string]string
// @Failure   404  {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID} [delete]
func (h *WorkspaceHandler) Delete(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	wid, err := workspaceParam(c)
	if err != nil {
		return err
	}
	if err := h.delete.Execute(c.Request().Context(), claims.UserID, orgID, wid); err != nil {
		return mapError(err)
	}
	return c.NoContent(http.StatusNoContent)
}
