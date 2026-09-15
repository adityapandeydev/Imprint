package security

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements an in-memory sliding-window token limiter.
type RateLimiter struct {
	mu      sync.Mutex
	records map[string][]time.Time
}

// NewRateLimiter creates a new RateLimiter instance and starts background eviction.
func NewRateLimiter(cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		records: make(map[string][]time.Time),
	}

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		for range ticker.C {
			rl.cleanup(time.Hour)
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for key, timestamps := range rl.records {
		var valid []time.Time
		for _, t := range timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.records, key)
		} else {
			rl.records[key] = valid
		}
	}
}

// Allow evaluates whether an action for key is within rate limits.
// Returns (allowed, retryAfter).
func (rl *RateLimiter) Allow(key string, limit int, window time.Duration) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	timestamps := rl.records[key]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= limit {
		// Oldest timestamp determines when the next slot frees up
		oldest := valid[0]
		retryAfter := window - now.Sub(oldest)
		if retryAfter < time.Second {
			retryAfter = time.Second
		}
		rl.records[key] = valid
		return false, retryAfter
	}

	valid = append(valid, now)
	rl.records[key] = valid
	return true, 0
}

// ExtractIP retrieves the real client IP from standard proxy headers or remote address.
func ExtractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// Middleware creates a Chi-compatible HTTP middleware enforcing the specified rate limit.
func (rl *RateLimiter) Middleware(limit int, window time.Duration, bucketName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ExtractIP(r)
			key := fmt.Sprintf("%s:%s", bucketName, ip)

			allowed, retryAfter := rl.Allow(key, limit, window)
			if !allowed {
				retrySeconds := int(retryAfter.Seconds())
				if retrySeconds <= 0 {
					retrySeconds = 1
				}

				w.Header().Set("Retry-After", fmt.Sprintf("%d", retrySeconds))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error":       "Too many requests. Please slow down.",
					"retry_after": retrySeconds,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
