package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/novarod/polina/apps/api/internal/server"
)

func main() {
	cfg := server.Config{
		DBURL:          mustEnv("DATABASE_URL"),
		JWTSecret:      mustEnv("JWT_SECRET"),
		JWTExpiryHours: envInt("JWT_EXPIRY_HOURS", 24),
		BcryptRounds:   envInt("BCRYPT_ROUNDS", 12),
		Port:           envStr("PORT", "8080"),
		FrontendURL:    envStr("FRONTEND_URL", "http://localhost:3000"),
		ThrottleLimit:  envInt("THROTTLE_LIMIT", 30),
	}

	srv, err := server.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	log.Printf("Polina API listening on :%s", cfg.Port)
	err = srv.Start()
	srv.Close()
	if err != nil {
		log.Fatalf("server: %v", err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
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
