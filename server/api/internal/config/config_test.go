package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Run("it returns error when POSTGRES_URL is missing", func(t *testing.T) {
		os.Unsetenv("POSTGRES_URL")

		_, err := Load()

		if err == nil {
			t.Error("Expected error when POSTGRES_URL is missing")
		}
	})

	t.Run("it returns valid config with defaults", func(t *testing.T) {
		os.Setenv("POSTGRES_URL", "postgresql://localhost/test")
		defer os.Unsetenv("POSTGRES_URL")

		cfg, err := Load()

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if cfg.Port != 3000 {
			t.Errorf("Expected default port 3000, got %d", cfg.Port)
		}

		if cfg.RateLimit != 100 {
			t.Errorf("Expected default rate limit 100, got %d", cfg.RateLimit)
		}
	})

	t.Run("it validates port range", func(t *testing.T) {
		os.Setenv("POSTGRES_URL", "postgresql://localhost/test")
		os.Setenv("PORT", "70000")
		defer os.Unsetenv("POSTGRES_URL")
		defer os.Unsetenv("PORT")

		_, err := Load()

		if err == nil {
			t.Error("Expected error for invalid port")
		}
	})
}
