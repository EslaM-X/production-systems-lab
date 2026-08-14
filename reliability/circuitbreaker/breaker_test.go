package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

func TestClosedAllowsCalls(t *testing.T) {
	b := New(DefaultConfig())
	if !b.Allow() {
		t.Fatal("closed circuit must allow calls")
	}
}

func TestOpensAfterThreshold(t *testing.T) {
	b := New(Config{FailureThreshold: 3, Cooldown: time.Hour, SuccessThreshold: 2})
	for i := 0; i < 3; i++ {
		b.Fail()
	}
	if b.Allow() {
		t.Fatal("circuit must be open after threshold")
	}
	if err := b.Execute(func() error { return nil }); !errors.Is(err, ErrOpen) {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestHalfOpenProbeSucceedsCloses(t *testing.T) {
	b := New(Config{FailureThreshold: 2, Cooldown: time.Nanosecond, SuccessThreshold: 2})
	b.Fail()
	b.Fail()
	// force cooldown to pass
	b.openUntil = time.Now().Add(-time.Second)
	if !b.Allow() {
		t.Fatal("half-open must allow a probe")
	}
	b.Succeed()
	b.Succeed()
	if !b.Allow() {
		t.Fatal("circuit must close after successes")
	}
}

func TestHalfOpenProbeFailsReopens(t *testing.T) {
	b := New(Config{FailureThreshold: 2, Cooldown: time.Hour, SuccessThreshold: 1})
	b.Fail()
	b.Fail()
	b.openUntil = time.Now().Add(-time.Second)
	if !b.Allow() {
		t.Fatal("half-open must allow a probe")
	}
	b.Fail()
	if b.Allow() {
		t.Fatal("circuit must reopen after probe failure")
	}
}
