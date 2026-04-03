package child

import (
	"sync"
	"time"
)

// attempt tracks PIN attempt count and lockout state for a single student.
type attempt struct {
	count       int
	lockedUntil time.Time
}

// PINRateLimiter enforces a maximum number of PIN attempts before locking the account.
type PINRateLimiter struct {
	mu              sync.Mutex
	attempts        map[string]*attempt
	maxAttempts     int
	lockoutDuration time.Duration
}

// NewPINRateLimiter creates a PINRateLimiter with the given limits.
// maxAttempts: number of allowed attempts before lockout.
// lockoutDuration: how long the account is locked after reaching max attempts.
func NewPINRateLimiter(maxAttempts int, lockoutDuration time.Duration) *PINRateLimiter {
	return &PINRateLimiter{
		attempts:        make(map[string]*attempt),
		maxAttempts:     maxAttempts,
		lockoutDuration: lockoutDuration,
	}
}

// Check records an attempt for studentID and returns whether the attempt is allowed
// and how many attempts remain. Thread-safe.
//
// Logic:
//   - If locked and lockout not yet expired: return false, 0
//   - If locked and lockout expired: reset the counter, allow
//   - Increment count
//   - If count >= maxAttempts: lock, return false, 0
//   - Otherwise: return true, maxAttempts - count
func (rl *PINRateLimiter) Check(studentID string) (allowed bool, remaining int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	a, ok := rl.attempts[studentID]
	if !ok {
		a = &attempt{}
		rl.attempts[studentID] = a
	}

	// Check if currently locked
	if !a.lockedUntil.IsZero() {
		if time.Now().Before(a.lockedUntil) {
			// Still locked
			return false, 0
		}
		// Lockout expired — reset
		a.count = 0
		a.lockedUntil = time.Time{}
	}

	// Increment count
	a.count++

	// Check if we just hit the limit
	if a.count >= rl.maxAttempts {
		a.lockedUntil = time.Now().Add(rl.lockoutDuration)
		return false, 0
	}

	return true, rl.maxAttempts - a.count
}

// Reset clears the attempt record for studentID. Should be called after a successful login.
func (rl *PINRateLimiter) Reset(studentID string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, studentID)
}
