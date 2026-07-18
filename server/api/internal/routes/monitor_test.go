package routes

import (
	"strings"
	"testing"
)

func TestValidMonitorURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{"plain http url", "http://example.com/health", true},
		{"https url with path and query", "https://api.example.com/status?check=1", true},
		{"empty", "", false},
		{"too long", "https://example.com/" + strings.Repeat("a", 300), false},
		{"carriage return + newline (header injection)", "http://x\r\nBcc: attacker@evil.com", false},
		{"bare newline", "http://x\nSubject: spoof", false},
		{"tab", "http://x\ty", false},
		{"null byte", "http://x\x00y", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validMonitorURL(tt.url); got != tt.valid {
				t.Errorf("validMonitorURL(%q) = %v, want %v", tt.url, got, tt.valid)
			}
		})
	}
}

func TestInternalMonitorTarget(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		internal bool
	}{
		{"public host", "https://example.com/health", false},
		{"public ip", "http://8.8.8.8/", false},
		{"scheme-less public host", "example.com/health", false},
		{"localhost", "http://localhost:8080/", true},
		{"loopback ip", "http://127.0.0.1/", true},
		{"private ip", "http://10.0.0.5/health", true},
		{"private ip 192.168", "https://192.168.1.1/", true},
		{"cloud metadata endpoint", "http://169.254.169.254/latest/meta-data/", true},
		{"scheme-less loopback", "127.0.0.1:9000", true},
		{"ipv6 loopback", "http://[::1]/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := internalMonitorTarget(tt.url); got != tt.internal {
				t.Errorf("internalMonitorTarget(%q) = %v, want %v", tt.url, got, tt.internal)
			}
		})
	}
}
