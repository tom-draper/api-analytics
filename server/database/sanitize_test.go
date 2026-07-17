package database

import (
	"strings"
	"testing"
	"time"
)

func TestValidDate(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected bool
	}{
		{
			name:     "zero time",
			date:     time.Time{},
			expected: false,
		},
		{
			name:     "current time",
			date:     time.Now(),
			expected: true,
		},
		{
			name:     "valid past date",
			date:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "valid future date",
			date:     time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "too far in past (101 years ago)",
			date:     time.Now().AddDate(-101, 0, 0),
			expected: false,
		},
		{
			name:     "too far in future (11 years from now)",
			date:     time.Now().AddDate(11, 0, 0),
			expected: false,
		},
		{
			name:     "edge case - exactly 100 years ago",
			date:     time.Now().AddDate(-100, 0, 1), // 1 day after 100 years ago
			expected: true,
		},
		{
			name:     "edge case - exactly 10 years from now",
			date:     time.Now().AddDate(10, 0, -1), // 1 day before 10 years from now
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidDate(tt.date)
			if result != tt.expected {
				t.Errorf("ValidDate(%v) = %v, expected %v", tt.date, result, tt.expected)
			}
		})
	}
}

func TestValidString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "empty string",
			value:    "",
			expected: false,
		},
		{
			name:     "valid string",
			value:    "Hello World",
			expected: true,
		},
		{
			name:     "string with allowed whitespace",
			value:    "Hello\tWorld\nTest\r",
			expected: true,
		},
		{
			name:     "string too long (>10000 chars)",
			value:    strings.Repeat("a", 10001),
			expected: false,
		},
		{
			name:     "string exactly 10000 chars",
			value:    strings.Repeat("a", 10000),
			expected: true,
		},
		{
			name:     "string with null byte",
			value:    "Hello\x00World",
			expected: false,
		},
		{
			name:     "string with control character",
			value:    "Hello\x01World",
			expected: false,
		},
		{
			name:     "invalid UTF-8",
			value:    "user\xff\xfename",
			expected: false,
		},
		{
			name:     "valid multi-byte UTF-8",
			value:    "/café/日本語",
			expected: true,
		},
		{
			name:     "valid string with numbers",
			value:    "user123",
			expected: true,
		},
		{
			name:     "valid string with special chars (safe)",
			value:    "user@domain.com",
			expected: true,
		},
		// SQL-shaped content is stored as data, never parsed as SQL: all
		// queries are parameterized and the logger inserts via COPY. These
		// are accepted so that legitimate traffic is not silently dropped.
		{
			name:     "SQL keyword is ordinary text",
			value:    "SELECT * FROM users",
			expected: true,
		},
		{
			name:     "injection attempt stored verbatim as data",
			value:    "user'; DROP TABLE users; --",
			expected: true,
		},
		{
			name:     "string with backslash",
			value:    "user\\name",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidString(tt.value)
			if result != tt.expected {
				t.Errorf("ValidString(%q) = %v, expected %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestValidHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		expected bool
	}{
		{
			name:     "empty hostname",
			hostname: "",
			expected: false,
		},
		{
			name:     "valid hostname",
			hostname: "example.com",
			expected: true,
		},
		{
			name:     "valid subdomain",
			hostname: "api.example.com",
			expected: true,
		},
		{
			name:     "hostname too long (>255 chars)",
			hostname: strings.Repeat("a", 256),
			expected: false,
		},
		{
			name:     "hostname exactly 255 chars",
			hostname: strings.Repeat("a", 255),
			expected: true,
		},
		{
			name:     "hostname with SQL-shaped content is stored as data",
			hostname: "example.com'; DROP TABLE hosts; --",
			expected: true,
		},
		{
			name:     "localhost",
			hostname: "localhost",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidHostname(tt.hostname)
			if result != tt.expected {
				t.Errorf("ValidHostname(%q) = %v, expected %v", tt.hostname, result, tt.expected)
			}
		})
	}
}

func TestValidPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "valid root path",
			path:     "/",
			expected: true,
		},
		{
			name:     "valid path",
			path:     "/api/v1/users",
			expected: true,
		},
		{
			name:     "path with query params",
			path:     "/search?q=test",
			expected: true,
		},
		{
			name:     "path too long (>255 chars)",
			path:     "/" + strings.Repeat("a", 255),
			expected: false,
		},
		{
			name:     "path exactly 255 chars",
			path:     strings.Repeat("a", 255),
			expected: true,
		},
		{
			name:     "path with SQL-shaped content is stored as data",
			path:     "/users'; DROP TABLE paths; --",
			expected: true,
		},
		// Regression: these ordinary REST paths were rejected by a SQL keyword
		// blacklist, silently dropping the requests at ingest.
		{
			name:     "path containing SQL keyword DELETE",
			path:     "/api/v1/users/delete",
			expected: true,
		},
		{
			name:     "path containing SQL keyword CREATE",
			path:     "/users/create",
			expected: true,
		},
		{
			name:     "path containing SQL keyword SELECT",
			path:     "/api/select-plan",
			expected: true,
		},
		{
			name:     "path containing SQL keyword SCRIPT",
			path:     "/scripts/main.js",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidPath(tt.path)
			if result != tt.expected {
				t.Errorf("ValidPath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestValidUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		userAgent string
		expected  bool
	}{
		{
			name:      "empty user agent",
			userAgent: "",
			expected:  false,
		},
		{
			name:      "valid Chrome user agent",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expected:  true,
		},
		{
			name:      "valid Firefox user agent",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0",
			expected:  true,
		},
		{
			name:      "user agent too long (>255 chars)",
			userAgent: strings.Repeat("a", 256),
			expected:  false,
		},
		{
			name:      "user agent exactly 255 chars",
			userAgent: strings.Repeat("a", 255),
			expected:  true,
		},
		{
			name:      "user agent with SQL-shaped content is stored as data",
			userAgent: "Mozilla'; DROP TABLE agents; --",
			expected:  true,
		},
		{
			name:      "custom bot user agent",
			userAgent: "MyBot/1.0",
			expected:  true,
		},
		// Regression: the semicolons in every real browser user agent matched a
		// SQL injection pattern, so effectively all browser traffic was dropped.
		{
			name:      "full Chrome user agent",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			expected:  true,
		},
		{
			name:      "full Safari user agent",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
			expected:  true,
		},
		{
			name:      "full iPhone user agent",
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidUserAgent(tt.userAgent)
			if result != tt.expected {
				t.Errorf("ValidUserAgent(%q) = %v, expected %v", tt.userAgent, result, tt.expected)
			}
		})
	}
}

