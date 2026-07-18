package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/api/internal/log"
)

// Config holds validated configuration for the API service
type Config struct {
	PostgresURL string
	Port        int
	RateLimit   uint
	MaxLoad     int
	PageSize    int
	// AllowedOrigins restricts CORS to these origins. Empty means allow all
	// origins (the historical default), which is safe here because responses
	// carry no cookies and access is gated by an API key or an unguessable
	// user ID, not by the browser's ambient credentials.
	AllowedOrigins []string
}

// Load loads environment variables and validates them
func Load() (*Config, error) {
	// Load .env file (non-fatal if missing)
	_ = godotenv.Load()

	cfg := &Config{
		PostgresURL:    os.Getenv("POSTGRES_URL"),
		Port:           getIntWithDefault("API_PORT", 3000),
		RateLimit:      uint(getIntWithDefault("API_RATE_LIMIT", 100)),
		MaxLoad:        getIntWithDefault("API_MAX_LOAD", 1_000_000),
		PageSize:       getIntWithDefault("API_PAGE_SIZE", 250_000),
		AllowedOrigins: parseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}

	// Validate required fields
	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	// Validate ranges
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("API_PORT must be between 1 and 65535, got %d", cfg.Port)
	}

	if cfg.RateLimit < 1 || cfg.RateLimit > 10000 {
		return nil, fmt.Errorf("API_RATE_LIMIT must be between 1 and 10000, got %d", cfg.RateLimit)
	}

	if cfg.PageSize < 1000 || cfg.PageSize > 1_000_000 {
		return nil, fmt.Errorf("API_PAGE_SIZE must be between 1000 and 1000000, got %d", cfg.PageSize)
	}

	if cfg.MaxLoad < cfg.PageSize {
		return nil, fmt.Errorf("API_MAX_LOAD (%d) must be >= API_PAGE_SIZE (%d)", cfg.MaxLoad, cfg.PageSize)
	}

	log.Info(fmt.Sprintf("configuration loaded: port=%d, rate_limit=%d, page_size=%d, max_load=%d",
		cfg.Port, cfg.RateLimit, cfg.PageSize, cfg.MaxLoad))

	return cfg, nil
}

// parseOrigins splits a comma-separated CORS origin list, trimming blanks.
func parseOrigins(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var origins []string
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

// getIntWithDefault is a helper that doesn't log (used internally)
func getIntWithDefault(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Info(fmt.Sprintf("invalid integer for %s, using default %d", name, defaultValue))
		return defaultValue
	}

	return value
}
