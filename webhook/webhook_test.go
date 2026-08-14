package webhook

import (
	"errors"
	"testing"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	secret := []byte("whsec_prod_123")
	payload := []byte(`{"event":"payment.completed","id":"pay_1"}`)
	sig := Sign(secret, payload)
	if err := Verify(secret, payload, []byte(sig)); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	payload := []byte("data")
	sig := Sign([]byte("secret-a"), payload)
	if err := Verify([]byte("secret-b"), payload, []byte(sig)); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestVerifyRejectsTamperedPayload(t *testing.T) {
	secret := []byte("whsec")
	payload := []byte("data")
	sig := Sign(secret, payload)
	tampered := []byte("datb")
	if err := Verify(secret, tampered, []byte(sig)); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}

func TestVerifyRejectsGarbageSignature(t *testing.T) {
	if err := Verify([]byte("s"), []byte("p"), []byte("not-hex!!")); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}
