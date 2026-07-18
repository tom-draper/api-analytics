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

func TestValidMonitorURLBoundaries(t *testing.T) {
	base := "http://x/"
	at255 := base + strings.Repeat("a", 255-len(base))
	if len(at255) != 255 {
		t.Fatalf("test setup wrong: len=%d", len(at255))
	}
	if !validMonitorURL(at255) {
		t.Error("a 255-char URL should be valid (matches the column width)")
	}
	if validMonitorURL(at255 + "a") {
		t.Error("a 256-char URL should be rejected")
	}

	// Every C0 control character, DEL, and the C1 NEL must be rejected, since a
	// CR/LF could inject headers into the alert email that embeds the URL.
	for _, r := range []rune{'\x00', '\x07', '\x08', '\x0b', '\x0c', '\x1b', '\x7f', '\u0085'} {
		if validMonitorURL("http://x/" + string(r)) {
			t.Errorf("URL containing control char %U should be rejected", r)
		}
	}

	// A legitimate internationalized URL with a multibyte path is accepted.
	if !validMonitorURL("https://例え.テスト/パス") {
		t.Error("a URL with a multibyte path should be accepted")
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
