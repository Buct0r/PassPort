package main

import (
	"sync"
	"time"
)

type RateLimiter struct {
	maxAttempts int
	lockoutTime time.Duration
	attempts    int
	lastAttempt time.Time
	lockedUntil time.Time
	mu          sync.Mutex
}

func NewRateLimiter(maxAttempts int, lockoutTime time.Duration) *RateLimiter {
	return &RateLimiter{
		maxAttempts: maxAttempts,
		lockoutTime: lockoutTime,
	}
}

func (rl *RateLimiter) IsLocked() (locked bool, remaining time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if time.Now().Before(rl.lockedUntil) {
		return true, time.Until(rl.lockedUntil)
	}
	return false, 0
}

func (rl *RateLimiter) RecordFailure() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.attempts++
	rl.lastAttempt = time.Now()

	if rl.attempts >= rl.maxAttempts {
		rl.lockedUntil = time.Now().Add(rl.lockoutTime)
		return true // Locked
	}
	return false
}

func (rl *RateLimiter) RecordSuccess() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.attempts = 0
	rl.lockedUntil = time.Time{}
}

func (rl *RateLimiter) GetRemainingAttempts() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return rl.maxAttempts - rl.attempts
}
