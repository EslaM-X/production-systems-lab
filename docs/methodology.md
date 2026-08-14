# Methodology

## Why these packages exist

Distributed and financial systems fail in *predictable* ways. The five
failures this lab defends against are the ones I have hit directly:

1. **Double settlement / double write.** A client times out, retries, and the
   server applies the side effect twice. → Idempotency keys.
2. **Replay of captured requests.** An attacker records a valid request and
   re-sends it later to trigger a second settlement. → Nonce + time window.
3. **Cascading failure.** One slow dependency backs up callers until the whole
   service collapses. → Circuit breakers.
4. **Transient errors treated as fatal.** A packet is dropped and a whole job
   dies instead of retrying. → Backoff + jitter.
5. **Tampered records.** A trail that can be edited is no trail at all. →
   Append-only hash-chained audit.

## Trade-offs made explicit

- **Idempotency keyed by payload hash:** strong protection, but requires the
  client to include a payload; callers without keys are not protected.
- **In-memory stores:** fast, portable, but single-process. Swap for
  Postgres/Redis behind the same interface for distributed deployments.
- **HMAC over JWT for API keys:** no signature-verification dependency chain,
  easy rotation by key id; but no built-in expiry — the application must add it.

## Testing philosophy

- Every defence is tested for *its* failure mode: conflict, replay, open
  circuit, exhaustion, tamper — plus the happy path.
- Benchmarks are provided (`go test -bench`) so performance changes are
  measured, not guessed.
