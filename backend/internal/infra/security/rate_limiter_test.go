package security

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(time.Minute)

	key := "test-client-ip"
	limit := 3
	window := 100 * time.Millisecond

	// 1st request -> allowed
	allowed, _ := rl.Allow(key, limit, window)
	if !allowed {
		t.Errorf("expected 1st request to be allowed")
	}

	// 2nd request -> allowed
	allowed, _ = rl.Allow(key, limit, window)
	if !allowed {
		t.Errorf("expected 2nd request to be allowed")
	}

	// 3rd request -> allowed
	allowed, _ = rl.Allow(key, limit, window)
	if !allowed {
		t.Errorf("expected 3rd request to be allowed")
	}

	// 4th request -> blocked! (429)
	allowed, retryAfter := rl.Allow(key, limit, window)
	if allowed {
		t.Errorf("expected 4th request within window to be blocked")
	}
	if retryAfter <= 0 {
		t.Errorf("expected positive retryAfter duration")
	}

	// Wait for window to slide
	time.Sleep(120 * time.Millisecond)

	// 5th request -> allowed again
	allowed, _ = rl.Allow(key, limit, window)
	if !allowed {
		t.Errorf("expected request after window expiration to be allowed")
	}
}
