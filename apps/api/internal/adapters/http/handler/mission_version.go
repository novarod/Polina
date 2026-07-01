package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	appmission "github.com/novarod/polina/apps/api/internal/application/mission"
	"github.com/novarod/polina/apps/api/internal/ports"
)

type MissionVersionHandler struct {
	publish      *appmission.PublishUseCase
	listVersions *appmission.ListVersionsUseCase
	getVersion   *appmission.GetVersionUseCase
}

func NewMissionVersionHandler(
	publish *appmission.PublishUseCase,
	listVersions *appmission.ListVersionsUseCase,
	getVersion *appmission.GetVersionUseCase,
) *MissionVersionHandler {
	return &MissionVersionHandler{publish: publish, listVersions: listVersions, getVersion: getVersion}
}

type publishResponse struct {
	MissionID  uuid.UUID `json:"mission_id" swaggertype:"string" format:"uuid"`
	Version    int       `json:"version"`
	Hash       string    `json:"hash"`
	Status     string    `json:"status"`
	ActiveHash *string   `json:"active_hash"`
}

type versionSummary struct {
	ID            uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	VersionNumber int       `json:"version_number"`
	Hash          string    `json:"hash"`
	PublishedByID uuid.UUID `json:"published_by_id" swaggertype:"string" format:"uuid"`
	CreatedAt     time.Time `json:"created_at"`
}

type versionDetail struct {
	versionSummary
	MissionData json.RawMessage `json:"mission_data" swaggertype:"object"`
}

func toVersionSummary(v ports.MissionVersion) versionSummary {
	return versionSummary{
		ID:            v.ID,
		VersionNumber: v.VersionNumber,
		Hash:          v.Hash,
		PublishedByID: v.PublishedByID,
		CreatedAt:     v.CreatedAt,
	}
}

// @Summary   Publish a mission version
// @Tags      missions
// @Security  BearerAuth
// @Produce   json
// @Param     id           path      string  true  "Organization ID (uuid)"
// @Param     workspaceID  path      string  true  "Workspace ID (uuid)"
// @Param     missionID    path      string  true  "Mission ID (uuid)"
// @Success   200          {object}  publishResponse
// @Failure   403          {object}  map[string]string
// @Failure   404          {object}  map[string]string
// @Failure   422          {object}  map[string]string  "invalid graph"
// @Router    /organizations/{id}/workspaces/{workspaceID}/missions/{missionID}/publish [post]
func (h *MissionVersionHandler) Publish(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, wsID, mID, err := missionPathParams(c)
	if err != nil {
		return err
	}
	res, err := h.publish.Execute(c.Request().Context(), appmission.PublishInput{
		UserID: claims.UserID, OrgID: orgID, WorkspaceID: wsID, MissionID: mID,
	})
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, publishResponse{
		MissionID:  res.Mission.ID,
		Version:    res.Version.VersionNumber,
		Hash:       res.Version.Hash,
		Status:     res.Mission.Status,
		ActiveHash: res.Mission.ActiveHash,
	})
}

// @Summary   List mission versions
// @Tags      missions
// @Security  BearerAuth
// @Produce   json
// @Param     id           path      string  true  "Organization ID (uuid)"
// @Param     workspaceID  path      string  true  "Workspace ID (uuid)"
// @Param     missionID    path      string  true  "Mission ID (uuid)"
// @Success   200          {array}   versionSummary
// @Failure   403          {object}  map[string]string
// @Failure   404          {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID}/missions/{missionID}/versions [get]
func (h *MissionVersionHandler) ListVersions(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, wsID, mID, err := missionPathParams(c)
	if err != nil {
		return err
	}
	list, err := h.listVersions.Execute(c.Request().Context(), claims.UserID, orgID, wsID, mID)
	if err != nil {
		return mapError(err)
	}
	out := make([]versionSummary, 0, len(list))
	for _, v := range list {
		out = append(out, toVersionSummary(v))
	}
	return c.JSON(http.StatusOK, out)
}

// @Summary   Get a mission version (by hash)
// @Tags      missions
// @Security  BearerAuth
// @Produce   json
// @Param     id           path      string  true  "Organization ID (uuid)"
// @Param     workspaceID  path      string  true  "Workspace ID (uuid)"
// @Param     missionID    path      string  true  "Mission ID (uuid)"
// @Param     hash         path      string  true  "Version content hash"
// @Success   200          {object}  versionDetail
// @Failure   403          {object}  map[string]string
// @Failure   404          {object}  map[string]string
// @Router    /organizations/{id}/workspaces/{workspaceID}/missions/{missionID}/versions/{hash} [get]
func (h *MissionVersionHandler) GetVersion(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, wsID, mID, err := missionPathParams(c)
	if err != nil {
		return err
	}
	v, err := h.getVersion.Execute(c.Request().Context(), claims.UserID, orgID, wsID, mID, c.Param("hash"))
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, versionDetail{
		versionSummary: toVersionSummary(v),
		MissionData:    v.MissionData,
	})
}
