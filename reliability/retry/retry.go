// Package retry implements retries with exponential backoff and jitter,
// plus a max-attempts cap. Used for transient-failure resilience.
package retry

import (
	"errors"
	"math/rand"
	"time"
)

// Config tunes retry behaviour.
type Config struct {
	// MaxAttempts is the total number of attempts (including the first).
	MaxAttempts int
	// BaseDelay is the initial backoff.
	BaseDelay time.Duration
	// MaxDelay caps each backoff.
	MaxDelay time.Duration
	// Jitter factor in [0,1]; 0 = deterministic backoff.
	Jitter float64
	// Retryable returns whether an error deserves another attempt.
	Retryable func(err error) bool
}

// DefaultConfig retries up to 5 times with capped exponential backoff and
// treats every error as retryable.
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 5,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		Jitter:      0.3,
		Retryable:   func(err error) bool { return err != nil },
	}
}

// ErrExhausted is returned when all attempts fail.
var ErrExhausted = errors.New("retries exhausted")

// Retry runs op, retrying per Config.
//
//   - Returns nil when an attempt succeeds.
//   - Returns the last error unwrapped when it is non-retryable.
//   - Returns the last error wrapped in ErrExhausted when attempts run out.
func Retry(cfg Config, op func() error) error {
	retryable := cfg.Retryable
	if retryable == nil {
		retryable = func(err error) bool { return err != nil }
	}
	var lastErr error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := op(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if !retryable(lastErr) {
			return lastErr
		}
		if attempt == cfg.MaxAttempts-1 {
			break
		}
		delay := backoff(cfg, attempt)
		time.Sleep(delay)
	}
	return errors.Join(ErrExhausted, lastErr)
}

func backoff(cfg Config, attempt int) time.Duration {
	exp := time.Duration(1<<uint(attempt)) * cfg.BaseDelay
	if exp > cfg.MaxDelay {
		exp = cfg.MaxDelay
	}
	if cfg.Jitter > 0 {
		j := rand.Float64()*2 - 1
		exp = time.Duration(float64(exp) * (1 + cfg.Jitter*j))
	}
	return exp
}
