# production-systems-lab

Production-grade building blocks for distributed, financial, and API systems.
Each package is small, dependency-free, tested, and documents the real-world
failure mode it defends against.

Designed and developed by EslaM-X.

## Packages

### Reliability
| Package | Failure mode | Defence |
| --- | --- | --- |
| `reliability/idempotency` | Retries cause double settlement | Idempotency-key store; first write wins, retries return stored result |
| `reliability/replay` | Captured requests replayed later | Monotonic nonce + time-window guard |
| `reliability/circuitbreaker` | Downstream outage causes cascading failure | Three-state circuit breaker (closed/open/half-open) |
| `reliability/retry` | Transient errors abort work | Exponential backoff + jitter, capped attempts, selective retryability |

### Security
| Package | Failure mode | Defence |
| --- | --- | --- |
| `security/auth` | Stolen or forged API keys | HMAC key minting/verification + RBAC permission checks |
| `security/ratelimit` | Abuse and burst overload | Token-bucket rate limiting |
| `security/validation` | Invalid input corrupts state | Luhn checksums, monetary-amount validation, rule combinators |

### Observability
| Package | Failure mode | Defence |
| --- | --- | --- |
| `observability/audit` | Tampered audit trails | Append-only, hash-chained log with integrity verification |

### Webhooks
| Package | Failure mode | Defence |
| --- | --- | --- |
| `webhook` | Forged or replayed webhook deliveries | HMAC-SHA256 signatures, constant-time verification |

## Design principles
- **Zero external dependencies** — the whole lab runs on the standard library.
- **Swap-friendly** — storage (idempotency, audit) is behind interfaces so the
  in-memory implementation can be replaced with Postgres/Redis.
- **Tested** — every package ships with unit tests (`go test ./...`).

## Usage

```go
gw := idempotency.NewGateway(idempotency.NewMemStore(30*time.Minute))
status, body, err := gw.Execute(ctx, idemKey, payload, op)
```

```go
br := circuitbreaker.New(circuitbreaker.DefaultConfig())
if err := br.Execute(downstreamCall); err == circuitbreaker.ErrOpen {
    // fail fast, shed load
}
```

```go
log := audit.New()
log.Append("alice", "payments.approve", "pay_123")
if !log.Verify() {
    // someone tampered with the chain
}
```

## Run

```sh
go test ./...
go test -bench=Benchmark -benchmem ./benchmarks
```

## Benchmarks (Windows 11, i5-1235U, Go 1.26)

| Benchmark | ns/op | allocs/op |
| --- | --- | --- |
| RateLimitAllow | 18.3 | 0 |
| CircuitBreakerExecute | 30.9 | 0 |
| BreakerOpenFailFast | 36.7 | 0 |

Full results: `benchmarks/result.json`.

## Documentation

- [Failure matrix](docs/failure-matrix.md) - every failure mode, its control,
  and the test that proves it (Evidence over claims)
- [Architecture](docs/architecture.md) - flows and state machines of the core controls
- [Methodology](docs/methodology.md)

## License

Apache-2.0. See `LICENSE`.
