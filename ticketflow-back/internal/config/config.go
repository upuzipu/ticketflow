// Package config loads application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds all application settings.
type Config struct {
	HTTPAddr         string
	PGDSN            string
	WSAllowedOrigins []string
	JWTSecret        []byte
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
}

// Load reads configuration from the environment and validates it.
func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:   getenv("HTTP_ADDR", ":8080"),
		PGDSN:      os.Getenv("PG_DSN"),
		JWTSecret:  []byte(os.Getenv("JWT_SECRET")),
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}
	cfg.WSAllowedOrigins = strings.Split(getenv("WS_ALLOWED_ORIGINS", "localhost:3000,localhost:3005,localhost:8080"), ",")
	if cfg.PGDSN == "" {
		return nil, errors.New("config: PG_DSN is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("config: JWT_SECRET is required and must be at least 32 bytes, got %d", len(cfg.JWTSecret))
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
