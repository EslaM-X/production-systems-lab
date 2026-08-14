package retry

import (
	"errors"
	"testing"
	"time"
)

func TestRetrySucceedsImmediately(t *testing.T) {
	calls := 0
	err := Retry(Config{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetryRetriesUntilSuccess(t *testing.T) {
	calls := 0
	err := Retry(Config{MaxAttempts: 5, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}, func() error {
		calls++
		if calls < 3 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetryExhausts(t *testing.T) {
	calls := 0
	err := Retry(Config{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}, func() error {
		calls++
		return errors.New("always fails")
	})
	if !errors.Is(err, ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected exactly MaxAttempts calls, got %d", calls)
	}
}

func TestRetryStopsOnNonRetryable(t *testing.T) {
	fatal := errors.New("fatal")
	calls := 0
	err := Retry(Config{
		MaxAttempts: 5,
		BaseDelay:   time.Millisecond,
		MaxDelay:    time.Millisecond,
		Retryable:   func(e error) bool { return !errors.Is(e, fatal) },
	}, func() error {
		calls++
		return fatal
	})
	if errors.Is(err, ErrExhausted) {
		t.Fatal("non-retryable must not wrap in ErrExhausted semantics")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call for non-retryable, got %d", calls)
	}
}
