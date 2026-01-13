package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
)

// Config holds validated configuration for the logger service
type Config struct {
	PostgresURL string
	RateLimit   int
	MaxInsert   int
}

// LoadAndValidate loads environment variables and validates them
func LoadAndValidate() (*Config, error) {
	// Load .env file (non-fatal if missing)
	if err := godotenv.Load(".env"); err != nil {
		log.LogToFile("Warning: could not load .env file")
	}

	cfg := &Config{
		PostgresURL: os.Getenv("POSTGRES_URL"),
		RateLimit:   getIntWithDefault("LOGGER_RATE_LIMIT", 10),
		MaxInsert:   getIntWithFallback("LOGGER_MAX_INSERT", "MAX_INSERT", 2000),
	}

	// Validate required fields
	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	// Validate ranges
	if cfg.RateLimit < 1 || cfg.RateLimit > 1000 {
		return nil, fmt.Errorf("LOGGER_RATE_LIMIT must be between 1 and 1000, got %d", cfg.RateLimit)
	}

	if cfg.MaxInsert < 1 || cfg.MaxInsert > 10000 {
		return nil, fmt.Errorf("LOGGER_MAX_INSERT must be between 1 and 10000, got %d", cfg.MaxInsert)
	}

	log.LogToFile(fmt.Sprintf("Configuration loaded: RateLimit=%d, MaxInsert=%d", cfg.RateLimit, cfg.MaxInsert))

	return cfg, nil
}

// getIntWithDefault is a helper for parsing integer environment variables
func getIntWithDefault(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.LogToFile(fmt.Sprintf("Invalid integer for %s, using default %d", name, defaultValue))
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
		log.LogToFile(fmt.Sprintf("Invalid integer for %s, trying fallback", primaryName))
	}

	// Try fallback name for backwards compatibility
	valueStr = os.Getenv(fallbackName)
	if valueStr != "" {
		value, err := strconv.Atoi(valueStr)
		if err == nil {
			return value
		}
		log.LogToFile(fmt.Sprintf("Invalid integer for %s, using default", fallbackName))
	}

	// Use default
	return defaultValue
}
