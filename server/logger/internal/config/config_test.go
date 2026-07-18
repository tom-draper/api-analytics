package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("POSTGRES_URL", "postgres://localhost/test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 8000 {
		t.Errorf("Port = %d, want default 8000", cfg.Port)
	}
	if cfg.RateLimit != 10 {
		t.Errorf("RateLimit = %d, want default 10", cfg.RateLimit)
	}
	if cfg.IPRateLimit != 100 {
		t.Errorf("IPRateLimit = %d, want default 100", cfg.IPRateLimit)
	}
	if cfg.MaxInsert != 2000 {
		t.Errorf("MaxInsert = %d, want default 2000", cfg.MaxInsert)
	}
}

func TestLoadRequiresPostgresURL(t *testing.T) {
	t.Setenv("POSTGRES_URL", "")
	if _, err := Load(); err == nil {
		t.Error("expected an error when POSTGRES_URL is unset")
	}
}

func TestLoadValidatesRanges(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{"port too high", map[string]string{"LOGGER_PORT": "70000"}},
		{"rate limit zero", map[string]string{"LOGGER_RATE_LIMIT": "0"}},
		{"rate limit too high", map[string]string{"LOGGER_RATE_LIMIT": "2000"}},
		{"ip rate limit zero", map[string]string{"LOGGER_IP_RATE_LIMIT": "0"}},
		{"ip rate limit too high", map[string]string{"LOGGER_IP_RATE_LIMIT": "200000"}},
		{"max insert zero", map[string]string{"LOGGER_MAX_INSERT": "0"}},
		{"max insert too high", map[string]string{"LOGGER_MAX_INSERT": "20000"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("POSTGRES_URL", "postgres://localhost/test")
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			if _, err := Load(); err == nil {
				t.Errorf("expected a validation error for %s", tt.name)
			}
		})
	}
}

func TestLoadAcceptsValidCustomValues(t *testing.T) {
	t.Setenv("POSTGRES_URL", "postgres://localhost/test")
	t.Setenv("LOGGER_PORT", "9000")
	t.Setenv("LOGGER_RATE_LIMIT", "50")
	t.Setenv("LOGGER_IP_RATE_LIMIT", "500")
	t.Setenv("LOGGER_MAX_INSERT", "5000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != 9000 || cfg.RateLimit != 50 || cfg.IPRateLimit != 500 || cfg.MaxInsert != 5000 {
		t.Errorf("custom values not applied: %+v", cfg)
	}
}
