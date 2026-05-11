package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/monitor/internal/log"
)

// Config holds validated configuration for the monitor service
type Config struct {
	PostgresURL string
}

// Load loads environment variables and validates them
func Load() (*Config, error) {
	// Load .env file (non-fatal if missing)
	if err := godotenv.Load(".env"); err != nil {
		log.Info("Could not load .env file, using environment variables")
	}

	cfg := &Config{
		PostgresURL: os.Getenv("POSTGRES_URL"),
	}

	// Validate required fields
	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	log.Info("Configuration loaded successfully")

	return cfg, nil
}
