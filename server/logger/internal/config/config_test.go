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

func TestParseTrustedProxies(t *testing.T) {
	t.Run("empty falls back to the safe private/loopback default", func(t *testing.T) {
		got, err := parseTrustedProxies("")
		if err != nil {
			t.Fatalf("parseTrustedProxies: %v", err)
		}
		if len(got) != len(defaultTrustedProxies) {
			t.Errorf("got %v, want the default set", got)
		}
	})

	t.Run("valid CIDRs and IPs are kept, blanks trimmed", func(t *testing.T) {
		got, err := parseTrustedProxies(" 10.1.2.0/24 , 203.0.113.7 ,")
		if err != nil {
			t.Fatalf("parseTrustedProxies: %v", err)
		}
		want := []string{"10.1.2.0/24", "203.0.113.7"}
		if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("an invalid entry fails startup", func(t *testing.T) {
		if _, err := parseTrustedProxies("not-an-ip"); err == nil {
			t.Error("expected an error for an unparseable proxy entry")
		}
	})
}
