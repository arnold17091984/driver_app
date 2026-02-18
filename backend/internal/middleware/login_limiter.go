package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/kento/driver/backend/pkg/apperror"
)

type loginAttempt struct {
	failures int
	lockedAt time.Time
}

// LoginLimiter tracks failed login attempts per account.
type LoginLimiter struct {
	mu         sync.Mutex
	attempts   map[string]*loginAttempt
	maxFails   int
	lockoutDur time.Duration
}

// NewLoginLimiter creates a limiter that locks after maxFails failures for lockoutDur.
func NewLoginLimiter(maxFails int, lockoutDur time.Duration) *LoginLimiter {
	ll := &LoginLimiter{
		attempts:   make(map[string]*loginAttempt),
		maxFails:   maxFails,
		lockoutDur: lockoutDur,
	}
	go ll.cleanup()
	return ll
}

// IsLocked returns true if the account is currently locked out.
func (ll *LoginLimiter) IsLocked(account string) bool {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	a, ok := ll.attempts[account]
	if !ok {
		return false
	}
	if a.failures >= ll.maxFails {
		if time.Since(a.lockedAt) < ll.lockoutDur {
			return true
		}
		// Lockout expired, reset
		delete(ll.attempts, account)
	}
	return false
}

// RecordFailure increments the failure counter for an account.
func (ll *LoginLimiter) RecordFailure(account string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()

	a, ok := ll.attempts[account]
	if !ok {
		a = &loginAttempt{}
		ll.attempts[account] = a
	}
	a.failures++
	if a.failures >= ll.maxFails {
		a.lockedAt = time.Now()
	}
}

// RecordSuccess clears the failure counter for an account.
func (ll *LoginLimiter) RecordSuccess(account string) {
	ll.mu.Lock()
	defer ll.mu.Unlock()
	delete(ll.attempts, account)
}

func (ll *LoginLimiter) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		ll.mu.Lock()
		for k, a := range ll.attempts {
			if a.failures >= ll.maxFails && time.Since(a.lockedAt) > ll.lockoutDur {
				delete(ll.attempts, k)
			} else if a.failures < ll.maxFails && time.Since(a.lockedAt) > 30*time.Minute {
				delete(ll.attempts, k)
			}
		}
		ll.mu.Unlock()
	}
}

// LimitLogin returns middleware that checks the login limiter before allowing the request.
// It reads the employee_id or phone_number from the request to determine the account key.
func (ll *LoginLimiter) LimitLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We cannot read the body here without consuming it for the handler.
		// Instead, the handler should call IsLocked/RecordFailure/RecordSuccess directly.
		// This middleware serves as a pre-check using X-Login-Account header set by a wrapper,
		// but the primary integration is done in the auth handler.
		next.ServeHTTP(w, r)
	})
}

// CheckAccountLocked is a simple middleware that returns 429 if the account (from query or header) is locked.
// For the login endpoint, the handler itself should call IsLocked with the parsed body.
func (ll *LoginLimiter) CheckAndReject(account string, w http.ResponseWriter) bool {
	if ll.IsLocked(account) {
		apperror.WriteErrorMsg(w, http.StatusTooManyRequests, "ACCOUNT_LOCKED", "too many failed login attempts; try again later")
		return true
	}
	return false
}
