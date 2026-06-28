package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/novarod/polina/apps/api/docs"
	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	"github.com/novarod/polina/apps/api/internal/adapters/postgres"
	appauth "github.com/novarod/polina/apps/api/internal/application/auth"
	apporg "github.com/novarod/polina/apps/api/internal/application/organization"
)

type Config struct {
	DBURL          string
	JWTSecret      string
	JWTExpiryHours int
	BcryptRounds   int
	Port           string
	FrontendURL    string
	ThrottleLimit  int
	Production     bool
}

type Server struct {
	echo *echo.Echo
	pool *pgxpool.Pool
	port string
}

func New(ctx context.Context, cfg Config) (*Server, error) {
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
	orgRepo := store.Organizations()
	memberRepo := store.Members()

	// Use cases
	registerUC := appauth.NewRegisterUseCase(store.Users(), cfg.BcryptRounds)
	loginUC := appauth.NewLoginUseCase(store.Users(), memberRepo, cfg.JWTSecret, cfg.JWTExpiryHours)

	createOrgUC := apporg.NewCreateUseCase(store)
	listOrgUC := apporg.NewListUseCase(orgRepo)
	getOrgUC := apporg.NewGetUseCase(orgRepo, memberRepo)
	updateOrgUC := apporg.NewUpdateUseCase(orgRepo, memberRepo)
	deleteOrgUC := apporg.NewDeleteUseCase(store)

	// Handlers
	authHandler := handler.NewAuthHandler(registerUC, loginUC, handler.CookieConfig{
		Secure:      cfg.Production,
		ExpiryHours: cfg.JWTExpiryHours,
	})
	orgHandler := handler.NewOrganizationHandler(createOrgUC, listOrgUC, getOrgUC, updateOrgUC, deleteOrgUC)

	e := echo.New()
	e.HideBanner = true
	e.IPExtractor = echo.ExtractIPDirect()
	e.Validator = &echoValidator{v: validator.New()}
	e.HTTPErrorHandler = errorHandler

	// Global middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
		AllowCredentials: true,
	}))

	// Auth routes (rate-limited)
	auth := e.Group("/auth")
	auth.Use(httpmw.RateLimit(cfg.ThrottleLimit))
	auth.POST("/register", authHandler.Register, httpmw.RateLimit(5))
	auth.POST("/login", authHandler.Login, httpmw.RateLimit(5))
	auth.POST("/logout", authHandler.Logout)

	// Organization routes (authenticated)
	orgs := e.Group("/organizations")
	orgs.Use(httpmw.Auth(cfg.JWTSecret))
	orgs.POST("", orgHandler.Create)
	orgs.GET("", orgHandler.List)
	orgs.GET("/:id", orgHandler.Get)
	orgs.PATCH("/:id", orgHandler.Update)
	orgs.DELETE("/:id", orgHandler.Delete)

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

func (s *Server) Start() error {
	if err := s.echo.Start(":" + s.port); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
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

func errorHandler(err error, c echo.Context) {
	he, ok := err.(*echo.HTTPError)
	if !ok {
		he = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !c.Response().Committed {
		_ = c.JSON(he.Code, map[string]any{
			"status_code": he.Code,
			"message":     he.Message,
		})
	}
}
