# Changelog

All notable changes to `production-systems-lab`.

## [Unreleased]

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
