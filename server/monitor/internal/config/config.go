package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds validated configuration for the monitor service
type Config struct {
	PostgresURL string
}

// Load loads environment variables and validates them
func Load() (*Config, error) {
	// Load .env file (non-fatal if missing)
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Warning: could not load .env file")
	}

	cfg := &Config{
		PostgresURL: os.Getenv("POSTGRES_URL"),
	}

	// Validate required fields
	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	fmt.Println("Configuration loaded successfully")

	return cfg, nil
}
