// Package auth provides HMAC-based API key verification and a simple RBAC
// permission checker. Keys are derived from a server secret using a key id,
// so key rotation is possible without touching stored secrets.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// ErrInvalidKey is returned when the presented key is not valid for any known id.
var ErrInvalidKey = errors.New("invalid API key")

// ErrForbidden is returned when a principal lacks the required permission.
var ErrForbidden = errors.New("forbidden")

// Issuer mints and verifies HMAC API keys.
type Issuer struct {
	// secretKey -> id allows a small set of active secrets (rotation).
	secrets map[string]string // id -> secret
}

// NewIssuer builds an Issuer with one secret.
func NewIssuer(secret string) *Issuer {
	return &Issuer{secrets: map[string]string{"1": secret}}
}

// AddSecret adds another active secret under an id (for rotation).
func (i *Issuer) AddSecret(id, secret string) {
	i.secrets[id] = secret
}

// Mint creates a key for a principal. Format: <id>.<principal>.<hmac>
func (i *Issuer) Mint(principal string) string {
	id, secret := i.primary()
	payload := id + "." + principal
	return payload + "." + i.sign(payload, secret)
}

func (i *Issuer) primary() (string, string) {
	for id, secret := range i.secrets {
		return id, secret
	}
	return "", ""
}

func (i *Issuer) sign(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks a key and returns the principal.
func (i *Issuer) Verify(key string) (string, error) {
	parts := splitN(key, ".", 3)
	if len(parts) != 3 {
		return "", ErrInvalidKey
	}
	id, principal, sig := parts[0], parts[1], parts[2]
	secret, ok := i.secrets[id]
	if !ok {
		return "", ErrInvalidKey
	}
	expected := i.sign(id+"."+principal, secret)
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return "", ErrInvalidKey
	}
	return principal, nil
}

func splitN(s, sep string, n int) []string {
	out := []string{}
	acc := ""
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			out = append(out, acc)
			acc = ""
			if len(out) == n-1 {
				out = append(out, s[i+1:])
				return out
			}
			continue
		}
		acc += string(s[i])
	}
	out = append(out, acc)
	return out
}

// RBAC is a permission map: role -> allowed permissions.
type RBAC map[string]map[string]bool

// Can checks whether principal has permission given their role mapping.
// roleOf returns the role for a principal.
func (r RBAC) Can(roleOf func(principal string) string, principal, permission string) error {
	role := roleOf(principal)
	if r[role] != nil && r[role][permission] {
		return nil
	}
	return ErrForbidden
}
