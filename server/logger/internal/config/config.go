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
		MaxInsert:   getIntWithDefault("MAX_INSERT", 2000),
	}

	// Validate required fields
	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	// Validate ranges
	if cfg.MaxInsert < 1 || cfg.MaxInsert > 10000 {
		return nil, fmt.Errorf("MAX_INSERT must be between 1 and 10000, got %d", cfg.MaxInsert)
	}

	log.LogToFile(fmt.Sprintf("Configuration loaded: MaxInsert=%d", cfg.MaxInsert))

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
