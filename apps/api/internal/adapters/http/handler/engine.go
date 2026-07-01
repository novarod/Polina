package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	appengine "github.com/novarod/polina/apps/api/internal/application/engine"
)

type EngineHandler struct {
	activeHash     *appengine.GetActiveHashUseCase
	activeContract *appengine.GetActiveContractUseCase
}

func NewEngineHandler(activeHash *appengine.GetActiveHashUseCase, activeContract *appengine.GetActiveContractUseCase) *EngineHandler {
	return &EngineHandler{activeHash: activeHash, activeContract: activeContract}
}

type engineHashResponse struct {
	Hash string `json:"hash"`
}

// @Summary   Poll the active version hash of a mission (engine)
// @Tags      engine
// @Security  ApiKeyAuth
// @Produce   json
// @Param     missionID  path      string  true  "Mission ID (uuid)"
// @Success   200        {object}  engineHashResponse
// @Failure   401        {object}  map[string]string
// @Failure   404        {object}  map[string]string
// @Router    /engine/missions/{missionID}/active/hash [get]
func (h *EngineHandler) ActiveHash(c echo.Context) error {
	orgID := httpmw.MustGetEngineOrg(c)
	missionID, err := missionParam(c)
	if err != nil {
		return err
	}
	hash, err := h.activeHash.Execute(c.Request().Context(), orgID, missionID)
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusOK, engineHashResponse{Hash: hash})
}

// @Summary   Fetch the active compiled contract of a mission (engine)
// @Tags      engine
// @Security  ApiKeyAuth
// @Produce   json
// @Param     missionID  path      string  true  "Mission ID (uuid)"
// @Success   200        {object}  map[string]interface{}  "compiled mission contract"
// @Failure   401        {object}  map[string]string
// @Failure   404        {object}  map[string]string
// @Router    /engine/missions/{missionID}/active [get]
func (h *EngineHandler) ActiveContract(c echo.Context) error {
	orgID := httpmw.MustGetEngineOrg(c)
	missionID, err := missionParam(c)
	if err != nil {
		return err
	}
	contract, err := h.activeContract.Execute(c.Request().Context(), orgID, missionID)
	if err != nil {
		return mapError(err)
	}
	return c.JSONBlob(http.StatusOK, contract)
}
