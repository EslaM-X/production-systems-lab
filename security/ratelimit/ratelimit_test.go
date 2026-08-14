package ratelimit

import "testing"

func TestAllowConsumesTokens(t *testing.T) {
	l := New(100, 2)
	if !l.Allow() || !l.Allow() {
		t.Fatal("burst of 2 should be allowed")
	}
	if l.Allow() {
		t.Fatal("bucket exhausted, third call must be denied")
	}
}

func TestAllowN(t *testing.T) {
	l := New(100, 10)
	if !l.AllowN(10) {
		t.Fatal("10 tokens should be available")
	}
	if l.AllowN(1) {
		t.Fatal("no tokens left")
	}
}

func TestRefill(t *testing.T) {
	l := New(100, 1)
	if !l.Allow() {
		t.Fatal("initial token missing")
	}
	if l.Allow() {
		t.Fatal("should be empty")
	}
}
