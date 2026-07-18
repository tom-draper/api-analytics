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
	// UUID shape: 8-4-4-4-12 lowercase hex characters. This checks the layout
	// only, not the version or variant nibbles, so it accepts any UUID-shaped
	// string, not strictly v4.
	uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// Maximum accepted lengths for stored string fields. These match the
// character varying(255) columns in the requests and user_agents tables (see
// server/self-hosting/database/schema.sql): a longer value cannot be stored,
// so the logger truncates to these lengths before validation.
const (
	MaxPathLength      = 255
	MaxHostnameLength  = 255
	MaxUserAgentLength = 255
	MaxUserIDLength    = 255
	MaxReferrerLength  = 255
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
	if hostname == "" || len(hostname) > MaxHostnameLength {
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

	if len(path) > MaxPathLength {
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

	if len(userAgent) > MaxUserAgentLength {
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

	if len(userID) < 1 || len(userID) > MaxUserIDLength {
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

// cgnatRange is the 100.64.0.0/10 carrier-grade NAT block (RFC 6598), which
// net.IP.IsPrivate does not cover but which is not publicly routable.
var cgnatRange = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

// IsPublicIP reports whether ip is a globally routable public address. It
// returns false for loopback, unspecified, link-local (including the
// 169.254.169.254 cloud metadata endpoint), private (RFC 1918 and IPv6 ULA),
// carrier-grade NAT and multicast addresses. It is the single source of truth
// for SSRF address filtering, used both to validate monitor URLs on submission
// and to guard the monitor pinger's outbound dialer.
func IsPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsPrivate() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil && cgnatRange.Contains(ip4) {
		return false
	}
	return true
}

// ValidAPIKey reports whether a string has UUID shape (used for API keys and
// user IDs). It gates format only; it does not check the UUID version.
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

	// Match UUID shape (layout only, not version/variant)
	return uuidRegex.MatchString(apiKey)
}
