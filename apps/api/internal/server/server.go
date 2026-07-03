package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/novarod/polina/apps/api/docs"
	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	"github.com/novarod/polina/apps/api/internal/adapters/postgres"
	appapikey "github.com/novarod/polina/apps/api/internal/application/apikey"
	appauth "github.com/novarod/polina/apps/api/internal/application/auth"
	appengine "github.com/novarod/polina/apps/api/internal/application/engine"
	appmission "github.com/novarod/polina/apps/api/internal/application/mission"
	apporg "github.com/novarod/polina/apps/api/internal/application/organization"
	appws "github.com/novarod/polina/apps/api/internal/application/workspace"
)

const maxRequestBody = "1M"

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	idleTimeout       = 60 * time.Second
)

type Config struct {
	DBURL                    string
	JWTSecret                string
	JWTExpiryHours           int
	BcryptRounds             int
	Port                     string
	FrontendURL              string
	ThrottleLimit            int
	EngineThrottleLimit      int
	EngineLastUsedThrottleMs int
	Production               bool
	Logger                   *slog.Logger
}

type Server struct {
	echo *echo.Echo
	pool *pgxpool.Pool
	port string
}

func New(ctx context.Context, cfg Config) (*Server, error) {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// Store (repositories + transaction manager)
	store := postgres.NewStore(pool)
	userRepo := store.Users()
	orgRepo := store.Organizations()
	memberRepo := store.Members()

	// Use cases
	registerUC := appauth.NewRegisterUseCase(userRepo, cfg.BcryptRounds)
	loginUC := appauth.NewLoginUseCase(userRepo, memberRepo, cfg.JWTSecret, cfg.JWTExpiryHours, cfg.BcryptRounds)
	logoutAllUC := appauth.NewLogoutAllUseCase(userRepo)

	createOrgUC := apporg.NewCreateUseCase(store)
	listOrgUC := apporg.NewListUseCase(orgRepo)
	getOrgUC := apporg.NewGetUseCase(orgRepo, memberRepo)
	updateOrgUC := apporg.NewUpdateUseCase(orgRepo, memberRepo)
	deleteOrgUC := apporg.NewDeleteUseCase(store)

	wsRepo := store.Workspaces()
	createWsUC := appws.NewCreateUseCase(wsRepo, memberRepo)
	listWsUC := appws.NewListUseCase(wsRepo, memberRepo)
	getWsUC := appws.NewGetUseCase(wsRepo, memberRepo)
	updateWsUC := appws.NewUpdateUseCase(wsRepo, memberRepo)
	deleteWsUC := appws.NewDeleteUseCase(wsRepo, memberRepo)

	missionRepo := store.Missions()
	createMissionUC := appmission.NewCreateUseCase(missionRepo, wsRepo, memberRepo)
	listMissionUC := appmission.NewListUseCase(missionRepo, memberRepo)
	getMissionUC := appmission.NewGetUseCase(missionRepo, memberRepo)
	updateMissionUC := appmission.NewUpdateUseCase(missionRepo, memberRepo)
	updateMissionGraphUC := appmission.NewUpdateGraphUseCase(missionRepo, memberRepo)
	deleteMissionUC := appmission.NewDeleteUseCase(missionRepo, memberRepo)

	missionVersionRepo := store.MissionVersions()
	publishMissionUC := appmission.NewPublishUseCase(store)
	listVersionsUC := appmission.NewListVersionsUseCase(missionRepo, missionVersionRepo, memberRepo)
	getVersionUC := appmission.NewGetVersionUseCase(missionRepo, missionVersionRepo, memberRepo)

	apiKeyRepo := store.OrganizationAPIKeys()
	createAPIKeyUC := appapikey.NewCreateUseCase(apiKeyRepo, memberRepo)
	listAPIKeyUC := appapikey.NewListUseCase(apiKeyRepo, memberRepo)
	revokeAPIKeyUC := appapikey.NewRevokeUseCase(apiKeyRepo, memberRepo)
	engineHashUC := appengine.NewGetActiveHashUseCase(missionRepo)
	engineContractUC := appengine.NewGetActiveContractUseCase(missionVersionRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(registerUC, loginUC, logoutAllUC, handler.CookieConfig{
		Secure:      cfg.Production,
		ExpiryHours: cfg.JWTExpiryHours,
	})
	orgHandler := handler.NewOrganizationHandler(createOrgUC, listOrgUC, getOrgUC, updateOrgUC, deleteOrgUC)
	wsHandler := handler.NewWorkspaceHandler(createWsUC, listWsUC, getWsUC, updateWsUC, deleteWsUC)
	missionHandler := handler.NewMissionHandler(createMissionUC, listMissionUC, getMissionUC, updateMissionUC, updateMissionGraphUC, deleteMissionUC)
	missionVersionHandler := handler.NewMissionVersionHandler(publishMissionUC, listVersionsUC, getVersionUC)
	apiKeyHandler := handler.NewAPIKeyHandler(createAPIKeyUC, listAPIKeyUC, revokeAPIKeyUC)
	engineHandler := handler.NewEngineHandler(engineHashUC, engineContractUC)

	e := echo.New()
	e.HideBanner = true
	configureTimeouts(e.Server)
	e.IPExtractor = echo.ExtractIPDirect()
	e.Validator = &echoValidator{v: validator.New()}
	e.HTTPErrorHandler = newErrorHandler(logger)

	// Global middleware
	useObservability(e, logger)
	e.Use(echomiddleware.BodyLimit(maxRequestBody))
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
		AllowCredentials: true,
	}))

	authMW := httpmw.Auth(cfg.JWTSecret, userRepo)

	// Auth routes (rate-limited)
	auth := e.Group("/auth")
	auth.Use(httpmw.RateLimit(cfg.ThrottleLimit))
	auth.POST("/register", authHandler.Register, httpmw.RateLimit(5))
	auth.POST("/login", authHandler.Login, httpmw.RateLimit(5))
	auth.POST("/logout", authHandler.Logout)
	auth.POST("/logout-all", authHandler.LogoutAll, authMW)

	// Organization routes (authenticated)
	orgs := e.Group("/organizations")
	orgs.Use(authMW)
	orgs.POST("", orgHandler.Create)
	orgs.GET("", orgHandler.List)
	orgs.GET("/:id", orgHandler.Get)
	orgs.PATCH("/:id", orgHandler.Update)
	orgs.DELETE("/:id", orgHandler.Delete)

	// Workspace routes (nested under the organization tenant)
	orgs.POST("/:id/workspaces", wsHandler.Create)
	orgs.GET("/:id/workspaces", wsHandler.List)
	orgs.GET("/:id/workspaces/:workspaceID", wsHandler.Get)
	orgs.PATCH("/:id/workspaces/:workspaceID", wsHandler.Update)
	orgs.DELETE("/:id/workspaces/:workspaceID", wsHandler.Delete)

	// Mission routes (nested under the workspace)
	orgs.POST("/:id/workspaces/:workspaceID/missions", missionHandler.Create)
	orgs.GET("/:id/workspaces/:workspaceID/missions", missionHandler.List)
	orgs.GET("/:id/workspaces/:workspaceID/missions/:missionID", missionHandler.Get)
	orgs.PATCH("/:id/workspaces/:workspaceID/missions/:missionID", missionHandler.Update)
	orgs.PUT("/:id/workspaces/:workspaceID/missions/:missionID/graph", missionHandler.UpdateGraph)
	orgs.DELETE("/:id/workspaces/:workspaceID/missions/:missionID", missionHandler.Delete)

	// Mission version routes (publish + immutable snapshots)
	orgs.POST("/:id/workspaces/:workspaceID/missions/:missionID/publish", missionVersionHandler.Publish)
	orgs.GET("/:id/workspaces/:workspaceID/missions/:missionID/versions", missionVersionHandler.ListVersions)
	orgs.GET("/:id/workspaces/:workspaceID/missions/:missionID/versions/:hash", missionVersionHandler.GetVersion)

	// Organization API key routes (ADMIN; enforced in the use case)
	orgs.POST("/:id/api-keys", apiKeyHandler.Create)
	orgs.GET("/:id/api-keys", apiKeyHandler.List)
	orgs.DELETE("/:id/api-keys/:keyID", apiKeyHandler.Revoke)

	// Engine routes (UE5 plugin, x-api-key auth — outside the JWT group)
	engineThrottle := time.Duration(cfg.EngineLastUsedThrottleMs) * time.Millisecond
	engine := e.Group("/engine")
	engine.Use(httpmw.APIKeyAuth(apiKeyRepo))
	engine.Use(httpmw.RateLimitByEngineKey(cfg.EngineThrottleLimit))
	engine.Use(httpmw.TouchAPIKey(apiKeyRepo, engineThrottle))
	engine.GET("/missions/:missionID/active/hash", engineHandler.ActiveHash)
	engine.GET("/missions/:missionID/active", engineHandler.ActiveContract)

	// Health
	e.GET("/health", health)

	// API docs (Swagger UI), non-production only
	if !cfg.Production {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	return &Server{echo: e, pool: pool, port: cfg.Port}, nil
}

