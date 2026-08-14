package auth

import (
	"errors"
	"testing"
)

func TestMintVerifyRoundTrip(t *testing.T) {
	iss := NewIssuer("s3cr3t")
	key := iss.Mint("alice")
	principal, err := iss.Verify(key)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if principal != "alice" {
		t.Fatalf("expected alice, got %s", principal)
	}
}

func TestVerifyRejectsTamperedKey(t *testing.T) {
	iss := NewIssuer("s3cr3t")
	key := iss.Mint("alice")
	tampered := key[:len(key)-2] + "zz"
	if _, err := iss.Verify(tampered); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	iss := NewIssuer("s3cr3t")
	key := iss.Mint("alice")
	other := NewIssuer("d1ff3r3nt")
	if _, err := other.Verify(key); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("expected ErrInvalidKey, got %v", err)
	}
}

func TestRBACAllowsAndDenies(t *testing.T) {
	rbac := RBAC{
		"admin":  {"payments.approve": true, "reports.read": true},
		"viewer": {"reports.read": true},
	}
	roleOf := func(p string) string { return p }
	if err := rbac.Can(roleOf, "admin", "payments.approve"); err != nil {
		t.Fatalf("admin should approve: %v", err)
	}
	if err := rbac.Can(roleOf, "viewer", "payments.approve"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer must be forbidden, got %v", err)
	}
	if err := rbac.Can(roleOf, "viewer", "reports.read"); err != nil {
		t.Fatalf("viewer should read reports: %v", err)
	}
}
