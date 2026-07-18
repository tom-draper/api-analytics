package main

import (
	"errors"
	"net"
	"testing"
	"time"
)

func sampleResults() []pingResult {
	return []pingResult{
		{
			monitor: MonitorRow{APIKey: "key-1", URL: "https://up.example"},
			status:  200,
			elapsed: 1500 * time.Millisecond,
		},
		{
			monitor: MonitorRow{APIKey: "key-1", URL: "https://down.example"},
			err:     errors.New("dial timeout"),
		},
	}
}

func TestSuccessfulPingsDropsUnreachable(t *testing.T) {
	pings := successfulPings(sampleResults())

	if len(pings) != 1 {
		t.Fatalf("expected 1 stored ping (the reachable one), got %d", len(pings))
	}
	p := pings[0]
	if p.URL != "https://up.example" || p.APIKey != "key-1" {
		t.Errorf("wrong ping recorded: %+v", p)
	}
	if p.Status != 200 {
		t.Errorf("Status = %d, want 200", p.Status)
	}
	if p.ResponseTime != 1500 {
		t.Errorf("ResponseTime = %d ms, want 1500 (from elapsed)", p.ResponseTime)
	}
}

func TestAlertResultsIncludeEveryOutcome(t *testing.T) {
	results := alertResults(sampleResults())

	if len(results) != 2 {
		t.Fatalf("expected every URL to be reported for alerting, got %d", len(results))
	}

	// Both reachable and unreachable URLs must be reported, and the unreachable
	// one must carry its error so the alerter can classify it.
	errByURL := make(map[string]error)
	seen := make(map[string]bool)
	for _, r := range results {
		errByURL[r.URL] = r.Err
		seen[r.URL] = true
	}

	if !seen["https://up.example"] || !seen["https://down.example"] {
		t.Fatalf("missing a URL in alert results: %v", seen)
	}
	if errByURL["https://up.example"] != nil {
		t.Errorf("reachable URL should have no error, got %v", errByURL["https://up.example"])
	}
	if errByURL["https://down.example"] == nil {
		t.Error("unreachable URL should carry a non-nil error")
	}
}

func TestGetMethod(t *testing.T) {
	if got := getMethod(true); got != "HEAD" {
		t.Errorf("getMethod(true) = %q, want HEAD", got)
	}
	if got := getMethod(false); got != "GET" {
		t.Errorf("getMethod(false) = %q, want GET", got)
	}
}

func TestIsPublicIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1",          // loopback
		"::1",                // loopback v6
		"169.254.169.254",    // cloud metadata (link-local)
		"10.0.0.5",           // private RFC1918
		"172.16.3.4",         // private RFC1918
		"192.168.1.1",        // private RFC1918
		"100.64.0.1",         // carrier-grade NAT
		"0.0.0.0",            // unspecified
		"fd00::1",            // IPv6 unique local
		"fe80::1",            // IPv6 link-local
		"224.0.0.1",          // multicast
		"::ffff:127.0.0.1",   // IPv4-mapped loopback
		"::ffff:192.168.0.1", // IPv4-mapped private
	}
	for _, s := range blocked {
		if ip := net.ParseIP(s); ip == nil || isPublicIP(ip) {
			t.Errorf("isPublicIP(%s) = true, want false (should be blocked)", s)
		}
	}

	public := []string{
		"1.1.1.1",
		"8.8.8.8",
		"93.184.216.34", // example.com
		"2606:4700:4700::1111",
	}
	for _, s := range public {
		if ip := net.ParseIP(s); ip == nil || !isPublicIP(ip) {
			t.Errorf("isPublicIP(%s) = false, want true (should be allowed)", s)
		}
	}
}

func TestGuardPrivateAddressBlocksMetadataEndpoint(t *testing.T) {
	err := guardPrivateAddress("tcp", "169.254.169.254:80", nil)
	if !errors.Is(err, errPrivateAddress) {
		t.Errorf("guardPrivateAddress(metadata) error = %v, want errPrivateAddress", err)
	}

	if err := guardPrivateAddress("tcp", "8.8.8.8:443", nil); err != nil {
		t.Errorf("guardPrivateAddress(public) error = %v, want nil", err)
	}
}

func TestApplyScheme(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		secure bool
		want   string
	}{
		{"keeps existing https", "https://example.com", false, "https://example.com"},
		{"keeps existing http", "http://example.com", true, "http://example.com"},
		{"adds https when secure", "example.com/health", true, "https://example.com/health"},
		{"adds http when not secure", "example.com/health", false, "http://example.com/health"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applyScheme(tt.url, tt.secure); got != tt.want {
				t.Errorf("applyScheme(%q, %v) = %q, want %q", tt.url, tt.secure, got, tt.want)
			}
		})
	}
}
