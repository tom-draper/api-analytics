package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/database"
)

// Config holds all configuration for tools
type Config struct {
	// Database
	PostgresURL string

	// API configuration
	APIBaseURL    string
	MonitorAPIKey string
	MonitorUserID string

	// Email configuration
	EmailAddress string
}

// Load loads configuration from environment variables
// Attempts to load from .env file first
func Load() (*Config, error) {
	// Try to load .env file (not fatal if it doesn't exist)
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: could not load .env file, using system environment variables")
	}

	cfg := &Config{
		PostgresURL:   os.Getenv("POSTGRES_URL"),
		APIBaseURL:    os.Getenv("API_BASE_URL"),
		MonitorAPIKey: os.Getenv("MONITOR_API_KEY"),
		MonitorUserID: os.Getenv("MONITOR_USER_ID"),
		EmailAddress:  os.Getenv("EMAIL_ADDRESS"),
	}

	return cfg, nil
}

// LoadWithRequired loads configuration and validates required fields
func LoadWithRequired(required ...string) (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// Validate required fields
	for _, field := range required {
		if err := cfg.validateField(field); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// validateField checks if a required field is set
func (c *Config) validateField(field string) error {
	switch field {
	case "POSTGRES_URL":
		if c.PostgresURL == "" {
			return fmt.Errorf("POSTGRES_URL environment variable is required")
		}
	case "API_BASE_URL":
		if c.APIBaseURL == "" {
			return fmt.Errorf("API_BASE_URL environment variable is required")
		}
	case "MONITOR_API_KEY":
		if c.MonitorAPIKey == "" {
			return fmt.Errorf("MONITOR_API_KEY environment variable is required")
		}
	case "MONITOR_USER_ID":
		if c.MonitorUserID == "" {
			return fmt.Errorf("MONITOR_USER_ID environment variable is required")
		}
	case "EMAIL_ADDRESS":
		if c.EmailAddress == "" {
			return fmt.Errorf("EMAIL_ADDRESS environment variable is required")
		}
	default:
		return fmt.Errorf("unknown required field: %s", field)
	}
	return nil
}

// NewDatabase creates a new database connection from config
func (c *Config) NewDatabase(ctx context.Context) (*database.DB, error) {
	if c.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is not configured")
	}
	return database.New(ctx, c.PostgresURL)
}

// HasAPIConfig checks if API configuration is available
func (c *Config) HasAPIConfig() bool {
	return c.APIBaseURL != "" && c.MonitorAPIKey != "" && c.MonitorUserID != ""
}

// HasEmailConfig checks if email configuration is available
func (c *Config) HasEmailConfig() bool {
	return c.EmailAddress != ""
}
