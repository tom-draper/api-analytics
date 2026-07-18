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
		os.Setenv("API_PORT", "70000")
		defer os.Unsetenv("POSTGRES_URL")
		defer os.Unsetenv("API_PORT")

		_, err := Load()

		if err == nil {
			t.Error("Expected error for invalid port")
		}
	})
}

func TestLoadValidatesBounds(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{"rate limit zero", map[string]string{"API_RATE_LIMIT": "0"}},
		{"rate limit too high", map[string]string{"API_RATE_LIMIT": "20000"}},
		{"page size too small", map[string]string{"API_PAGE_SIZE": "500"}},
		{"page size too large", map[string]string{"API_PAGE_SIZE": "2000000"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("POSTGRES_URL", "postgresql://localhost/test")
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			if _, err := Load(); err == nil {
				t.Errorf("expected a validation error for %s", tt.name)
			}
		})
	}
}

func TestLoadMaxLoadMustNotBeBelowPageSize(t *testing.T) {
	t.Setenv("POSTGRES_URL", "postgresql://localhost/test")
	t.Setenv("API_PAGE_SIZE", "250000")
	t.Setenv("API_MAX_LOAD", "100000") // below page size

	if _, err := Load(); err == nil {
		t.Error("expected an error when API_MAX_LOAD is below API_PAGE_SIZE")
	}
}

func TestLoadAcceptsValidCustomValues(t *testing.T) {
	t.Setenv("POSTGRES_URL", "postgresql://localhost/test")
	t.Setenv("API_RATE_LIMIT", "50")
	t.Setenv("API_PAGE_SIZE", "100000")
	t.Setenv("API_MAX_LOAD", "500000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RateLimit != 50 || cfg.PageSize != 100000 || cfg.MaxLoad != 500000 {
		t.Errorf("custom values not applied: %+v", cfg)
	}
}
