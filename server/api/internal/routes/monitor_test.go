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
