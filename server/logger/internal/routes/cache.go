package routes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/tom-draper/api-analytics/server/database"
	"github.com/tom-draper/api-analytics/server/logger/internal/log"
)

type Cache struct {
	userAgentMap map[string]int
	userAgentMu  sync.RWMutex
	geoIPMap     map[string]*geoIPEntry
	geoIPMu      sync.RWMutex
	maxSize      int
}

type geoIPEntry struct {
	countryCode string
	lastAccess  int64 // unix nanoseconds; read/written atomically
}

func newCache(maxSize int) *Cache {
	return &Cache{
		userAgentMap: make(map[string]int),
		geoIPMap:     make(map[string]*geoIPEntry),
		maxSize:      maxSize,
	}
}

func preloadUserAgentCache(ctx context.Context, db *database.DB, cache *Cache) (int, error) {
	rows, err := db.Pool.Query(ctx, "SELECT user_agent, id FROM user_agents LIMIT 50000")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type entry struct {
		ua string
		id int
	}
	entries := make([]entry, 0, 50000)
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.ua, &e.id); err != nil {
			continue
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	cache.userAgentMu.Lock()
	for _, e := range entries {
		cache.userAgentMap[e.ua] = e.id
	}
	cache.userAgentMu.Unlock()

	return len(entries), nil
}

func ensureUserAgentIDs(ctx context.Context, db *database.DB, cache *Cache, userAgents []string) (map[string]int, error) {
	if len(userAgents) == 0 {
		return make(map[string]int), nil
	}

	result := make(map[string]int, len(userAgents))
	newUserAgents := make([]string, 0, len(userAgents))

	cache.userAgentMu.RLock()
	for _, ua := range userAgents {
		if id, exists := cache.userAgentMap[ua]; exists {
			result[ua] = id
		} else {
			newUserAgents = append(newUserAgents, ua)
		}
	}
	cache.userAgentMu.RUnlock()

	if len(newUserAgents) == 0 {
		return result, nil
	}

	// Bulk insert with ON CONFLICT handles concurrent inserts correctly.
	// COPY does not support ON CONFLICT, so we use unnest() instead.
	_, err := db.Pool.Exec(ctx,
		"INSERT INTO user_agents (user_agent) SELECT unnest($1::text[]) ON CONFLICT (user_agent) DO NOTHING",
		newUserAgents,
	)
	if err != nil {
		log.Error(fmt.Sprintf("failed to insert user agents: %v", err))
	}

	rows, err := db.Pool.Query(ctx,
		"SELECT user_agent, id FROM user_agents WHERE user_agent = ANY($1)",
		newUserAgents,
	)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	cache.userAgentMu.Lock()
	for rows.Next() {
		var userAgent string
		var id int
		if err := rows.Scan(&userAgent, &id); err != nil {
			continue
		}
		result[userAgent] = id
		if len(cache.userAgentMap) < cache.maxSize {
			cache.userAgentMap[userAgent] = id
		}
	}
	cache.userAgentMu.Unlock()

	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

func getCountryCode(geoIPDB *geoip2.Reader, cache *Cache, ipAddress string) string {
	if ipAddress == "" || geoIPDB == nil {
		return ""
	}

	now := time.Now().Unix()

	cache.geoIPMu.RLock()
	if entry, exists := cache.geoIPMap[ipAddress]; exists {
		atomic.StoreInt64(&entry.lastAccess, now)
		countryCode := entry.countryCode
		cache.geoIPMu.RUnlock()
		return countryCode
	}
	cache.geoIPMu.RUnlock()

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return ""
	}

	record, err := geoIPDB.Country(ip)
	if err != nil {
		return ""
	}

	countryCode := record.Country.IsoCode

	cache.geoIPMu.Lock()
	// Double-check: another goroutine may have inserted while we did the lookup
	if _, exists := cache.geoIPMap[ipAddress]; !exists {
		if len(cache.geoIPMap) >= cache.maxSize {
			evictLRUEntries(cache)
		}
		cache.geoIPMap[ipAddress] = &geoIPEntry{
			countryCode: countryCode,
			lastAccess:  now,
		}
	}
	cache.geoIPMu.Unlock()

	return countryCode
}

// evictLRUEntries removes the least recently used entries from the GeoIP cache.
// Caller must hold cache.geoIPMu write lock.
func evictLRUEntries(cache *Cache) {
	cutoff := time.Now().Unix() - 3600
	var toDelete []string

	for ip, entry := range cache.geoIPMap {
		if atomic.LoadInt64(&entry.lastAccess) < cutoff {
			toDelete = append(toDelete, ip)
		}
	}

	for _, ip := range toDelete {
		delete(cache.geoIPMap, ip)
	}

	if len(cache.geoIPMap) >= cache.maxSize {
		type entry struct {
			ip         string
			lastAccess int64
		}
		entries := make([]entry, 0, len(cache.geoIPMap))
		for ip, e := range cache.geoIPMap {
			entries = append(entries, entry{ip: ip, lastAccess: atomic.LoadInt64(&e.lastAccess)})
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].lastAccess < entries[j].lastAccess
		})

		removeCount := max(len(entries)/4, 1)
		for i := 0; i < removeCount; i++ {
			delete(cache.geoIPMap, entries[i].ip)
		}
	}
}

var hasherPool = sync.Pool{
	New: func() any { return sha256.New() },
}

func getUserHash(ipAddress, userAgent string) string {
	if ipAddress == "" && userAgent == "" {
		return ""
	}

	h := hasherPool.Get().(hash.Hash)
	h.Reset()
	h.Write([]byte(ipAddress + "|" + userAgent))
	result := hex.EncodeToString(h.Sum(nil))[:32]
	hasherPool.Put(h)
	return result
}
