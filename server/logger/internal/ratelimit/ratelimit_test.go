package ratelimit

import (
	"testing"
)

func TestRateLimit(t *testing.T) {
	// Create a rate limiter with 10 accesses per minute
	ratelimiter := NewRateLimiter(10)

	expecteds := []bool{false, false, false, false, false, false, false, false, false, false, true, true, true, true, true}

	for i, expected := range expecteds {
		got := ratelimiter.RateLimited("test1")
		if got != expected {
			t.Errorf("%d: got %t, expected %t", i, got, expected)
		}
	}

	for i, expected := range expecteds {
		got := ratelimiter.RateLimited("test2")
		if got != expected {
			t.Errorf("%d: got %t, expected %t", i, got, expected)
		}
	}

	for i, expected := range expecteds {
		got := ratelimiter.RateLimited("test3")
		if got != expected {
			t.Errorf("%d: got %t, expected %t", i, got, expected)
		}
	}
}