// @Summary  Health check
// @Tags     health
// @Produce  json
// @Success  200  {object}  map[string]string
// @Router   /health [get]
func health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func configureTimeouts(srv *http.Server) {
	srv.ReadHeaderTimeout = readHeaderTimeout
	srv.ReadTimeout = readTimeout
	srv.IdleTimeout = idleTimeout
}

func useObservability(e *echo.Echo, logger *slog.Logger) {
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogMethod:    true,
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogRequestID: true,
		LogError:     true,
		LogValuesFunc: func(c echo.Context, v echomiddleware.RequestLoggerValues) error {
			level := slog.LevelInfo
			switch {
			case v.Status >= http.StatusInternalServerError:
				level = slog.LevelError
			case v.Status >= http.StatusBadRequest:
				level = slog.LevelWarn
			}
			attrs := []slog.Attr{
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
				slog.String("ip", v.RemoteIP),
				slog.String("request_id", v.RequestID),
			}
			if v.Error != nil {
				attrs = append(attrs, slog.String("error", v.Error.Error()))
			}
			logger.LogAttrs(c.Request().Context(), level, "request", attrs...)
			return nil
		},
	}))
	e.Use(echomiddleware.Recover())

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		Subsystem:  "polina_api",
		Registerer: registry,
	}))
	e.GET("/metrics", echoprometheus.NewHandlerWithConfig(echoprometheus.HandlerConfig{
		Gatherer: registry,
	}))
}

func (s *Server) Start() error {
	if err := s.echo.Start(":" + s.port); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

func (s *Server) Close() {
	s.pool.Close()
}

type echoValidator struct{ v *validator.Validate }

func (ev *echoValidator) Validate(i any) error {
	if err := ev.v.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	return nil
}

func newErrorHandler(logger *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		he, ok := err.(*echo.HTTPError)
		if !ok {
			logger.LogAttrs(c.Request().Context(), slog.LevelError, "unhandled error",
				slog.String("error", err.Error()),
				slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
			)
			he = echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
		if !c.Response().Committed {
			_ = c.JSON(he.Code, map[string]any{
				"status_code": he.Code,
				"message":     he.Message,
			})
		}
	}
}
