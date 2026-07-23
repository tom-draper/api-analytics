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
	userAgentMap     map[string]int
	userAgentMu      sync.RWMutex
	userAgentMaxSize int
	geoIPMap         map[string]*geoIPEntry
	geoIPMu          sync.RWMutex
	geoIPMaxSize     int
}

type geoIPEntry struct {
	countryCode string
	lastAccess  int64 // unix seconds; read/written atomically
}

// newCache builds the caches with independent capacities. The user-agent map is
// far larger because the full set is preloaded and reused across every request,
// whereas the geoIP map is a bounded LRU of recently seen IPs.
func newCache(geoIPMaxSize, userAgentMaxSize int) *Cache {
	return &Cache{
		userAgentMap:     make(map[string]int),
		userAgentMaxSize: userAgentMaxSize,
		geoIPMap:         make(map[string]*geoIPEntry),
		geoIPMaxSize:     geoIPMaxSize,
	}
}

// reseedUserAgentSequence advances the user_agents id sequence past the current
// maximum id. A restore or backfill that inserts rows with explicit ids can
// leave the serial sequence behind max(id); the next generated id would then
// collide with an existing primary key and fail every insert, silently dropping
// requests that carry a new user agent. Reseeding once at startup closes that
// gap. It is a no-op on an empty table so the first insert still gets id 1.
func reseedUserAgentSequence(ctx context.Context, db *database.DB) error {
	_, err := db.Pool.Exec(ctx, `
		SELECT setval(
			pg_get_serial_sequence('user_agents', 'id'),
			(SELECT MAX(id) FROM user_agents)
		)
		WHERE EXISTS (SELECT 1 FROM user_agents)`)
	return err
}

func preloadUserAgentCache(ctx context.Context, db *database.DB, cache *Cache) (int, error) {
	rows, err := db.Pool.Query(ctx, "SELECT user_agent, id FROM user_agents LIMIT $1", cache.userAgentMaxSize)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type entry struct {
		ua string
		id int
	}
	entries := make([]entry, 0, cache.userAgentMaxSize)
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
		if len(cache.userAgentMap) >= cache.userAgentMaxSize {
			evictUserAgents(cache)
		}
		cache.userAgentMap[userAgent] = id
	}
	cache.userAgentMu.Unlock()

	if err := rows.Err(); err != nil {
		return result, err
	}

	return result, nil
}

// evictUserAgents drops roughly a quarter of the user-agent cache to make room
// once it reaches capacity, so the cache keeps admitting newly seen agents
// instead of freezing at its cap. The map carries no access metadata, so
// eviction is arbitrary; a dropped agent is simply re-resolved (and re-cached)
// on its next request, so correctness is unaffected. Caller must hold
// cache.userAgentMu write lock.
func evictUserAgents(cache *Cache) {
	target := max(len(cache.userAgentMap)/4, 1)
	for ua := range cache.userAgentMap {
		delete(cache.userAgentMap, ua)
		target--
		if target == 0 {
			break
		}
	}
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
		if len(cache.geoIPMap) >= cache.geoIPMaxSize {
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

	if len(cache.geoIPMap) >= cache.geoIPMaxSize {
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

// hashSeparator delimits the hash inputs. Kept as a package-level byte slice so
// it isn't reallocated on every getUserHash call (it's on the per-request path).
var hashSeparator = []byte("|")

// getUserHash derives an anonymous per-user identifier from the IP and user
// agent. secret is a server-side pepper: when set, it is mixed into the digest
// so the output cannot be brute-forced back to the source IP (the IPv4 space is
// small enough to enumerate against an unsalted hash). An empty secret keeps the
// legacy unsalted digest, so hashes are stable across deployments that do not
// set one.
func getUserHash(secret, ipAddress, userAgent string) string {
	if ipAddress == "" && userAgent == "" {
		return ""
	}

	h := hasherPool.Get().(hash.Hash)
	h.Reset()
	if secret != "" {
		h.Write([]byte(secret))
		h.Write(hashSeparator)
	}
	// Write the fields separately rather than concatenating them, so no
	// intermediate joined string is allocated. The byte stream is identical to
	// ipAddress+"|"+userAgent, so the digest stays compatible.
	h.Write([]byte(ipAddress))
	h.Write(hashSeparator)
	h.Write([]byte(userAgent))

	// Sum into a stack array (cap == sha256.Size) so no heap slice is allocated,
	// then hex-encode the first 16 bytes into a fixed buffer. The result is the
	// only allocation, versus the previous 32-byte digest plus 64-char string.
	var digest [sha256.Size]byte
	sum := h.Sum(digest[:0])
	hasherPool.Put(h)

	var out [32]byte
	hex.Encode(out[:], sum[:16])
	return string(out[:])
}
