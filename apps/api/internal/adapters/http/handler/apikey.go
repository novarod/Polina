package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	appapikey "github.com/novarod/polina/apps/api/internal/application/apikey"
	"github.com/novarod/polina/apps/api/internal/ports"
)

type APIKeyHandler struct {
	create *appapikey.CreateUseCase
	list   *appapikey.ListUseCase
	revoke *appapikey.RevokeUseCase
}

func NewAPIKeyHandler(create *appapikey.CreateUseCase, list *appapikey.ListUseCase, revoke *appapikey.RevokeUseCase) *APIKeyHandler {
	return &APIKeyHandler{create: create, list: list, revoke: revoke}
}

type createAPIKeyRequest struct {
	Name string `json:"name"`
}

type createAPIKeyResponse struct {
	ID        uuid.UUID `json:"id" swaggertype:"string" format:"uuid"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}

type apiKeyResponse struct {
	ID         uuid.UUID  `json:"id" swaggertype:"string" format:"uuid"`
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
}

func toAPIKeyResponse(k ports.OrganizationAPIKey) apiKeyResponse {
	return apiKeyResponse{
		ID:         k.ID,
		Name:       k.Name,
		LastUsedAt: k.LastUsedAt,
		CreatedAt:  k.CreatedAt,
		RevokedAt:  k.RevokedAt,
	}
}

func apiKeyParam(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("keyID"))
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "invalid api key id")
	}
	return id, nil
}

// @Summary   Create an organization API key
// @Tags      api-keys
// @Security  BearerAuth
// @Accept    json
// @Produce   json
// @Param     id       path      string               true  "Organization ID (uuid)"
// @Param     payload  body      createAPIKeyRequest  true  "API key payload"
// @Success   201      {object}  createAPIKeyResponse  "the raw key is returned only here"
// @Failure   403      {object}  map[string]string
// @Failure   422      {object}  map[string]string
// @Router    /organizations/{id}/api-keys [post]
func (h *APIKeyHandler) Create(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	var req createAPIKeyRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	res, err := h.create.Execute(c.Request().Context(), appapikey.CreateInput{
		UserID: claims.UserID, OrgID: orgID, Name: req.Name,
	})
	if err != nil {
		return mapError(err)
	}
	return c.JSON(http.StatusCreated, createAPIKeyResponse{
		ID:        res.Key.ID,
		Name:      res.Key.Name,
		Key:       res.Raw,
		CreatedAt: res.Key.CreatedAt,
	})
}

// @Summary   List organization API keys
// @Tags      api-keys
// @Security  BearerAuth
// @Produce   json
// @Param     id   path      string  true  "Organization ID (uuid)"
// @Success   200  {array}   apiKeyResponse
// @Failure   403  {object}  map[string]string
// @Router    /organizations/{id}/api-keys [get]
func (h *APIKeyHandler) List(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	list, err := h.list.Execute(c.Request().Context(), claims.UserID, orgID)
	if err != nil {
		return mapError(err)
	}
	out := make([]apiKeyResponse, 0, len(list))
	for _, k := range list {
		out = append(out, toAPIKeyResponse(k))
	}
	return c.JSON(http.StatusOK, out)
}

// @Summary   Revoke an organization API key
// @Tags      api-keys
// @Security  BearerAuth
// @Param     id     path  string  true  "Organization ID (uuid)"
// @Param     keyID  path  string  true  "API key ID (uuid)"
// @Success   204  "no content"
// @Failure   403  {object}  map[string]string
// @Failure   404  {object}  map[string]string
// @Router    /organizations/{id}/api-keys/{keyID} [delete]
func (h *APIKeyHandler) Revoke(c echo.Context) error {
	claims := httpmw.MustGetClaims(c)
	orgID, err := orgParam(c)
	if err != nil {
		return err
	}
	keyID, err := apiKeyParam(c)
	if err != nil {
		return err
	}
	if err := h.revoke.Execute(c.Request().Context(), claims.UserID, orgID, keyID); err != nil {
		return mapError(err)
	}
	return c.NoContent(http.StatusNoContent)
}
