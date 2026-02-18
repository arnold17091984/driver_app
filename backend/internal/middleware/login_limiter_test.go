package middleware

import (
	"testing"
	"time"
)

func TestLoginLimiter_AllowsUnderLimit(t *testing.T) {
	ll := NewLoginLimiter(5, 15*time.Minute)

	for i := 0; i < 4; i++ {
		ll.RecordFailure("user@test.com")
	}

	if ll.IsLocked("user@test.com") {
		t.Error("account should not be locked after 4 failures (limit is 5)")
	}
}

func TestLoginLimiter_LocksAfterMaxFailures(t *testing.T) {
	ll := NewLoginLimiter(5, 15*time.Minute)

	for i := 0; i < 5; i++ {
		ll.RecordFailure("user@test.com")
	}

	if !ll.IsLocked("user@test.com") {
		t.Error("account should be locked after 5 failures")
	}
}

func TestLoginLimiter_UnlocksAfterDuration(t *testing.T) {
	ll := NewLoginLimiter(3, 50*time.Millisecond)

	for i := 0; i < 3; i++ {
		ll.RecordFailure("user@test.com")
	}

	if !ll.IsLocked("user@test.com") {
		t.Error("account should be locked immediately after max failures")
	}

	time.Sleep(60 * time.Millisecond)

	if ll.IsLocked("user@test.com") {
		t.Error("account should be unlocked after lockout duration")
	}
}

func TestLoginLimiter_SuccessResetsCounter(t *testing.T) {
	ll := NewLoginLimiter(5, 15*time.Minute)

	for i := 0; i < 4; i++ {
		ll.RecordFailure("user@test.com")
	}

	ll.RecordSuccess("user@test.com")

	// Should be reset, so 4 more failures should not lock
	for i := 0; i < 4; i++ {
		ll.RecordFailure("user@test.com")
	}

	if ll.IsLocked("user@test.com") {
		t.Error("account should not be locked after success reset + 4 failures")
	}
}

func TestLoginLimiter_DifferentAccountsAreIndependent(t *testing.T) {
	ll := NewLoginLimiter(3, 15*time.Minute)

	for i := 0; i < 3; i++ {
		ll.RecordFailure("user1")
	}

	if !ll.IsLocked("user1") {
		t.Error("user1 should be locked")
	}
	if ll.IsLocked("user2") {
		t.Error("user2 should not be locked")
	}
}

func TestLoginLimiter_UnknownAccountIsNotLocked(t *testing.T) {
	ll := NewLoginLimiter(5, 15*time.Minute)

	if ll.IsLocked("unknown") {
		t.Error("unknown account should not be locked")
	}
}
