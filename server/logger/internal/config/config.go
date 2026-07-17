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
	Port        int
	RateLimit   int
	IPRateLimit int
	MaxInsert   int
}

// Load loads environment variables and validates them
func Load() (*Config, error) {
	// Load .env file (non-fatal if missing)
	if err := godotenv.Load(".env"); err != nil {
		log.Info("could not load .env file, using environment variables")
	}

	cfg := &Config{
		PostgresURL: os.Getenv("POSTGRES_URL"),
		Port:        getIntWithDefault("LOGGER_PORT", 8000),
		RateLimit:   getIntWithDefault("LOGGER_RATE_LIMIT", 10),
		IPRateLimit: getIntWithDefault("LOGGER_IP_RATE_LIMIT", 100),
		MaxInsert:   getIntWithDefault("LOGGER_MAX_INSERT", 2000),
	}

	// Validate required fields
	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	// Validate ranges
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("LOGGER_PORT must be between 1 and 65535, got %d", cfg.Port)
	}

	// Validate ranges
	if cfg.RateLimit < 1 || cfg.RateLimit > 1000 {
		return nil, fmt.Errorf("LOGGER_RATE_LIMIT must be between 1 and 1000, got %d", cfg.RateLimit)
	}

	if cfg.IPRateLimit < 1 || cfg.IPRateLimit > 100000 {
		return nil, fmt.Errorf("LOGGER_IP_RATE_LIMIT must be between 1 and 100000, got %d", cfg.IPRateLimit)
	}

	if cfg.MaxInsert < 1 || cfg.MaxInsert > 10000 {
		return nil, fmt.Errorf("LOGGER_MAX_INSERT must be between 1 and 10000, got %d", cfg.MaxInsert)
	}

	log.Info(fmt.Sprintf("configuration loaded: port=%d, rate_limit=%d, ip_rate_limit=%d, max_insert=%d", cfg.Port, cfg.RateLimit, cfg.IPRateLimit, cfg.MaxInsert))

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
		log.Info(fmt.Sprintf("invalid integer for %s, using default %d", name, defaultValue))
		return defaultValue
	}

	return value
}
