package database

import (
	"net"
	"strings"
	"time"
)

func ValidDate(date time.Time) bool {
	return !date.IsZero()
}

func ValidString(value string) bool {
	value = strings.ToUpper(value)
	return (!strings.Contains(value, "DROP TABLE") &&
		!strings.Contains(value, "INSERT") &&
		!strings.Contains(value, "UPDATE") &&
		!strings.Contains(value, "SELECT") &&
		!strings.Contains(value, "--") &&
		!strings.Contains(value, "'"))
}

func ValidHostname(hostname string) bool {
	return ValidString(hostname)
}

func ValidPath(path string) bool {
	return ValidString(path)
}

func ValidUserAgent(userAgent string) bool {
	return ValidString(userAgent)
}

func ValidUserID(userID string) bool {
	return ValidString(userID)
}

func ValidLocation(location string) bool {
	return (len(location) == 2 &&
		!strings.Contains(location, "--") &&
		!strings.Contains(location, ";") &&
		!strings.Contains(location, "'"))
}

func ValidStatus(status int) bool {
	return status >= 100 && status <= 599
}

func ValidIPAddress(ipAddress string) bool {
	if ipAddress == "" {
		return false
	}
	ip := net.ParseIP(ipAddress)
	return ip != nil
}
