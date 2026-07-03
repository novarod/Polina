package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/novarod/polina/apps/api/internal/server"
)

const shutdownTimeout = 10 * time.Second

// @title       Polina API
// @version     0.1.0
// @description Mission-orchestration backend for Unreal Engine 5.
// @BasePath    /
// @securityDefinitions.apikey BearerAuth
// @in          header
// @name        Authorization
// @securityDefinitions.apikey ApiKeyAuth
// @in          header
// @name        x-api-key
func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	production := os.Getenv("ENV") == "production"
	logger := newLogger(os.Stdout, production)
	slog.SetDefault(logger)

	dbURL, err := requireEnv("DATABASE_URL")
	if err != nil {
		return err
	}
	jwtSecret, err := requireEnv("JWT_SECRET")
	if err != nil {
		return err
	}

	cfg := server.Config{
		DBURL:                    dbURL,
		JWTSecret:                jwtSecret,
		JWTExpiryHours:           envInt("JWT_EXPIRY_HOURS", 24),
		BcryptRounds:             envInt("BCRYPT_ROUNDS", 12),
		Port:                     envStr("PORT", "8080"),
		FrontendURL:              envStr("FRONTEND_URL", "http://localhost:3000"),
		ThrottleLimit:            envInt("THROTTLE_LIMIT", 30),
		EngineThrottleLimit:      envInt("ENGINE_THROTTLE_LIMIT", 600),
		EngineLastUsedThrottleMs: envInt("ENGINE_LAST_USED_THROTTLE_MS", 60000),
		Production:               production,
		Logger:                   logger,
	}

	if len(cfg.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 bytes (got %d)", len(cfg.JWTSecret))
	}
	if cfg.ThrottleLimit <= 0 {
		return fmt.Errorf("THROTTLE_LIMIT must be greater than 0 (got %d)", cfg.ThrottleLimit)
	}
	if cfg.EngineThrottleLimit <= 0 {
		return fmt.Errorf("ENGINE_THROTTLE_LIMIT must be greater than 0 (got %d)", cfg.EngineThrottleLimit)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	srv, err := server.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()
	logger.Info("polina api listening", "port", cfg.Port)

	select {
	case err := <-errCh:
		srv.Close()
		if err != nil {
			return fmt.Errorf("server: %w", err)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received, draining requests")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown", "error", err)
		}
		<-errCh
		srv.Close()
		logger.Info("shutdown complete")
	}
	return nil
}

func newLogger(w io.Writer, production bool) *slog.Logger {
	if production {
		return slog.New(slog.NewJSONHandler(w, nil))
	}
	return slog.New(slog.NewTextHandler(w, nil))
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("missing required env var: %s", key)
	}
	return v, nil
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
