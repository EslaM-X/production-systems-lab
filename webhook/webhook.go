// Package webhook implements HMAC-SHA256 signature verification for incoming
// webhooks, protecting against forged or replayed deliveries.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// ErrInvalidSignature is returned when the signature does not match.
var ErrInvalidSignature = errors.New("invalid webhook signature")

// Verify checks that `sig` matches the HMAC-SHA256 of `payload` under `secret`.
// It uses constant-time comparison to avoid timing attacks.
func Verify(secret, payload, sig []byte) error {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	expected := mac.Sum(nil)
	provided, err := hex.DecodeString(string(sig))
	if err != nil {
		return ErrInvalidSignature
	}
	if !hmac.Equal(expected, provided) {
		return ErrInvalidSignature
	}
	return nil
}

// Sign computes the hex HMAC-SHA256 signature for a payload.
func Sign(secret, payload []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
