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
}

func Load() (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Info("could not load .env file, using environment variables")
	}

	cfg := &Config{
		PostgresURL: os.Getenv("POSTGRES_URL"),
		Interval:    getIntWithDefault("MONITOR_INTERVAL", 30),
	}

	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	if cfg.Interval < 1 {
		return nil, fmt.Errorf("MONITOR_INTERVAL must be at least 1 minute, got %d", cfg.Interval)
	}

	log.Info(fmt.Sprintf("configuration loaded: interval=%dm", cfg.Interval))

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
