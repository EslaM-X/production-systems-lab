// Package ratelimit implements a token-bucket rate limiter.
//
// A bucket holds up to capacity tokens; each request consumes one; tokens
// refill at rate/sec. Bursty clients are limited smoothly while a healthy
// average rate is preserved.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a token bucket.
type Limiter struct {
	mu       sync.Mutex
	capacity float64
	tokens   float64
	refill   float64 // tokens per second
	last     time.Time
}

// New builds a limiter allowing up to `capacity` burst tokens and refilling
// at `rate` tokens per second.
func New(rate, capacity float64) *Limiter {
	return &Limiter{
		capacity: capacity,
		tokens:   capacity,
		refill:   rate,
		last:     time.Now(),
	}
}

// Allow reports whether a request may proceed (consumes a token).
func (l *Limiter) Allow() bool {
	return l.AllowN(1)
}

// AllowN reports whether n tokens are available (consumes them if so).
func (l *Limiter) AllowN(n float64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.tokens += now.Sub(l.last).Seconds() * l.refill
	if l.tokens > l.capacity {
		l.tokens = l.capacity
	}
	l.last = now
	if l.tokens >= n {
		l.tokens -= n
		return true
	}
	return false
}
