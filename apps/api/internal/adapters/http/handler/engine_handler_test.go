package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	appengine "github.com/novarod/polina/apps/api/internal/application/engine"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

func newEngineApp(keyRepo ports.OrganizationAPIKeyRepository, missions ports.MissionRepository, versions ports.MissionVersionRepository) *echo.Echo {
	h := handler.NewEngineHandler(
		appengine.NewGetActiveHashUseCase(missions),
		appengine.NewGetActiveContractUseCase(versions),
	)
	e := echo.New()
	g := e.Group("/engine", httpmw.APIKeyAuth(keyRepo))
	g.GET("/missions/:missionID/active/hash", h.ActiveHash)
	g.GET("/missions/:missionID/active", h.ActiveContract)
	return e
}

func engineGET(app *echo.Echo, path, apiKey string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
	}
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	return rec
}

func validKey() *fakeAPIKeyRepo {
	return &fakeAPIKeyRepo{active: ports.OrganizationAPIKey{ID: uuid.New(), OrganizationID: uuid.New()}}
}

func ptr(s string) *string { return &s }

func TestEngineHandler_ActiveHash_200(t *testing.T) {
	mid := uuid.New()
	missions := &fakeMissionRepo{findByID: ports.Mission{ActiveHash: ptr("abc123")}}
	app := newEngineApp(validKey(), missions, &fakeVersionRepo{})
	rec := engineGET(app, "/engine/missions/"+mid.String()+"/active/hash", "pol_ok")
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]string
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "abc123", body["hash"])
}

func TestEngineHandler_ActiveHash_404_Unpublished(t *testing.T) {
	missions := &fakeMissionRepo{findByIDErr: apierr.NotFound("active mission version")}
	app := newEngineApp(validKey(), missions, &fakeVersionRepo{})
	rec := engineGET(app, "/engine/missions/"+uuid.New().String()+"/active/hash", "pol_ok")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEngineHandler_ActiveContract_200(t *testing.T) {
	versions := &fakeVersionRepo{active: &ports.MissionVersion{MissionData: json.RawMessage(`{"mission_id":"m","nodes":{}}`)}}
	app := newEngineApp(validKey(), &fakeMissionRepo{}, versions)
	rec := engineGET(app, "/engine/missions/"+uuid.New().String()+"/active", "pol_ok")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"mission_id":"m","nodes":{}}`, rec.Body.String())
}

func TestEngineHandler_ActiveContract_404(t *testing.T) {
	app := newEngineApp(validKey(), &fakeMissionRepo{}, &fakeVersionRepo{})
	rec := engineGET(app, "/engine/missions/"+uuid.New().String()+"/active", "pol_ok")
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestEngineHandler_MissingKey_401(t *testing.T) {
	app := newEngineApp(&fakeAPIKeyRepo{findErr: apierr.NotFound("api key")}, &fakeMissionRepo{}, &fakeVersionRepo{})
	rec := engineGET(app, "/engine/missions/"+uuid.New().String()+"/active/hash", "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
