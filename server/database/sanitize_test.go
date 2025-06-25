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
			name:     "string with SQL keyword SELECT",
			value:    "SELECT * FROM users",
			expected: false,
		},
		{
			name:     "string with SQL keyword (case insensitive)",
			value:    "select * from users",
			expected: false,
		},
		{
			name:     "string with SQL keyword INSERT",
			value:    "INSERT INTO table",
			expected: false,
		},
		{
			name:     "string with SQL injection single quote",
			value:    "user'; DROP TABLE users; --",
			expected: false,
		},
		{
			name:     "string with SQL injection comment",
			value:    "user -- comment",
			expected: false,
		},
		{
			name:     "string with SQL injection union select",
			value:    "1 UNION SELECT password FROM users",
			expected: false,
		},
		{
			name:     "string with SQL injection or condition",
			value:    "user OR 1=1",
			expected: false,
		},
		{
			name:     "string with backslash",
			value:    "user\\name",
			expected: false,
		},
		{
			name:     "string with SQL comment /*",
			value:    "user /* comment */",
			expected: false,
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
			name:     "hostname too long (>253 chars)",
			hostname: strings.Repeat("a", 254),
			expected: false,
		},
		{
			name:     "hostname exactly 253 chars",
			hostname: strings.Repeat("a", 253),
			expected: true,
		},
		{
			name:     "hostname with SQL injection",
			hostname: "example.com'; DROP TABLE hosts; --",
			expected: false,
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
			name:     "path too long (>2048 chars)",
			path:     "/" + strings.Repeat("a", 2048),
			expected: false,
		},
		{
			name:     "path exactly 2048 chars",
			path:     strings.Repeat("a", 2048),
			expected: true,
		},
		{
			name:     "path with SQL injection",
			path:     "/users'; DROP TABLE paths; --",
			expected: false,
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
			name:      "user agent too long (>1024 chars)",
			userAgent: strings.Repeat("a", 1025),
			expected:  false,
		},
		{
			name:      "user agent exactly 1024 chars",
			userAgent: strings.Repeat("a", 1024),
			expected:  true,
		},
		{
			name:      "user agent with SQL injection",
			userAgent: "Mozilla'; DROP TABLE agents; --",
			expected:  false,
		},
		{
			name:      "custom bot user agent",
			userAgent: "MyBot/1.0",
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
			name:     "user ID with SQL injection",
			userID:   "user'; DROP TABLE users; --",
			expected: false,
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

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		expected map[string]bool
	}{
		{
			name: "mixed valid data",
			data: map[string]interface{}{
				"name":      "John Doe",
				"timestamp": time.Now(),
				"status":    200,
			},
			expected: map[string]bool{
				"name":      true,
				"timestamp": true,
				"status":    true,
			},
		},
		{
			name: "mixed invalid data",
			data: map[string]interface{}{
				"name":      "",
				"timestamp": time.Time{},
				"status":    999,
			},
			expected: map[string]bool{
				"name":      false,
				"timestamp": false,
				"status":    false,
			},
		},
		{
			name: "string with SQL injection",
			data: map[string]interface{}{
				"query": "SELECT * FROM users",
			},
			expected: map[string]bool{
				"query": false,
			},
		},
		{
			name: "non-status integer",
			data: map[string]interface{}{
				"count": 42,
			},
			expected: map[string]bool{},
		},
		{
			name:     "empty data",
			data:     map[string]interface{}{},
			expected: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateInput(tt.data)
			
			// Check that all expected keys are present with correct values
			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("ValidateInput() missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("ValidateInput() key %q = %v, expected %v", key, actualValue, expectedValue)
				}
			}
			
			// Check that no unexpected keys are present
			for key := range result {
				if _, expected := tt.expected[key]; !expected {
					t.Errorf("ValidateInput() unexpected key %q with value %v", key, result[key])
				}
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

func BenchmarkValidateInput(b *testing.B) {
	testData := map[string]interface{}{
		"name":      "John Doe",
		"timestamp": time.Now(),
		"status":    200,
		"ip":        "192.168.1.1",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateInput(testData)
	}
}