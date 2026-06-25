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

	"github.com/novarod/polina/apps/api/internal/adapters/http/handler"
	httpmw "github.com/novarod/polina/apps/api/internal/adapters/http/middleware"
	"github.com/novarod/polina/apps/api/internal/adapters/postgres/repository"
	appauth "github.com/novarod/polina/apps/api/internal/application/auth"
)

type Config struct {
	DBURL          string
	JWTSecret      string
	JWTExpiryHours int
	BcryptRounds   int
	Port           string
	FrontendURL    string
	ThrottleLimit  int
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

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	memberRepo := repository.NewMemberRepository(pool)

	// Use cases
	registerUC := appauth.NewRegisterUseCase(userRepo, cfg.BcryptRounds)
	loginUC := appauth.NewLoginUseCase(userRepo, memberRepo, cfg.JWTSecret, cfg.JWTExpiryHours)

	// Handlers
	authHandler := handler.NewAuthHandler(registerUC, loginUC)

	e := echo.New()
	e.HideBanner = true
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

	// Health
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	return &Server{echo: e, pool: pool, port: cfg.Port}, nil
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
