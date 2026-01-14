package config

import (
	"fmt"
	"os"
	"strconv"

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
}

// Load loads environment variables and validates them
func Load() (*Config, error) {
	// Load .env file (non-fatal if missing)
	_ = godotenv.Load()

	cfg := &Config{
		PostgresURL: os.Getenv("POSTGRES_URL"),
		Port:        getIntWithDefault("PORT", 3000),
		RateLimit:   uint(getIntWithDefault("API_RATE_LIMIT", 100)),
		MaxLoad:     getIntWithFallback("API_MAX_LOAD", "MAX_LOAD", 1_000_000),
		PageSize:    getIntWithFallback("API_PAGE_SIZE", "PAGE_SIZE", 250_000),
	}

	// Validate required fields
	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	// Validate ranges
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("PORT must be between 1 and 65535, got %d", cfg.Port)
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

	log.Info(fmt.Sprintf("Configuration loaded: Port=%d, RateLimit=%d, PageSize=%d, MaxLoad=%d",
		cfg.Port, cfg.RateLimit, cfg.PageSize, cfg.MaxLoad))

	return cfg, nil
}

// getIntWithDefault is a helper that doesn't log (used internally)
func getIntWithDefault(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Info(fmt.Sprintf("Invalid integer for %s, using default %d", name, defaultValue))
		return defaultValue
	}

	return value
}

// getIntWithFallback tries the primary name first, then falls back to legacy name for backwards compatibility
func getIntWithFallback(primaryName, fallbackName string, defaultValue int) int {
	// Try primary name first
	valueStr := os.Getenv(primaryName)
	if valueStr != "" {
		value, err := strconv.Atoi(valueStr)
		if err == nil {
			return value
		}
		log.Info(fmt.Sprintf("Invalid integer for %s, trying fallback", primaryName))
	}

	// Try fallback name for backwards compatibility
	valueStr = os.Getenv(fallbackName)
	if valueStr != "" {
		value, err := strconv.Atoi(valueStr)
		if err == nil {
			return value
		}
		log.Info(fmt.Sprintf("Invalid integer for %s, using default", fallbackName))
	}

	// Use default
	return defaultValue
}
