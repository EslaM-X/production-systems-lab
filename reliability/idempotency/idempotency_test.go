package idempotency

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecuteFirstCallRunsOp(t *testing.T) {
	store := NewMemStore(time.Minute)
	gw := NewGateway(store)

	calls := 0
	status, body, err := gw.Execute(context.Background(), "key-1", "payload", func() (int, []byte, error) {
		calls++
		return 201, []byte("created"), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != 201 || string(body) != "created" {
		t.Fatalf("unexpected result: %d %s", status, body)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestExecuteRetryReturnsStoredResult(t *testing.T) {
	store := NewMemStore(time.Minute)
	gw := NewGateway(store)

	calls := 0
	op := func() (int, []byte, error) {
		calls++
		return 200, []byte("ok"), nil
	}
	if _, _, err := gw.Execute(context.Background(), "key-2", "payload", op); err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	status, body, err := gw.Execute(context.Background(), "key-2", "payload", op)
	if err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if calls != 1 {
		t.Fatalf("retry must not re-run the op, calls=%d", calls)
	}
	if status != 200 || string(body) != "ok" {
		t.Fatalf("unexpected stored result: %d %s", status, body)
	}
}

func TestExecuteSameKeyDifferentPayloadConflicts(t *testing.T) {
	store := NewMemStore(time.Minute)
	gw := NewGateway(store)

	if _, _, err := gw.Execute(context.Background(), "key-3", "payload-a", func() (int, []byte, error) {
		return 200, nil, nil
	}); err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	_, _, err := gw.Execute(context.Background(), "key-3", "payload-b", func() (int, []byte, error) {
		return 200, nil, nil
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestHashDeterministic(t *testing.T) {
	if Hash("same") != Hash("same") {
		t.Fatal("hash must be deterministic")
	}
	if Hash("same") == Hash("different") {
		t.Fatal("hash must differ across payloads")
	}
}

func TestCleanupRemovesExpired(t *testing.T) {
	store := NewMemStore(time.Nanosecond)
	// force an "old" entry via a tiny sleep
	store.items["old"] = &Result{Key: "old", CreatedAt: time.Now().Add(-time.Hour)}
	store.claims["old"] = "hash"
	store.Cleanup()
	if _, ok := store.items["old"]; ok {
		t.Fatal("expired entry must be removed")
	}
}
