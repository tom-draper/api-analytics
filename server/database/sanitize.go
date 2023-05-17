package database

import (
	"net"
	"strings"
	"time"
)

func SanitizeDate(date time.Time) bool {
	return !date.IsZero()
}

func SanitizeString(value string) bool {
	value = strings.ToUpper(value)
	return (value != "" &&
		!strings.Contains(value, "DROP TABLE") &&
		!strings.Contains(value, "INSERT") &&
		!strings.Contains(value, "UPDATE") &&
		!strings.Contains(value, "SELECT") &&
		!strings.Contains(value, "--") &&
		!strings.Contains(value, "'"))
}

func SanitizeHostname(hostname string) bool {
	hostname = strings.ToUpper(hostname)
	return (hostname != "" &&
		!strings.Contains(hostname, "DROP TABLE") &&
		!strings.Contains(hostname, "INSERT") &&
		!strings.Contains(hostname, "UPDATE") &&
		!strings.Contains(hostname, "SELECT") &&
		!strings.Contains(hostname, "--") &&
		!strings.Contains(hostname, "'"))
}

func SanitizePath(path string) bool {
	return SanitizeString(path)
}

func SanitizeUserAgent(userAgent string) bool {
	return SanitizeString(userAgent)
}

func SanitizeLocation(location string) bool {
	return (len(location) == 2 &&
		!strings.Contains(location, "--") &&
		!strings.Contains(location, ";") &&
		!strings.Contains(location, "'"))
}

func SanitizeStatus(status int) bool {
	return status >= 100 && status <= 599
}

func SanitizeIPAddress(ipAddress string) bool {
	if ipAddress == "" {
		return false
	}
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}
	return true
}
