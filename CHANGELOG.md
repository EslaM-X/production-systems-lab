# Changelog

All notable changes to `production-systems-lab`.

## [Unreleased]

### Added
- `docs/failure-matrix.md`: every failure mode mapped to its control and the
  unit test that proves it (Failure -> Control -> Evidence), with an honesty
  boundary stating that rows exist only where `go test ./...` proves them.
- Architecture diagram (`docs/diagrams/production-systems.mmd`, rendered
  inline in `docs/architecture.md`): request pipeline with failure/deny
  branches, each stage mapped to the failure matrix and its tests. The wiring
  is labelled as the operator's responsibility; every stage is an independent,
  tested package.

## [v0.1.0] — 2026-08-14

Initial release.

### Added
- Reliability: `idempotency` (key store + gateway + conflict detection),
  `replay` (nonce + window), `circuitbreaker` (3-state), `retry`
  (backoff + jitter + selective retryability).
- Security: `auth` (HMAC keys + RBAC), `ratelimit` (token bucket), `validation`
  (Luhn, monetary amounts, rule combinators).
- Observability: `audit` (append-only hash-chained log with verification).
- Webhook HMAC-SHA256 signing and constant-time verification.
- Benchmarks for hot paths.
- CI (Go test with `-race`, `go vet`, `gofmt`).

### Design
- Zero external dependencies; storage behind interfaces (swap in Postgres/Redis).
- Each package defends exactly one failure mode, documented in README/methodology.
