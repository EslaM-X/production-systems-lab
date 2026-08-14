// Package validation provides input-validation helpers: checksum (Luhn) for
// card-like numbers, amount validation for monetary values, and generic rule
// combinators.
package validation

import (
	"errors"
	"strings"
)

// ErrValidation is a collection of field errors.
type ErrValidation struct {
	Errors []string
}

func (e *ErrValidation) Error() string {
	return "validation failed: " + strings.Join(e.Errors, "; ")
}

// LuhnValid checks a numeric string with the Luhn algorithm (cards, IDs).
func LuhnValid(number string) bool {
	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		c := number[i]
		if c < '0' || c > '9' {
			return false
		}
		d := int(c - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0 && len(number) > 0
}

// Amount is a valid positive monetary amount with at most `decimals` places.
func Amount(amount string, decimals int) bool {
	if amount == "" {
		return false
	}
	parts := strings.Split(amount, ".")
	if len(parts) > 2 {
		return false
	}
	intPart, fracPart := parts[0], ""
	if len(parts) == 2 {
		fracPart = parts[1]
	}
	if intPart == "" && fracPart == "" {
		return false
	}
	for _, c := range intPart {
		if c < '0' || c > '9' {
			return false
		}
	}
	for _, c := range fracPart {
		if c < '0' || c > '9' {
			return false
		}
	}
	if len(fracPart) > decimals {
		return false
	}
	if intPart == "" {
		intPart = "0"
	}
	if intPart != "0" && strings.HasPrefix(intPart, "0") {
		return false // no leading zeros
	}
	return true
}

// Require appends a field error to ErrValidation if cond is false.
func Require(cond bool, e *ErrValidation, msg string) {
	if !cond {
		e.Errors = append(e.Errors, msg)
	}
}

// Check runs the assertions; returns ErrValidation (or nil) — caller inspects.
func Check(f func(e *ErrValidation)) error {
	e := &ErrValidation{}
	f(e)
	if len(e.Errors) > 0 {
		return e
	}
	return nil
}

// IsErrValidation reports whether err is a validation error.
func IsErrValidation(err error) bool {
	var ve *ErrValidation
	return errors.As(err, &ve)
}
