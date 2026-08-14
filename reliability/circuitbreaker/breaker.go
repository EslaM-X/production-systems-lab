// Package circuitbreaker implements a classic three-state circuit breaker.
//
// States:
//
//	CLOSED   — requests flow; failures accumulate.
//	OPEN     — failures exceeded threshold; requests fail fast (no downstream hit).
//	HALF_OPEN — after a cooldown, a probe request tests recovery.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned while the circuit is open (fast-fail).
var ErrOpen = errors.New("circuit open")

// Config tunes breaker behaviour.
type Config struct {
	// FailureThreshold opens the circuit after N consecutive failures.
	FailureThreshold int
	// Cooldown is how long the circuit stays open before a probe.
	Cooldown time.Duration
	// SuccessThreshold closes the circuit after N consecutive probe successes.
	SuccessThreshold int
}

// DefaultConfig returns a sane starting point.
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 5,
		Cooldown:         10 * time.Second,
		SuccessThreshold: 2,
	}
}

type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

// Breaker guards a downstream call.
type Breaker struct {
	mu        sync.Mutex
	cfg       Config
	cur       state
	fail      int
	succ      int
	openUntil time.Time
	clock     func() time.Time
}

// New builds a Breaker.
func New(cfg Config) *Breaker {
	return &Breaker{cfg: cfg, cur: stateClosed, clock: time.Now}
}

// Allow reports whether a call may proceed now.
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cur == stateOpen {
		if b.clock().After(b.openUntil) {
			b.cur = stateHalfOpen
			return true
		}
		return false
	}
	return true
}

// Succeed records a success and possibly closes the circuit.
func (b *Breaker) Succeed() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cur == stateHalfOpen {
		b.succ++
		if b.succ >= b.cfg.SuccessThreshold {
			b.cur = stateClosed
			b.succ = 0
		}
	} else {
		b.fail = 0
	}
}

// Fail records a failure and possibly opens the circuit.
func (b *Breaker) Fail() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cur == stateHalfOpen {
		b.openUntil = b.clock().Add(b.cfg.Cooldown)
		b.cur = stateOpen
		return
	}
	b.fail++
	if b.fail >= b.cfg.FailureThreshold {
		b.openUntil = b.clock().Add(b.cfg.Cooldown)
		b.cur = stateOpen
	}
}

// Execute runs op through the breaker.
func (b *Breaker) Execute(op func() error) error {
	if !b.Allow() {
		return ErrOpen
	}
	if err := op(); err != nil {
		b.Fail()
		return err
	}
	b.Succeed()
	return nil
}
