package routes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/tom-draper/api-analytics/server/database"
)

var hex32 = regexp.MustCompile(`^[0-9a-f]{32}$`)

func TestGetUserHash(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		a := getUserHash("", "1.2.3.4", "Mozilla/5.0")
		b := getUserHash("", "1.2.3.4", "Mozilla/5.0")
		if a != b {
			t.Errorf("same input hashed differently: %q vs %q", a, b)
		}
		if !hex32.MatchString(a) {
			t.Errorf("hash %q is not 32 hex characters", a)
		}
	})

	t.Run("distinct inputs differ", func(t *testing.T) {
		if getUserHash("", "1.2.3.4", "UA") == getUserHash("", "5.6.7.8", "UA") {
			t.Error("different IPs produced the same hash")
		}
		if getUserHash("", "1.2.3.4", "UA-a") == getUserHash("", "1.2.3.4", "UA-b") {
			t.Error("different user agents produced the same hash")
		}
	})

	t.Run("empty when both fields empty", func(t *testing.T) {
		if got := getUserHash("", "", ""); got != "" {
			t.Errorf("empty inputs = %q, want empty", got)
		}
	})

	t.Run("field separator prevents collisions", func(t *testing.T) {
		// Without a delimiter, ("ab","c") and ("a","bc") would hash the same.
		if getUserHash("", "ab", "c") == getUserHash("", "a", "bc") {
			t.Error("hash does not delimit IP from user agent")
		}
	})

	t.Run("empty secret matches legacy unsalted hash", func(t *testing.T) {
		// An unset secret must not change the digest, so hashes stay stable
		// across deployments that never set USER_HASH_SECRET.
		want := sha256.Sum256([]byte("1.2.3.4|Mozilla/5.0"))
		if got := getUserHash("", "1.2.3.4", "Mozilla/5.0"); got != hex.EncodeToString(want[:])[:32] {
			t.Errorf("empty-secret hash = %q, does not match legacy unsalted digest", got)
		}
	})

	t.Run("secret changes the hash and stays deterministic", func(t *testing.T) {
		unsalted := getUserHash("", "1.2.3.4", "UA")
		peppered := getUserHash("pepper", "1.2.3.4", "UA")
		if peppered == unsalted {
			t.Error("secret did not change the hash")
		}
		if peppered != getUserHash("pepper", "1.2.3.4", "UA") {
			t.Error("peppered hash is not deterministic")
		}
		if getUserHash("pepper-a", "1.2.3.4", "UA") == getUserHash("pepper-b", "1.2.3.4", "UA") {
			t.Error("different secrets produced the same hash")
		}
	})
}

