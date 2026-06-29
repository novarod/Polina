package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	appmission "github.com/novarod/polina/apps/api/internal/application/mission"
	"github.com/novarod/polina/apps/api/internal/ports"
)

type MissionHandler struct {
	create      *appmission.CreateUseCase
	list        *appmission.ListUseCase
	get         *appmission.GetUseCase
	update      *appmission.UpdateUseCase
	updateGraph *appmission.UpdateGraphUseCase
	delete      *appmission.DeleteUseCase
}

func NewMissionHandler(
	create *appmission.CreateUseCase,
	list *appmission.ListUseCase,
	get *appmission.GetUseCase,
	update *appmission.UpdateUseCase,
	updateGraph *appmission.UpdateGraphUseCase,
	del *appmission.DeleteUseCase,
) *MissionHandler {
	return &MissionHandler{create: create, list: list, get: get, update: update, updateGraph: updateGraph, delete: del}
}

type createMissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateMissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type missionResponse struct {
	ID             uuid.UUID       `json:"id" swaggertype:"string" format:"uuid"`
	OrganizationID uuid.UUID       `json:"organization_id" swaggertype:"string" format:"uuid"`
	WorkspaceID    uuid.UUID       `json:"workspace_id" swaggertype:"string" format:"uuid"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Status         string          `json:"status"`
	ActiveHash     *string         `json:"active_hash"`
	Graph          json.RawMessage `json:"graph" swaggertype:"object"`
	CreatedByID    uuid.UUID       `json:"created_by_id" swaggertype:"string" format:"uuid"`
	CreatedAt      time.Time       `json:"created_at"`
}

func toMissionResponse(m ports.Mission) missionResponse {
	return missionResponse{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		WorkspaceID:    m.WorkspaceID,
		Name:           m.Name,
		Description:    m.Description,
		Status:         m.Status,
		ActiveHash:     m.ActiveHash,
		Graph:          m.Graph,
		CreatedByID:    m.CreatedByID,
		CreatedAt:      m.CreatedAt,
	}
}

func missionParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("missionID"))
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "invalid mission id")
	}
	return id, nil
}

// Create adds a mission to a workspace (DESIGNER+).
//
// @Summary   Create a mission
// @Tags      missions
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id           path      string                true  "Organization ID (uuid)"
// @Param     workspaceID  path      string                true  "Workspace ID (uuid)"
// @Param     payload      body      createMissionRequest  true  "Mission payload"
// @Success   201          {object}  missionResponse
// @Failure   403          {object}  map[string]string
// @Failure   404          {object}  map[string]string
// @Failure   422          {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID}/missions [post]
func (h *MissionHandler) Create(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	wsID, err := workspaceParam(c)
	if err != nil {
		return err
	}
	var req createMissionRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	m, err := h.create.Execute(c.Request().Context(), appmission.CreateInput{
		UserID: claims.UserID, OrgID: orgID, WorkspaceID: wsID, Name: req.Name, Description: req.Description,
	})
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusCreated, toMissionResponse(m))
}

// List returns the missions of a workspace (VIEWER+).
//
// @Summary   List missions
// @Tags      missions
// @Security  BearerAuth
// @Produce   json
// @Param     id           path      string  true  "Organization ID (uuid)"
// @Param     workspaceID  path      string  true  "Workspace ID (uuid)"
// @Success   200          {array}   missionResponse
// @Failure   403          {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID}/missions [get]
func (h *MissionHandler) List(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	wsID, err := workspaceParam(c)
	if err != nil {
		return err
	}
	list, err := h.list.Execute(c.Request().Context(), claims.UserID, orgID, wsID)
	if err != nil {
		return mapError(err)
	}
	out := make([]missionResponse, 0, len(list))
	for _, m := range list {
		out = append(out, toMissionResponse(m))
	}
	return c.JSON(http.StatusOK, out)
}

// Get returns one mission including its graph (VIEWER+).
//
// @Summary   Get a mission
// @Tags      missions
// @Security  BearerAuth
// @Produce   json
// @Param     id           path      string  true  "Organization ID (uuid)"
// @Param     workspaceID  path      string  true  "Workspace ID (uuid)"
// @Param     missionID    path      string  true  "Mission ID (uuid)"
// @Success   200          {object}  missionResponse
// @Failure   403          {object}  map[string]string
// @Failure   404          {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID}/missions/{missionID} [get]
func (h *MissionHandler) Get(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, wsID, mID, err := missionPathParams(c)
	if err != nil {
		return err
	}
	m, err := h.get.Execute(c.Request().Context(), claims.UserID, orgID, wsID, mID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, toMissionResponse(m))
}

// Update changes a mission's name/description (DESIGNER+).
//
// @Summary   Update a mission
// @Tags      missions
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id           path      string                true  "Organization ID (uuid)"
// @Param     workspaceID  path      string                true  "Workspace ID (uuid)"
// @Param     missionID    path      string                true  "Mission ID (uuid)"
// @Param     payload      body      updateMissionRequest  true  "Fields to update"
// @Success   200          {object}  missionResponse
// @Failure   403          {object}  map[string]string
// @Failure   404          {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID}/missions/{missionID} [patch]
func (h *MissionHandler) Update(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, wsID, mID, err := missionPathParams(c)
	if err != nil {
		return err
	}
	var req updateMissionRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	m, err := h.update.Execute(c.Request().Context(), appmission.UpdateInput{
		UserID: claims.UserID, OrgID: orgID, WorkspaceID: wsID, MissionID: mID, Name: req.Name, Description: req.Description,
	})
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, toMissionResponse(m))
}

// UpdateGraph replaces a mission's quest graph after structural DAG validation (DESIGNER+).
//
// @Summary   Update a mission graph
// @Tags      missions
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id           path      string  true  "Organization ID (uuid)"
// @Param     workspaceID  path      string  true  "Workspace ID (uuid)"
// @Param     missionID    path      string  true  "Mission ID (uuid)"
// @Param     payload      body      object  true  "Graph object {nodes, edges}"
// @Success   200          {object}  missionResponse
// @Failure   403          {object}  map[string]string
// @Failure   404          {object}  map[string]string
// @Failure   422          {object}  map[string]string  "invalid graph"
// @Router    /organizations/{id}/workspaces/{workspaceID}/missions/{missionID}/graph [put]
func (h *MissionHandler) UpdateGraph(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, wsID, mID, err := missionPathParams(c)
	if err != nil {
		return err
	}
	raw, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "could not read request body")
	}
	m, err := h.updateGraph.Execute(c.Request().Context(), appmission.UpdateGraphInput{
		UserID: claims.UserID, OrgID: orgID, WorkspaceID: wsID, MissionID: mID, Graph: raw,
	})
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, toMissionResponse(m))
}

// Delete soft-deletes a mission (DESIGNER+).
//
// @Summary   Delete a mission
// @Tags      missions
// @Security  BearerAuth
// @Param     id           path  string  true  "Organization ID (uuid)"
// @Param     workspaceID  path  string  true  "Workspace ID (uuid)"
// @Param     missionID    path  string  true  "Mission ID (uuid)"
// @Success   204  "no content"
// @Failure   403  {object}  map[string]string
// @Failure   404  {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID}/missions/{missionID} [delete]
func (h *MissionHandler) Delete(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, wsID, mID, err := missionPathParams(c)
	if err != nil {
		return err
	}
	if err := h.delete.Execute(c.Request().Context(), claims.UserID, orgID, wsID, mID); err != nil {
		return mapError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// missionPathParams parses org, workspace and mission ids from the path.
func missionPathParams(c echo.Context) (orgID, wsID, missionID uuid.UUID, err error) {
	if orgID, err = orgParam(c); err != nil {
		return
	}
	if wsID, err = workspaceParam(c); err != nil {
		return
	}
	missionID, err = missionParam(c)
	return
}
