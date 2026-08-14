package validation

import "testing"

func TestLuhnValid(t *testing.T) {
	// Visa test numbers
	cases := map[string]bool{
		"4111111111111111": true,
		"4242424242424242": true,
		"4111111111111112": false,
		"1234":             false,
	}
	for num, want := range cases {
		if got := LuhnValid(num); got != want {
			t.Fatalf("LuhnValid(%s) = %v, want %v", num, got, want)
		}
	}
}

func TestAmount(t *testing.T) {
	cases := map[string]bool{
		"0.50":   true,
		"10":     true,
		"10.5":   true,
		"10.55":  true,
		"10.555": false, // 3 decimals, max 2
		"01.00":  false, // leading zero
		"":       false,
		"abc":    false,
		"1.2.3":  false,
		"0":      true,
	}
	for a, want := range cases {
		if got := Amount(a, 2); got != want {
			t.Fatalf("Amount(%q,2) = %v, want %v", a, got, want)
		}
	}
}

func TestCheckCollectsAllErrors(t *testing.T) {
	err := Check(func(e *ErrValidation) {
		Require(false, e, "field a is required")
		Require(false, e, "field b is invalid")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if len(err.(*ErrValidation).Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(err.(*ErrValidation).Errors))
	}
	if !IsErrValidation(err) {
		t.Fatal("IsErrValidation must detect it")
	}
}

func TestCheckPasses(t *testing.T) {
	err := Check(func(e *ErrValidation) {
		Require(true, e, "nope")
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
