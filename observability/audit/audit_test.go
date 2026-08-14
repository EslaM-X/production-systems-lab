package audit

import "testing"

func TestAppendAndVerify(t *testing.T) {
	l := New()
	l.Append("alice", "payments.approve", "pay-123")
	l.Append("bob", "reports.read", "q3")
	l.Append("carol", "users.update", "user-9")
	if len(l.Entries()) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(l.Entries()))
	}
	if !l.Verify() {
		t.Fatal("chain must verify clean")
	}
}

func TestTamperDetected(t *testing.T) {
	l := New()
	l.Append("alice", "payments.approve", "pay-123")
	l.Append("bob", "reports.read", "q3")
	entries := l.Entries()
	entries[0].Action = "payments.steal"
	if hashEntry(entries[0]) == entries[0].Hash {
		t.Fatal("tampered entry must have a different hash")
	}
}

func TestChainLinks(t *testing.T) {
	l := New()
	l.Append("a", "x", "r1")
	e2 := l.Append("b", "y", "r2")
	if e2.PrevHash != l.Entries()[0].Hash {
		t.Fatal("second entry must chain to the first")
	}
}

func TestEnforceAppendOnly(t *testing.T) {
	l := New()
	l.Append("a", "x", "r1")
	l.Append("b", "y", "r2")
	if !EnforceAppendOnly(l.Entries()) {
		t.Fatal("monotonic log must be accepted")
	}
	bad := []Entry{{Seq: 2, Timestamp: l.Entries()[0].Timestamp}, {Seq: 1, Timestamp: l.Entries()[0].Timestamp}}
	if EnforceAppendOnly(bad) {
		t.Fatal("non-monotonic seq must be rejected")
	}
}
