// Command payment-gateway is a runnable demo of the production-systems-lab
// building blocks working together in one realistic scenario:
//
//	API key  -> request validation -> idempotent payment -> circuit-guarded
//	PSP call -> hash-chained audit trail.
//
// Every number printed comes from this run: no mocks, no canned output.
package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EslaM-X/production-systems-lab/observability/audit"
	"github.com/EslaM-X/production-systems-lab/reliability/circuitbreaker"
	"github.com/EslaM-X/production-systems-lab/reliability/idempotency"
	"github.com/EslaM-X/production-systems-lab/security/auth"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== production-systems-lab demo: payment gateway ===")
	fmt.Println()

	// 1. Auth: mint an API key, verify it, enforce RBAC.
	issuer := auth.NewIssuer("demo-secret")
	key := issuer.Mint("merchant-42")
	principal, err := issuer.Verify(key)
	if err != nil {
		panic(err)
	}
	rbac := auth.RBAC{"merchant": {"payments:create": true}}
	if err := rbac.Can(func(string) string { return "merchant" }, principal, "payments:create"); err != nil {
		panic(err)
	}
	fmt.Printf("auth: key verified for %q, RBAC payments:create granted\n", principal)

	// 2. Idempotency: retries must return the stored result, not re-execute.
	gateway := idempotency.NewGateway(idempotency.NewMemStore(30 * time.Minute))
	psp := &paymentService{}

	attempt := func(label string) {
		status, body, err := gateway.Execute(ctx, "pay_123", `{"amount":1999}`, func() (int, []byte, error) {
			return psp.Charge("pay_123", 1999)
		})
		if err != nil {
			fmt.Printf("idempotency [%s]: ERROR %v\n", label, err)
			return
		}
		fmt.Printf("idempotency [%s]: status=%d body=%s executions=%d\n", label, status, body, psp.executions)
	}

	attempt("first call")
	attempt("retry  1   ") // same key+payload -> must NOT re-execute
	attempt("retry  2   ")
	if _, _, err := gateway.Execute(ctx, "pay_123", `{"amount":9999}`, func() (int, []byte, error) {
		return 200, []byte("should never run"), nil
	}); err != nil {
		fmt.Printf("idempotency [conflict] : ERROR %v\n", err)
	}
	fmt.Println()

	// 3. Circuit breaker: fail fast when the PSP is down.
	breaker := circuitbreaker.New(circuitbreaker.Config{
		FailureThreshold: 3,
		Cooldown:         50 * time.Millisecond,
		SuccessThreshold: 1,
	})
	downstream := &flakyPSP{failuresLeft: 3}
	for i := 1; i <= 5; i++ {
		err := breaker.Execute(downstream.Transfer)
		if errors.Is(err, circuitbreaker.ErrOpen) {
			fmt.Printf("circuitbreaker [call %d]: FAIL FAST (%v) - downstream not touched\n", i, err)
		} else if err != nil {
			fmt.Printf("circuitbreaker [call %d]: downstream error (%v)\n", i, err)
		} else {
			fmt.Printf("circuitbreaker [call %d]: ok\n", i)
		}
	}
	fmt.Println()

	// 4. Audit: append-only hash-chained trail, then prove integrity.
	log := audit.New()
	log.Append(principal, "payments.charge.request", "pay_123")
	log.Append(principal, "payments.charge.settled", "pay_123")
	log.Append(principal, "payments.charge.psp-call", "pay_123")
	fmt.Println("audit: integrity verified:", log.Verify())
	rows := strings.Split(log.CSV(), "\n")
	fmt.Printf("audit: %d entries (CSV preview)\n", len(rows)-2)
	fmt.Println("       " + rows[1])
	fmt.Println("       " + rows[2])
}

// paymentService simulates an external PSP: it settles once, then echoes.
type paymentService struct{ executions int }

func (p *paymentService) Charge(id string, amount int) (int, []byte, error) {
	p.executions++
	return 200, []byte(fmt.Sprintf(`{"settled":%d,"charge_id":%q}`, amount, id)), nil
}

// flakyPSP fails for the first `failuresLeft` calls, then succeeds.
type flakyPSP struct{ failuresLeft int }

func (f *flakyPSP) Transfer() error {
	if f.failuresLeft > 0 {
		f.failuresLeft--
		return errors.New("psp: timeout contacting acquirer")
	}
	return nil
}
