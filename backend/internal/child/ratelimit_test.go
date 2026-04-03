package child

import (
	"testing"
	"time"
)

func TestRateLimiterFirstAttemptAllowed(t *testing.T) {
	rl := NewPINRateLimiter(5, 15*time.Minute)
	allowed, remaining := rl.Check("student-1")
	if !allowed {
		t.Error("first attempt should be allowed")
	}
	if remaining != 4 {
		t.Errorf("remaining = %d, want 4", remaining)
	}
}

func TestRateLimiterLocksAfterMaxAttempts(t *testing.T) {
	rl := NewPINRateLimiter(5, 15*time.Minute)
	// Exhaust all 5 attempts
	for i := 0; i < 4; i++ {
		allowed, _ := rl.Check("student-1")
		if !allowed {
			t.Errorf("attempt %d should be allowed", i+1)
		}
	}
	// 5th attempt triggers lockout
	allowed, remaining := rl.Check("student-1")
	if allowed {
		t.Error("5th attempt should be denied (lockout triggered)")
	}
	if remaining != 0 {
		t.Errorf("remaining = %d, want 0", remaining)
	}
}

func TestRateLimiterLockedDuringLockout(t *testing.T) {
	rl := NewPINRateLimiter(5, 15*time.Minute)
	// Trigger lockout
	for i := 0; i < 5; i++ {
		rl.Check("student-1")
	}
	// Should still be locked
	allowed, remaining := rl.Check("student-1")
	if allowed {
		t.Error("should remain locked during lockout period")
	}
	if remaining != 0 {
		t.Errorf("remaining = %d, want 0", remaining)
	}
}

func TestRateLimiterResetsAfterLockoutExpires(t *testing.T) {
	// Use a very short lockout for testing
	rl := NewPINRateLimiter(5, 50*time.Millisecond)
	// Trigger lockout
	for i := 0; i < 5; i++ {
		rl.Check("student-1")
	}
	// Wait for lockout to expire
	time.Sleep(100 * time.Millisecond)
	// Should be allowed again and reset
	allowed, remaining := rl.Check("student-1")
	if !allowed {
		t.Error("should be allowed after lockout expires")
	}
	if remaining != 4 {
		t.Errorf("remaining = %d, want 4 (reset after expiry)", remaining)
	}
}

func TestRateLimiterResetClearsAttempts(t *testing.T) {
	rl := NewPINRateLimiter(5, 15*time.Minute)
	// Use up some attempts
	rl.Check("student-1")
	rl.Check("student-1")
	// Reset (successful login)
	rl.Reset("student-1")
	// Should be back to full attempts
	allowed, remaining := rl.Check("student-1")
	if !allowed {
		t.Error("after Reset, first attempt should be allowed")
	}
	if remaining != 4 {
		t.Errorf("remaining = %d, want 4 after Reset", remaining)
	}
}

func TestRateLimiterIndependentCounters(t *testing.T) {
	rl := NewPINRateLimiter(5, 15*time.Minute)
	// Exhaust student-1
	for i := 0; i < 5; i++ {
		rl.Check("student-1")
	}
	// student-2 should be unaffected
	allowed, remaining := rl.Check("student-2")
	if !allowed {
		t.Error("student-2 should be unaffected by student-1 lockout")
	}
	if remaining != 4 {
		t.Errorf("student-2 remaining = %d, want 4", remaining)
	}
}
