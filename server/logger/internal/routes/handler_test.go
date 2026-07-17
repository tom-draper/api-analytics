package routes

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/tom-draper/api-analytics/server/database"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		n        int
		expected string
	}{
		{
			name:     "shorter than limit is unchanged",
			value:    "/api/v1/users",
			n:        255,
			expected: "/api/v1/users",
		},
		{
			name:     "exactly at limit is unchanged",
			value:    strings.Repeat("a", 255),
			n:        255,
			expected: strings.Repeat("a", 255),
		},
		{
			name:     "ascii truncated at limit",
			value:    strings.Repeat("a", 300),
			n:        255,
			expected: strings.Repeat("a", 255),
		},
		{
			name:     "multi-byte rune not split at boundary",
			value:    strings.Repeat("a", 254) + "é",
			n:        255,
			expected: strings.Repeat("a", 254),
		},
		{
			name:     "value of only multi-byte runes",
			value:    strings.Repeat("日", 10),
			n:        4,
			expected: "日",
		},
		{
			name:     "empty value",
			value:    "",
			n:        255,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.value, tt.n)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, expected %q", tt.value, tt.n, result, tt.expected)
			}
			if len(result) > tt.n {
				t.Errorf("truncate(%q, %d) returned %d bytes, over the limit", tt.value, tt.n, len(result))
			}
			if !utf8.ValidString(result) {
				t.Errorf("truncate(%q, %d) = %q, which is not valid UTF-8", tt.value, tt.n, result)
			}
		})
	}
}

// Truncation runs before validation at ingest, so a value cut mid-rune would be
// rejected as invalid UTF-8 and the request silently dropped.
func TestTruncatedValuesStayValid(t *testing.T) {
	values := []string{
		strings.Repeat("a", 254) + "é",
		strings.Repeat("日", 200),
		"/api/" + strings.Repeat("café/", 100),
		strings.Repeat("Mozilla/5.0 (Windows NT 10.0; Win64; x64) ", 20),
	}

	for _, value := range values {
		truncated := truncate(value, 255)
		if !database.ValidString(truncated) {
			t.Errorf("ValidString(truncate(%.30q..., 255)) = false, expected a truncated value to stay valid", value)
		}
	}
}

func TestRealUserAgentsAccepted(t *testing.T) {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0",
		"curl/8.4.0",
	}

	for _, userAgent := range userAgents {
		if !database.ValidUserAgent(truncate(userAgent, 255)) {
			t.Errorf("ValidUserAgent(%q) = false, expected real browser traffic to be logged", userAgent)
		}
	}
}
