package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	cleanupInterval = 5 * time.Minute
	inactiveTimeout = time.Hour
	maxKeys         = 10000
)

// RateLimiter enforces a per-key token bucket rate limit.
type RateLimiter struct {
	mu   sync.RWMutex
	keys map[string]*tokenBucket
	rate float64 // tokens per nanosecond
	max  float64 // maximum token accumulation (burst = full per-minute quota)
}

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill int64 // unix nanoseconds
	lastAccess int64 // unix nanoseconds; accessed atomically for cleanup reads
}

// NewRateLimiter creates a rate limiter allowing requestsPerMinute calls per key.
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		keys: make(map[string]*tokenBucket),
		rate: float64(requestsPerMinute) / float64(time.Minute),
		max:  float64(requestsPerMinute),
	}
	go rl.cleanup()
	return rl
}

// RateLimited returns true if the key has exceeded its rate limit.
func (r *RateLimiter) RateLimited(key string) bool {
	r.mu.RLock()
	bucket, exists := r.keys[key]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()
		bucket, exists = r.keys[key]
		if !exists {
			now := time.Now().UnixNano()
			bucket = &tokenBucket{
				tokens:     r.max,
				lastRefill: now,
				lastAccess: now,
			}
			r.keys[key] = bucket
		}
		r.mu.Unlock()
	}

	return bucket.consume(r.rate, r.max)
}

func (b *tokenBucket) consume(rate, max float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now().UnixNano()
	elapsed := float64(now - b.lastRefill)
	b.tokens = min(b.tokens+elapsed*rate, max)
	b.lastRefill = now
	atomic.StoreInt64(&b.lastAccess, now)

	if b.tokens < 1 {
		return true
	}
	b.tokens--
	return false
}

func (r *RateLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-inactiveTimeout).UnixNano()
		r.mu.Lock()
		for key, bucket := range r.keys {
			if atomic.LoadInt64(&bucket.lastAccess) < cutoff {
				delete(r.keys, key)
			}
		}
		if len(r.keys) > maxKeys {
			count := 0
			target := len(r.keys) / 2
			for key := range r.keys {
				delete(r.keys, key)
				count++
				if count >= target {
					break
				}
			}
		}
		r.mu.Unlock()
	}
}
