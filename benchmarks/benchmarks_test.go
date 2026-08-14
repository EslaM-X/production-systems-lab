package benchmarks

import (
	"fmt"
	"testing"

	"github.com/EslaM-X/production-systems-lab/reliability/circuitbreaker"
	"github.com/EslaM-X/production-systems-lab/security/ratelimit"
)

func BenchmarkRateLimitAllow(b *testing.B) {
	l := ratelimit.New(1e9, 1e9)
	for i := 0; i < b.N; i++ {
		l.Allow()
	}
}

func BenchmarkCircuitBreakerExecute(b *testing.B) {
	br := circuitbreaker.New(circuitbreaker.DefaultConfig())
	for i := 0; i < b.N; i++ {
		_ = br.Execute(func() error { return nil })
	}
}

func BenchmarkBreakerOpenFailFast(b *testing.B) {
	br := circuitbreaker.New(circuitbreaker.Config{FailureThreshold: 1, Cooldown: 0, SuccessThreshold: 1})
	br.Fail() // open it
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = br.Execute(func() error { return nil })
	}
}

func ExampleBreaker() {
	br := circuitbreaker.New(circuitbreaker.DefaultConfig())
	for i := 0; i < 6; i++ {
		err := br.Execute(func() error {
			return fmt.Errorf("downstream failure")
		})
		if err == circuitbreaker.ErrOpen {
			fmt.Println("fail-fast: circuit open")
			break
		}
	}
	// Output: fail-fast: circuit open
}
