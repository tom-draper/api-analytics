package lib

import "time"

type RateLimiter map[string]UserRate

const accessesPerMinute int = 10

type UserRate struct {
	timestamps [accessesPerMinute]time.Time // Circular array
	current    int                          // Current timestamp index
}

func (u *UserRate) increment() {
	if u.current < len(u.timestamps)-1 {
		u.current += 1
	} else {
		u.current = 0
	}
}

func (u *UserRate) RateLimited() bool {
	oldest := u.timestamps[u.current]
	// If the oldest timestamp recorded is less than a minute ago -> rate limited
	if time.Since(oldest) < time.Minute {
		// Rate limited, user access denied, do not record access
		return true
	} else {
		u.RecordAccess() // Register user access and record current time
		return false
	}
}

func (u *UserRate) RecordAccess() {
	u.increment()
	u.timestamps[u.current] = time.Now()
}

func (r RateLimiter) RateLimited(apiKey string) bool {
	userRate, ok := r[apiKey]

	if ok {
		return userRate.RateLimited()
	} else {
		// Add new API to rate limiter
		userRate := UserRate{}
		userRate.RecordAccess()
		r[apiKey] = userRate
		return false
	}
}
