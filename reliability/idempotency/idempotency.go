// Package idempotency provides an idempotency-key store.
//
// The core problem: clients retry. If a payment or a write is processed twice,
// the system double-settles or double-applies. An idempotency key lets the
// first attempt win and makes every retry return the stored result instead of
// executing again.
package idempotency

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrConflict is returned when a request arrives with a key that is already
// bound to a different request payload.
var ErrConflict = errors.New("idempotency key already used with a different payload")

// Result is what the store remembers about a completed (or in-flight) request.
type Result struct {
	Key       string
	Response  []byte
	Status    int
	CreatedAt time.Time
	Complete  bool
}

// Store is the storage contract. Swap in Postgres/Redis by implementing it.
type Store interface {
	// TryClaim atomically claims the key for the given payload hash.
	// Returns the existing result if already claimed.
	TryClaim(ctx context.Context, key, payloadHash string) (*Result, error)
	// MarkComplete stores the final response for the key.
	MarkComplete(ctx context.Context, key string, status int, response []byte) error
	// Get returns the stored result for the key, or nil.
	Get(ctx context.Context, key string) (*Result, error)
}

// MemStore is a thread-safe in-memory Store for development and tests.
type MemStore struct {
	mu     sync.RWMutex
	items  map[string]*Result
	claims map[string]string // key -> payloadHash
	ttl    time.Duration
	clock  func() time.Time
}

// NewMemStore builds an in-memory store with a TTL for stale cleanup.
func NewMemStore(ttl time.Duration) *MemStore {
	return &MemStore{
		items:  map[string]*Result{},
		claims: map[string]string{},
		ttl:    ttl,
		clock:  time.Now,
	}
}

// TryClaim implements Store.
func (m *MemStore) TryClaim(ctx context.Context, key, payloadHash string) (*Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if hash, ok := m.claims[key]; ok {
		if hash != payloadHash {
			return nil, ErrConflict
		}
		if res, ok := m.items[key]; ok {
			return res, nil
		}
		return &Result{Key: key, CreatedAt: m.clock()}, nil
	}
	m.claims[key] = payloadHash
	return &Result{Key: key, CreatedAt: m.clock()}, nil
}

// MarkComplete implements Store.
func (m *MemStore) MarkComplete(ctx context.Context, key string, status int, response []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = &Result{
		Key:       key,
		Response:  response,
		Status:    status,
		CreatedAt: m.clock(),
		Complete:  true,
	}
	return nil
}

// Get implements Store.
func (m *MemStore) Get(ctx context.Context, key string) (*Result, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if res, ok := m.items[key]; ok {
		return res, nil
	}
	return nil, nil
}

// Cleanup removes expired entries. Call on a timer in production.
func (m *MemStore) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := m.clock().Add(-m.ttl)
	for k, res := range m.items {
		if res.CreatedAt.Before(cutoff) {
			delete(m.items, k)
			delete(m.claims, k)
		}
	}
}
