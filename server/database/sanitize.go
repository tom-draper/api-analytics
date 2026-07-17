package database

import (
	"net"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	locationRegex = regexp.MustCompile(`^[A-Z]{2}$`)
	// UUID v4 format: 8-4-4-4-12 hex characters
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
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

// ValidString reports whether a value is safe to store as Postgres text.
//
// Logged values are request data (paths, user agents, referrers), so content
// that merely looks like SQL is legitimate and accepted: every query in this
// package is parameterized, and the logger inserts via COPY, so a value is
// never parsed as SQL. This checks only what Postgres itself cannot store:
// null bytes, control characters, and invalid UTF-8, any of which would abort
// the enclosing batch insert.
func ValidString(value string) bool {
	if value == "" {
		return false
	}

	if len(value) > 10000 {
		return false
	}

	if !utf8.ValidString(value) {
		return false
	}

	for _, r := range value {
		if r == 0 || (unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r') {
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

// ValidAPIKey validates that a string is a valid UUID (used for API keys and user IDs)
func ValidAPIKey(apiKey string) bool {
	if apiKey == "" {
		return false
	}

	// Must be exactly 36 characters (UUID format with hyphens)
	if len(apiKey) != 36 {
		return false
	}

	// Convert to lowercase for validation
	apiKey = strings.ToLower(apiKey)

	// Match UUID v4 format
	return uuidRegex.MatchString(apiKey)
}
