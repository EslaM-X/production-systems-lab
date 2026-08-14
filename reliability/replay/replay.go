// Package replay protects against replay attacks: a captured request being
// re-sent later to trigger a second (unwanted) settlement or action.
//
// The classic defence is a monotonic nonce/timestamp pair: reject any request
// whose timestamp is too old (expiry window) or whose nonce was already seen.
package replay

import (
	"errors"
	"sync"
	"time"
)

// ErrReplay is returned when a request is detected as a replay.
var ErrReplay = errors.New("request replayed or expired")

// Config controls the replay window.
type Config struct {
	// MaxAge is how far in the past a timestamp may be.
	MaxAge time.Duration
	// MaxClockSkew is how far in the future we tolerate (NTP drift).
	MaxClockSkew time.Duration
}

// DefaultConfig is a sane default: 5 minutes back, 1 minute forward.
func DefaultConfig() Config {
	return Config{MaxAge: 5 * time.Minute, MaxClockSkew: time.Minute}
}

// Guard rejects replays using a monotonic nonce registry + time window.
type Guard struct {
	mu      sync.Mutex
	cfg     Config
	seen    map[string]bool
	cleanup time.Duration
	clock   func() time.Time
}

// NewGuard builds a Guard. nonces are retained for cleanupInterval.
func NewGuard(cfg Config, cleanupInterval time.Duration) *Guard {
	if cleanupInterval <= 0 {
		cleanupInterval = time.Hour
	}
	return &Guard{
		cfg:     cfg,
		seen:    map[string]bool{},
		cleanup: cleanupInterval,
		clock:   time.Now,
	}
}

// Validate checks a (nonce, timestamp) pair. Returns ErrReplay on rejection.
func (g *Guard) Validate(nonce string, timestamp time.Time) error {
	now := g.clock()
	if timestamp.After(now.Add(g.cfg.MaxClockSkew)) {
		return ErrReplay
	}
	if now.Sub(timestamp) > g.cfg.MaxAge {
		return ErrReplay
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.seen[nonce] {
		return ErrReplay
	}
	g.seen[nonce] = true
	return nil
}