func TestValidUserID(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		expected bool
	}{
		{
			name:     "empty user ID",
			userID:   "",
			expected: false,
		},
		{
			name:     "valid user ID",
			userID:   "user123",
			expected: true,
		},
		{
			name:     "valid UUID",
			userID:   "550e8400-e29b-41d4-a716-446655440000",
			expected: true,
		},
		{
			name:     "user ID too long (>255 chars)",
			userID:   strings.Repeat("a", 256),
			expected: false,
		},
		{
			name:     "user ID exactly 255 chars",
			userID:   strings.Repeat("a", 255),
			expected: true,
		},
		{
			name:     "single character user ID",
			userID:   "a",
			expected: true,
		},
		{
			name:     "user ID with SQL-shaped content is stored as data",
			userID:   "user'; DROP TABLE users; --",
			expected: true,
		},
		{
			name:     "user ID with email format",
			userID:   "user@example.com",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidUserID(tt.userID)
			if result != tt.expected {
				t.Errorf("ValidUserID(%q) = %v, expected %v", tt.userID, result, tt.expected)
			}
		})
	}
}

func TestValidLocation(t *testing.T) {
	tests := []struct {
		name     string
		location string
		expected bool
	}{
		{
			name:     "empty location",
			location: "",
			expected: false,
		},
		{
			name:     "valid US location",
			location: "US",
			expected: true,
		},
		{
			name:     "valid UK location",
			location: "GB",
			expected: true,
		},
		{
			name:     "lowercase location code",
			location: "us",
			expected: false,
		},
		{
			name:     "too short",
			location: "U",
			expected: false,
		},
		{
			name:     "too long",
			location: "USA",
			expected: false,
		},
		{
			name:     "with numbers",
			location: "U1",
			expected: false,
		},
		{
			name:     "with special characters",
			location: "U!",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidLocation(tt.location)
			if result != tt.expected {
				t.Errorf("ValidLocation(%q) = %v, expected %v", tt.location, result, tt.expected)
			}
		})
	}
}

func TestValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected bool
	}{
		{
			name:     "valid 200 OK",
			status:   200,
			expected: true,
		},
		{
			name:     "valid 404 Not Found",
			status:   404,
			expected: true,
		},
		{
			name:     "valid 500 Internal Server Error",
			status:   500,
			expected: true,
		},
		{
			name:     "minimum valid status",
			status:   100,
			expected: true,
		},
		{
			name:     "maximum valid status",
			status:   599,
			expected: true,
		},
		{
			name:     "below minimum",
			status:   99,
			expected: false,
		},
		{
			name:     "above maximum",
			status:   600,
			expected: false,
		},
		{
			name:     "negative status",
			status:   -1,
			expected: false,
		},
		{
			name:     "zero status",
			status:   0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidStatus(tt.status)
			if result != tt.expected {
				t.Errorf("ValidStatus(%d) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestValidIPAddress(t *testing.T) {
	tests := []struct {
		name      string
		ipAddress string
		expected  bool
	}{
		{
			name:      "empty IP",
			ipAddress: "",
			expected:  false,
		},
		{
			name:      "valid IPv4",
			ipAddress: "192.168.1.1",
			expected:  true,
		},
		{
			name:      "valid IPv4 localhost",
			ipAddress: "127.0.0.1",
			expected:  true,
		},
		{
			name:      "valid IPv6",
			ipAddress: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected:  true,
		},
		{
			name:      "valid IPv6 localhost",
			ipAddress: "::1",
			expected:  true,
		},
		{
			name:      "valid IPv6 compressed",
			ipAddress: "2001:db8::1",
			expected:  true,
		},
		{
			name:      "invalid IPv4",
			ipAddress: "256.256.256.256",
			expected:  false,
		},
		{
			name:      "invalid IPv4 format",
			ipAddress: "192.168.1",
			expected:  false,
		},
		{
			name:      "invalid IPv6",
			ipAddress: "2001:0db8:85a3::8a2e::7334",
			expected:  false,
		},
		{
			name:      "too long IP (>45 chars)",
			ipAddress: strings.Repeat("1", 46),
			expected:  false,
		},
		{
			name:      "exactly 45 chars valid IPv6",
			ipAddress: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", // 39 chars, but test the boundary
			expected:  true,
		},
		{
			name:      "not an IP address",
			ipAddress: "not.an.ip.address",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidIPAddress(tt.ipAddress)
			if result != tt.expected {
				t.Errorf("ValidIPAddress(%q) = %v, expected %v", tt.ipAddress, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidString(b *testing.B) {
	testString := "This is a test string with some content"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidString(testString)
	}
}

func BenchmarkValidIPAddress(b *testing.B) {
	testIP := "192.168.1.1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidIPAddress(testIP)
	}
}