func TestEnsureUserAgentIDsEmpty(t *testing.T) {
	// No user agents: returns an empty map without touching the database, so a
	// nil DB is safe.
	got, err := ensureUserAgentIDs(context.Background(), nil, newCache(10, 10), nil)
	if err != nil {
		t.Fatalf("ensureUserAgentIDs: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestEnsureUserAgentIDsAllCached(t *testing.T) {
	cache := newCache(10, 10)
	cache.userAgentMap["Mozilla/5.0"] = 1
	cache.userAgentMap["curl/8.4.0"] = 2

	// Every user agent is already cached, so ensureUserAgentIDs must resolve them
	// without a database round trip (nil DB proves it).
	got, err := ensureUserAgentIDs(context.Background(), nil, cache, []string{"Mozilla/5.0", "curl/8.4.0"})
	if err != nil {
		t.Fatalf("ensureUserAgentIDs: %v", err)
	}
	if got["Mozilla/5.0"] != 1 || got["curl/8.4.0"] != 2 {
		t.Errorf("cache lookup returned %v", got)
	}
}

func TestEvictUserAgents(t *testing.T) {
	// A cache at capacity must shed entries so newly seen agents can still be
	// admitted, rather than freezing at the cap.
	cache := newCache(10, 8)
	for i := 0; i < cache.userAgentMaxSize; i++ {
		cache.userAgentMap[fmt.Sprintf("agent/%d", i)] = i
	}

	evictUserAgents(cache)

	// Roughly a quarter is removed, leaving room for new entries.
	if got := len(cache.userAgentMap); got != 6 {
		t.Errorf("after eviction len = %d, want 6 (dropped a quarter of 8)", got)
	}
	if len(cache.userAgentMap) >= cache.userAgentMaxSize {
		t.Error("eviction left no room for new user agents")
	}
}

// The ip_address column is a Postgres cidr, so these client-supplied values must
// be rejected before insert; a single one would otherwise abort the whole COPY
// batch. Guards the ingest-side use of database.ValidIPAddress.
func TestValidIPAddressRejectsBatchPoisoners(t *testing.T) {
	poisoners := []string{
		"203.0.113.1, 70.41.3.18", // raw X-Forwarded-For list
		"unknown",
		"1.2.3.4:8080", // host:port
		"not-an-ip",
		"1.2.3",
	}
	for _, ip := range poisoners {
		if database.ValidIPAddress(ip) {
			t.Errorf("ValidIPAddress(%q) = true, want false (would poison the COPY batch)", ip)
		}
	}
	for _, ip := range []string{"203.0.113.1", "2001:db8::1"} {
		if !database.ValidIPAddress(ip) {
			t.Errorf("ValidIPAddress(%q) = false, want true (a real address must still store)", ip)
		}
	}
}

func TestGetCountryCodeGuards(t *testing.T) {
	cache := newCache(10, 10)
	// A nil GeoIP reader must not panic and yields no location.
	if got := getCountryCode(nil, cache, "1.2.3.4"); got != "" {
		t.Errorf("nil GeoIP db = %q, want empty", got)
	}
	// An empty IP short-circuits regardless of the reader.
	if got := getCountryCode(nil, cache, ""); got != "" {
		t.Errorf("empty ip = %q, want empty", got)
	}
}

func TestEvictLRUEntriesRemovesStale(t *testing.T) {
	cache := newCache(100, 100)
	now := time.Now().Unix()

	// Two entries older than the one hour cutoff, one fresh.
	cache.geoIPMap["stale-1"] = &geoIPEntry{countryCode: "US", lastAccess: now - 7200}
	cache.geoIPMap["stale-2"] = &geoIPEntry{countryCode: "GB", lastAccess: now - 5000}
	cache.geoIPMap["fresh"] = &geoIPEntry{countryCode: "FR", lastAccess: now}

	evictLRUEntries(cache)

	if _, ok := cache.geoIPMap["stale-1"]; ok {
		t.Error("stale-1 should have been evicted")
	}
	if _, ok := cache.geoIPMap["stale-2"]; ok {
		t.Error("stale-2 should have been evicted")
	}
	if _, ok := cache.geoIPMap["fresh"]; !ok {
		t.Error("fresh entry should have survived")
	}
}

func TestEvictLRUEntriesFallsBackToLRU(t *testing.T) {
	// All entries are fresh (none past the cutoff), and the map is at capacity,
	// so eviction must remove the least recently used quarter.
	cache := newCache(4, 4)
	now := time.Now().Unix()
	for i, key := range []string{"a", "b", "c", "d"} {
		// a is oldest, d is newest, all within the last hour.
		cache.geoIPMap[key] = &geoIPEntry{countryCode: "US", lastAccess: now - int64(40-i*10)}
	}

	evictLRUEntries(cache)

	// removeCount = max(len/4, 1) = 1, so the single oldest ("a") goes.
	if _, ok := cache.geoIPMap["a"]; ok {
		t.Error("least recently used entry 'a' should have been evicted")
	}
	if len(cache.geoIPMap) != 3 {
		t.Errorf("expected 3 entries after LRU eviction, got %d", len(cache.geoIPMap))
	}
}
