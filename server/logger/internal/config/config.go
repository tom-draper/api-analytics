package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
)

// Config holds validated configuration for the logger service
type Config struct {
	PostgresURL string
	Port        int
	RateLimit   int
	IPRateLimit int
	MaxInsert   int
	HashSecret  string
	// AllowedOrigins restricts CORS to these origins. Empty means allow all
	// origins (the historical default), which is safe for an ingest endpoint
	// that is called server-to-server and gated by an API key, not by the
	// browser's ambient credentials.
	AllowedOrigins []string
	// TrustedProxies is the set of proxy CIDRs/IPs whose X-Forwarded-For header
	// is believed. Only when the direct peer is in this set is the forwarded
	// client IP used for the per-IP rate limit and client-error logs; a request
	// arriving directly from a public address cannot spoof its IP to get a fresh
	// rate-limit budget. Defaults to loopback and the private ranges.
	TrustedProxies []string
}

// defaultTrustedProxies trusts only loopback and RFC1918/ULA private ranges, so
// a reverse proxy on the same host or private network is believed while a
// direct public client cannot forge X-Forwarded-For.
var defaultTrustedProxies = []string{
	"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
	"::1/128", "fc00::/7",
}

// Load loads environment variables and validates them
func Load() (*Config, error) {
	// Load .env file (non-fatal if missing)
	if err := godotenv.Load(".env"); err != nil {
		log.Info("could not load .env file, using environment variables")
	}

	cfg := &Config{
		PostgresURL:    os.Getenv("POSTGRES_URL"),
		Port:           getIntWithDefault("LOGGER_PORT", 8000),
		RateLimit:      getIntWithDefault("LOGGER_RATE_LIMIT", 10),
		IPRateLimit:    getIntWithDefault("LOGGER_IP_RATE_LIMIT", 100),
		MaxInsert:      getIntWithDefault("LOGGER_MAX_INSERT", 2000),
		HashSecret:     os.Getenv("USER_HASH_SECRET"),
		AllowedOrigins: parseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}

	trustedProxies, err := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	if err != nil {
		return nil, err
	}
	cfg.TrustedProxies = trustedProxies

	// Validate required fields
	if cfg.PostgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is required")
	}

	if cfg.HashSecret == "" {
		log.Info("USER_HASH_SECRET not set: user hashes are unsalted and can be brute-forced back to the source IP; set it to anonymise stored hashes")
	}

	// Validate ranges
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("LOGGER_PORT must be between 1 and 65535, got %d", cfg.Port)
	}

	// Validate ranges
	if cfg.RateLimit < 1 || cfg.RateLimit > 1000 {
		return nil, fmt.Errorf("LOGGER_RATE_LIMIT must be between 1 and 1000, got %d", cfg.RateLimit)
	}

	if cfg.IPRateLimit < 1 || cfg.IPRateLimit > 100000 {
		return nil, fmt.Errorf("LOGGER_IP_RATE_LIMIT must be between 1 and 100000, got %d", cfg.IPRateLimit)
	}

	if cfg.MaxInsert < 1 || cfg.MaxInsert > 10000 {
		return nil, fmt.Errorf("LOGGER_MAX_INSERT must be between 1 and 10000, got %d", cfg.MaxInsert)
	}

	log.Info(fmt.Sprintf("configuration loaded: port=%d, rate_limit=%d, ip_rate_limit=%d, max_insert=%d", cfg.Port, cfg.RateLimit, cfg.IPRateLimit, cfg.MaxInsert))

	return cfg, nil
}

// parseTrustedProxies reads a comma-separated list of proxy CIDRs or IPs,
// falling back to the safe private/loopback default when unset. Each entry is
// validated so a typo fails startup rather than silently disabling XFF trust.
func parseTrustedProxies(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return defaultTrustedProxies, nil
	}
	var proxies []string
	for _, entry := range strings.Split(value, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, _, err := net.ParseCIDR(entry); err != nil {
			if net.ParseIP(entry) == nil {
				return nil, fmt.Errorf("TRUSTED_PROXIES entry %q is not a valid IP or CIDR", entry)
			}
		}
		proxies = append(proxies, entry)
	}
	if len(proxies) == 0 {
		return defaultTrustedProxies, nil
	}
	return proxies, nil
}

// parseOrigins splits a comma-separated CORS origin list, trimming blanks.
func parseOrigins(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var origins []string
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

// getIntWithDefault is a helper for parsing integer environment variables
func getIntWithDefault(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Info(fmt.Sprintf("invalid integer for %s, using default %d", name, defaultValue))
		return defaultValue
	}

	return value
}
