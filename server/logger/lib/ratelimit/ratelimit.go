package ratelimit

import (
	"time"
)

type RateLimiter map[string]*userRate

const accessesPerMinute int = 10

type userRate struct {
	timestamps [accessesPerMinute]time.Time // Circular array
	current    int                          // Index of oldest timestamp to be replaced next
}

func (u *userRate) increment() {
	if u.current < len(u.timestamps)-1 {
		u.current += 1
	} else {
		u.current = 0
	}
}

func (u *userRate) rateLimited() bool {
	oldest := u.timestamps[u.current]

	// If the oldest timestamp recorded is less than a minute ago -> rate limited
	if time.Since(oldest) < time.Minute {
		// Rate limited, user access denied, do not record access
		return true
	} else {
		u.recordAccess() // Register user access and record current time
		return false
	}
}

func (u *userRate) recordAccess() {
	u.timestamps[u.current] = time.Now()
	u.increment()
}

func newUserRate() *userRate {
	ur := userRate{}
	ur.recordAccess()
	return &ur
}

func (r RateLimiter) RateLimited(apiKey string) bool {
	ur, ok := r[apiKey]

	if ok {
		return ur.rateLimited()
	} else {
		// Add new API to rate limiter
		r[apiKey] = newUserRate()
		return false
	}
}
