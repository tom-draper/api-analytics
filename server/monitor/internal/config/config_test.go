package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("POSTGRES_URL", "postgres://localhost/test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Interval != 30 {
		t.Errorf("Interval = %d, want default 30", cfg.Interval)
	}
	if cfg.AlertsEnabled {
		t.Error("alerts should be off by default")
	}
}

func TestLoadRequiresPostgresURL(t *testing.T) {
	t.Setenv("POSTGRES_URL", "")
	if _, err := Load(); err == nil {
		t.Error("expected an error when POSTGRES_URL is unset")
	}
}

func TestLoadRejectsBadInterval(t *testing.T) {
	t.Setenv("POSTGRES_URL", "postgres://localhost/test")
	t.Setenv("MONITOR_INTERVAL", "0")
	if _, err := Load(); err == nil {
		t.Error("expected an error for an interval below 1 minute")
	}
}

func TestLoadAlertsRequireRecipient(t *testing.T) {
	t.Setenv("POSTGRES_URL", "postgres://localhost/test")
	t.Setenv("MONITOR_ALERTS_ENABLED", "true")
	t.Setenv("MONITOR_ALERT_EMAIL", "")
	if _, err := Load(); err == nil {
		t.Error("enabling alerts without MONITOR_ALERT_EMAIL should be an error")
	}
}

func TestLoadAlertsEnabledWithRecipient(t *testing.T) {
	t.Setenv("POSTGRES_URL", "postgres://localhost/test")
	t.Setenv("MONITOR_ALERTS_ENABLED", "true")
	t.Setenv("MONITOR_ALERT_EMAIL", "ops@example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.AlertsEnabled || cfg.AlertEmail != "ops@example.com" {
		t.Errorf("alert config not applied: %+v", cfg)
	}
}

func TestLoadInvalidBoolFallsBackToDisabled(t *testing.T) {
	// A malformed MONITOR_ALERTS_ENABLED falls back to the default (false), so
	// the missing-recipient rule does not trip.
	t.Setenv("POSTGRES_URL", "postgres://localhost/test")
	t.Setenv("MONITOR_ALERTS_ENABLED", "not-a-bool")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AlertsEnabled {
		t.Error("a malformed boolean should leave alerts disabled")
	}
}
