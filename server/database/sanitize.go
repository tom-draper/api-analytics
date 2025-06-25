package database

import (
	"net"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	sqlKeywords = []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"TRUNCATE", "EXEC", "EXECUTE", "UNION", "SCRIPT", "DECLARE",
	}

	sqlPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)('|(\\)|;|--|/\*|\*/|xp_|sp_)`),
		regexp.MustCompile(`(?i)(union\s+select|drop\s+table|insert\s+into)`),
		regexp.MustCompile(`(?i)(\bor\b|\band\b)\s*['"]?\s*\d+\s*['"]?\s*[=><]`),
	}

	locationRegex = regexp.MustCompile(`^[A-Z]{2}$`)
)

func ValidDate(date time.Time) bool {
	if date.IsZero() {
		return false
	}

	// Check if date is within reasonable bounds (not too far in past/future)
	now := time.Now()
	minDate := now.AddDate(-100, 0, 0) // 100 years ago
	maxDate := now.AddDate(10, 0, 0)   // 10 years in future

	return date.After(minDate) && date.Before(maxDate)
}

func ValidString(value string) bool {
	if value == "" {
		return false
	}

	if len(value) > 10000 {
		return false
	}

	for _, r := range value {
		if r == 0 || (unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r') {
			return false
		}
	}

	upperValue := strings.ToUpper(value)

	// Check for SQL keywords
	for _, keyword := range sqlKeywords {
		if strings.Contains(upperValue, keyword) {
			return false
		}
	}

	// Check for SQL injection patterns
	for _, pattern := range sqlPatterns {
		if pattern.MatchString(value) {
			return false
		}
	}

	return true
}

func ValidHostname(hostname string) bool {
	if hostname == "" || len(hostname) > 253 {
		return false
	}

	if !ValidString(hostname) {
		return false
	}

	return true
}

func ValidPath(path string) bool {
	if path == "" {
		return false
	}

	if len(path) > 2048 {
		return false
	}

	if !ValidString(path) {
		return false
	}

	return true
}

func ValidUserAgent(userAgent string) bool {
	if userAgent == "" {
		return false
	}

	if len(userAgent) > 1024 {
		return false
	}

	if !ValidString(userAgent) {
		return false
	}

	return true
}

func ValidUserID(userID string) bool {
	if userID == "" {
		return false
	}

	if len(userID) < 1 || len(userID) > 255 {
		return false
	}

	if !ValidString(userID) {
		return false
	}

	return true
}

func ValidLocation(location string) bool {
	if len(location) != 2 {
		return false
	}

	// Must be uppercase letters only
	return locationRegex.MatchString(location)
}

func ValidStatus(status int) bool {
	return status >= 100 && status <= 599
}

func ValidIPAddress(ipAddress string) bool {
	if ipAddress == "" {
		return false
	}

	if len(ipAddress) > 45 { // Max IPv6 length
		return false
	}

	ip := net.ParseIP(ipAddress)
	return ip != nil
}

func ValidateInput(data map[string]interface{}) map[string]bool {
	results := make(map[string]bool)

	for key, value := range data {
		switch v := value.(type) {
		case string:
			results[key] = ValidString(v)
		case time.Time:
			results[key] = ValidDate(v)
		case int:
			if key == "status" {
				results[key] = ValidStatus(v)
			}
		}
	}

	return results
}
