package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/monitor/internal/log"
)

type Config struct {
	PostgresURL string
	Interval    int // minutes between monitor cycles

	// Email alerts are off by default. When AlertsEnabled is true and an email
	// provider (EMAIL_PROVIDER) and AlertEmail recipient are configured, the
	// monitor emails on downtime transitions.
	AlertsEnabled bool
	AlertEmail    string
}

func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Info("could not load .env file, using environment variables")
	}

	cfg := &Config{
		PostgresURL:   os.Getenv("POSTGRES_URL"),
		Interval:      getIntWithDefault("MONITOR_INTERVAL", 30),
		AlertsEnabled: getBoolWithDefault("MONITOR_ALERTS_ENABLED", false),
		AlertEmail:    os.Getenv("MONITOR_ALERT_EMAIL"),
	}

	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	if cfg.Interval < 1 {
		return nil, fmt.Errorf("MONITOR_INTERVAL must be at least 1 minute, got %d", cfg.Interval)
	}

	if cfg.AlertsEnabled && cfg.AlertEmail == "" {
		return nil, fmt.Errorf("MONITOR_ALERTS_ENABLED is set but MONITOR_ALERT_EMAIL is empty")
	}

	log.Info(fmt.Sprintf("configuration loaded: interval=%dm, alerts_enabled=%t", cfg.Interval, cfg.AlertsEnabled))

	return cfg, nil
}

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

func getBoolWithDefault(name string, defaultValue bool) bool {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Info(fmt.Sprintf("invalid boolean for %s, using default %t", name, defaultValue))
		return defaultValue
	}
	return value
}
