# production-systems-lab

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache-2.0-blue.svg)](LICENSE) [![CI](https://github.com/EslaM-X/production-systems-lab/actions/workflows/ci.yml/badge.svg)](https://github.com/EslaM-X/production-systems-lab/actions) [![ProofX Verified](https://img.shields.io/badge/ProofX-Verified-FFB627?logo=shield&logoColor=white)](https://github.com/EslaM-X/proofx)

Production-grade building blocks for distributed, financial, and API systems.
Small, dependency-free, tested Go packages — each one defends against a specific
real-world failure mode.

> **Designed and developed by [EslaM-X](https://github.com/EslaM-X).** Part of the
> [engineering portfolio](https://github.com/EslaM-X/portfolio).

---

## Why this exists

Systems fail in predictable ways: retries double-settle payments, forged webhooks
replay, an overloaded downstream takes the whole service down, audit trails get
tampered with. This lab is a reference implementation of the controls that stop
each one — with the test that proves it (evidence over claims).

## Demo (real run, ~2 seconds)

One scenario — a payment gateway — exercising **auth → validation → idempotency →
circuit breaker → audit** together. Run it yourself:

```sh
go run ./examples/payment-gateway
```

Actual output:

```
=== production-systems-lab demo: payment gateway ===

auth: key verified for "merchant-42", RBAC payments:create granted
idempotency [first call]: status=200 body={"settled":1999,"charge_id":"pay_123"} executions=1
idempotency [retry  1   ]: status=200 body={"settled":1999,"charge_id":"pay_123"} executions=1
idempotency [retry  2   ]: status=200 body={"settled":1999,"charge_id":"pay_123"} executions=1
idempotency [conflict] : ERROR idempotency key already used with a different payload

circuitbreaker [call 1]: downstream error (psp: timeout contacting acquirer)
circuitbreaker [call 2]: downstream error (psp: timeout contacting acquirer)
circuitbreaker [call 3]: downstream error (psp: timeout contacting acquirer)
circuitbreaker [call 4]: FAIL FAST (circuit open) - downstream not touched
circuitbreaker [call 5]: FAIL FAST (circuit open) - downstream not touched

audit: integrity verified: true
audit: 3 entries (CSV preview)
       2026-08-15T05:43:52+03:00,merchant-42,payments.charge.request,pay_123,29af3d44...
       2026-08-15T05:43:52+03:00,merchant-42,payments.charge.settled,pay_123,110b130d...
```

Notice the three things this lab exists to demonstrate: retries **did not**
re-execute the payment (`executions=1`), the breaker **fail-fast** while the PSP
was down, and the audit trail **verified** intact.

## Quick start

```sh
go test ./...            # every package, its tests
go run ./examples/payment-gateway   # the demo above
```

Requires Go 1.22+. No external dependencies — the whole lab runs on the standard
library.

## Packages

Each package is a control, its failure mode, and its defence:

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

## Evidence

### Benchmarks (Windows 11, i5-1235U, Go 1.26)

| Benchmark | ns/op | allocs/op |
| --- | --- | --- |
| RateLimitAllow | 18.3 | 0 |
| CircuitBreakerExecute | 30.9 | 0 |
| BreakerOpenFailFast | 36.7 | 0 |

Full results: `benchmarks/result.json`.

### Design principles
- **Zero external dependencies** — the whole lab runs on the standard library.
- **Swap-friendly** — storage (idempotency, audit) is behind interfaces so the
  in-memory implementation can be replaced with Postgres/Redis.
- **Tested** — every package ships with unit tests (`go test ./...`).

## Documentation

- [Failure matrix](docs/failure-matrix.md) - every failure mode, its control,
  and the test that proves it (Evidence over claims)
- [Architecture](docs/architecture.md) - flows and state machines of the core controls
- [Methodology](docs/methodology.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Security issues: [SECURITY.md](SECURITY.md).

If this work is useful to you, consider starring the repository — it helps the
project reach more engineers.

## License

Apache-2.0. See `LICENSE`.
