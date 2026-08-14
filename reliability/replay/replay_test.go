package replay

import (
	"errors"
	"testing"
	"time"
)

func TestValidateAcceptsFreshUniqueNonce(t *testing.T) {
	g := NewGuard(DefaultConfig(), time.Hour)
	if err := g.Validate("n1", time.Now()); err != nil {
		t.Fatalf("fresh nonce rejected: %v", err)
	}
}

func TestValidateRejectsDuplicateNonce(t *testing.T) {
	g := NewGuard(DefaultConfig(), time.Hour)
	ts := time.Now()
	if err := g.Validate("n2", ts); err != nil {
		t.Fatalf("first use rejected: %v", err)
	}
	if err := g.Validate("n2", ts); !errors.Is(err, ErrReplay) {
		t.Fatalf("expected ErrReplay, got %v", err)
	}
}

func TestValidateRejectsExpired(t *testing.T) {
	g := NewGuard(Config{MaxAge: time.Minute, MaxClockSkew: time.Second}, time.Hour)
	err := g.Validate("n3", time.Now().Add(-10*time.Minute))
	if !errors.Is(err, ErrReplay) {
		t.Fatalf("expected ErrReplay for old timestamp, got %v", err)
	}
}

func TestValidateRejectsFutureSkew(t *testing.T) {
	g := NewGuard(Config{MaxAge: time.Minute, MaxClockSkew: time.Second}, time.Hour)
	err := g.Validate("n4", time.Now().Add(5*time.Minute))
	if !errors.Is(err, ErrReplay) {
		t.Fatalf("expected ErrReplay for future timestamp, got %v", err)
	}
}

func TestValidateAllowsDifferentNoncesSameTime(t *testing.T) {
	g := NewGuard(DefaultConfig(), time.Hour)
	ts := time.Now()
	if err := g.Validate("a", ts); err != nil {
		t.Fatalf("nonce a rejected: %v", err)
	}
	if err := g.Validate("b", ts); err != nil {
		t.Fatalf("nonce b rejected: %v", err)
	}
}
