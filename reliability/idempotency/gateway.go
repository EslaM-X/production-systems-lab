package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Gateway wraps a Store and applies idempotency semantics to any operation.
//
// Use it like a middleware: hash the request, TryClaim, and if a completed
// result already exists, return it instead of running the operation.
type Gateway struct {
	Store Store
}

// NewGateway builds a Gateway around a Store.
func NewGateway(store Store) *Gateway {
	return &Gateway{Store: store}
}

// Hash returns a stable fingerprint of a request payload.
func Hash(payload string) string {
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

// Execute runs op exactly once per key.
//
//   - First call: claims the key, runs op, marks complete, returns result.
//   - Retry with same key+payload: returns the stored result immediately.
//   - Retry with same key but different payload: returns ErrConflict.
func (g *Gateway) Execute(ctx context.Context, key, payload string, op func() (int, []byte, error)) (int, []byte, error) {
	res, err := g.Store.TryClaim(ctx, key, Hash(payload))
	if err != nil {
		return 0, nil, err
	}
	if res.Complete {
		return res.Status, res.Response, nil
	}
	status, body, err := op()
	if err != nil {
		return 0, nil, fmt.Errorf("operation failed: %w", err)
	}
	if err := g.Store.MarkComplete(ctx, key, status, body); err != nil {
		return 0, nil, err
	}
	return status, body, nil
}
