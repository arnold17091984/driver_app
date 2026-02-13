package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/kento/driver/backend/pkg/apperror"
)

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// RateLimiter implements a token-bucket rate limiter per IP address.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // tokens per second
	burst    int     // max tokens (bucket size)
}

// NewRateLimiter creates a rate limiter. rate is requests/second, burst is the
// max number of requests allowed in a burst.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{tokens: float64(rl.burst) - 1, lastSeen: time.Now()}
		return true
	}

	elapsed := time.Since(v.lastSeen).Seconds()
	v.lastSeen = time.Now()
	v.tokens += elapsed * rl.rate
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}

	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// cleanup removes stale entries every 5 minutes.
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit returns an HTTP middleware that rate limits requests per IP.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(ip); err == nil {
			ip = host
		}

		if !rl.allow(ip) {
			w.Header().Set("Retry-After", "1")
			apperror.WriteErrorMsg(w, http.StatusTooManyRequests, "RATE_LIMITED", "too many requests")
			return
		}

		next.ServeHTTP(w, r)
	})
}
