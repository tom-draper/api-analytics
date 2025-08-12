package ratelimit

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu    sync.RWMutex
	users map[string]*userRate
}

const (
	accessesPerMinute int           = 10
	cleanupInterval   time.Duration = 5 * time.Minute
	maxUsers          int           = 10000 // Prevent memory leaks
)

type userRate struct {
	mu         sync.Mutex
	timestamps []time.Time
	lastAccess time.Time
}

// NewRateLimiter creates a new rate limiter with cleanup
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		users: make(map[string]*userRate),
	}
	
	// Start cleanup goroutine
	go rl.cleanup()
	
	return rl
}

// cleanup removes old users to prevent memory leaks
func (r *RateLimiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		r.mu.Lock()
		
		// Remove users who haven't accessed in the last hour
		cutoff := time.Now().Add(-time.Hour)
		for apiKey, user := range r.users {
			user.mu.Lock()
			if user.lastAccess.Before(cutoff) {
				delete(r.users, apiKey)
			}
			user.mu.Unlock()
		}
		
		// If still too many users, remove oldest half
		if len(r.users) > maxUsers {
			count := 0
			target := len(r.users) / 2
			for apiKey := range r.users {
				delete(r.users, apiKey)
				count++
				if count >= target {
					break
				}
			}
		}
		
		r.mu.Unlock()
	}
}

// RateLimited checks if the API key is rate limited
func (r *RateLimiter) RateLimited(apiKey string) bool {
	// Fast path: check if user exists with read lock
	r.mu.RLock()
	user, exists := r.users[apiKey]
	r.mu.RUnlock()
	
	if exists {
		return user.isRateLimited()
	}
	
	// Slow path: create new user with write lock
	r.mu.Lock()
	// Double-check in case another goroutine added it
	user, exists = r.users[apiKey]
	if !exists {
		user = newUserRate()
		r.users[apiKey] = user
	}
	r.mu.Unlock()
	
	return user.isRateLimited()
}

// newUserRate creates a new user rate tracker
func newUserRate() *userRate {
	now := time.Now()
	return &userRate{
		timestamps: []time.Time{now},
		lastAccess: now,
	}
}

// isRateLimited checks if this user is rate limited
func (u *userRate) isRateLimited() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	now := time.Now()
	u.lastAccess = now
	
	// Remove timestamps older than 1 minute
	cutoff := now.Add(-time.Minute)
	validCount := 0
	
	// Count and keep only valid timestamps
	for _, ts := range u.timestamps {
		if ts.After(cutoff) {
			u.timestamps[validCount] = ts
			validCount++
		}
	}
	u.timestamps = u.timestamps[:validCount]
	
	// Check if rate limited
	if len(u.timestamps) >= accessesPerMinute {
		return true
	}
	
	// Add current timestamp
	u.timestamps = append(u.timestamps, now)
	return false
}