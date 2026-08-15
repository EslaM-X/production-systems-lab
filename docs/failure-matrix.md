# Failure Matrix

Every row below maps a real-world failure mode to the control that this
repository actually implements, and to the unit test that proves it. A row
exists here **only because** the code and the test exist — no speculative
defences, no claimed mechanisms without evidence.

The chain for every row is:

```
Failure → Control → Evidence
            ↑          ↑
      implementation   test (file:test)
```

## Matrix

| Failure Mode | Detection | Default Response | Recovery | Data Safety | Evidence |
| --- | --- | --- | --- | --- | --- |
| Duplicate request | Idempotency key + payload hash | Reject on conflict, otherwise safe replay of stored result | Replay from store | Protected | `TestExecuteRetryReturnsStoredResult`, `TestExecuteSameKeyDifferentPayloadConflicts` |
| Transient dependency failure | Error classification (`Retryable`) | Retry with exponential backoff + jitter | Backoff then succeed | Preserved | `TestRetryRetriesUntilSuccess`, `TestRetryExhausts` |
| Persistent dependency failure | Failure threshold counter | Fail fast (`ErrOpen`) | Half-open probe recovery | Preserved | `TestOpensAfterThreshold`, `TestHalfOpenProbeSucceedsCloses`, `TestHalfOpenProbeFailsReopens` |
| Invalid signature | HMAC-SHA256 verification | Reject (`ErrInvalidKey` / `ErrInvalidSignature`) | None (re-issue key / re-sign) | Protected | `TestVerifyRejectsTamperedKey`, `TestVerifyRejectsWrongSecret`, `TestVerifyRejectsTamperedPayload`, `TestVerifyRejectsGarbageSignature` |
| Rate-limit exhaustion | Token-bucket state | Reject (`Allow` returns false) | Wait for refill | Preserved | `TestAllowConsumesTokens`, `TestRefill` |
| Audit-chain mismatch | Hash-chain verification | Fail closed (`Verify` returns false) | Investigate the trail | Protected | `TestTamperDetected`, `TestChainLinks`, `TestAppendAndVerify` |
| Webhook duplication | Replay nonce + signature | Deduplicate (reject seen nonce) / reject forged signature | Replay-safe delivery | Protected | `TestValidateRejectsDuplicateNonce`, `TestSignVerifyRoundTrip` |

## Evidence by row

### 1. Duplicate request → idempotency layer → replay stored result

- **Implementation**: `reliability/idempotency/gateway.go:34` — `Gateway.Execute`
  claims the key via `Store.TryClaim` (`reliability/idempotency/idempotency.go:60`),
  which binds the key to a SHA-256 payload hash. A completed key returns the
  stored response without re-running the operation; the same key with a
  different payload returns `ErrConflict` (`reliability/idempotency/idempotency.go:18`).
- **Tests**: `TestExecuteFirstCallRunsOp`, `TestExecuteRetryReturnsStoredResult`
  (op runs exactly once), `TestExecuteSameKeyDifferentPayloadConflicts`,
  `TestHashDeterministic`.
- **Data safety**: Protected — payload identity is bound to the key, so a key
  cannot silently absorb a different request.

### 2. Transient dependency failure → error classification → retry with backoff

- **Implementation**: `reliability/retry/retry.go:45` — `Retry` honours
  `Config.Retryable` to classify failures; transient errors sleep exponential
  backoff (`retry.go:69`) with jitter and a capped `MaxDelay`; a non-retryable
  error returns immediately; exhausted attempts join `ErrExhausted`
  (`reliability/retry/retry.go:38`).
- **Tests**: `TestRetryRetriesUntilSuccess` (3 attempts, succeeds),
  `TestRetryStopsOnNonRetryable` (exactly 1 call), `TestRetryExhausts`
  (exactly `MaxAttempts` calls).
- **Data safety**: Preserved — nothing is committed while an attempt is
  failing.

### 3. Persistent dependency failure → circuit breaker → fail fast, half-open recovery

- **Implementation**: `reliability/circuitbreaker/breaker.go:47` — `Breaker`
  opens after `FailureThreshold` consecutive failures, `Allow` fails fast with
  `ErrOpen` (`breaker.go:16`), a cooldown moves to `HALF-OPEN` where a probe
  decides (`Succeed` closes after `SuccessThreshold`; a probe failure reopens).
  `Execute` (`breaker.go:108`) wraps any downstream call.
- **Tests**: `TestOpensAfterThreshold`, `TestHalfOpenProbeSucceedsCloses`,
  `TestHalfOpenProbeFailsReopens`.
- **Data safety**: Preserved — load is shed before downstream mutations can
  cascade.

### 4. Invalid signature → HMAC verification → reject

- **Implementation**: `security/auth/auth.go:56` — `Issuer.Verify` recomputes
  the HMAC-SHA256 signature with constant-time `hmac.Equal`; tampered or
  foreign-signed keys yield `ErrInvalidKey`. For webhook payloads,
  `webhook/webhook.go:17` — `Verify` decodes the hex signature, recomputes the
  HMAC, and rejects with `ErrInvalidSignature`.
- **Tests**: `TestVerifyRejectsTamperedKey`, `TestVerifyRejectsWrongSecret`
  (auth); `TestVerifyRejectsWrongSecret`, `TestVerifyRejectsTamperedPayload`,
  `TestVerifyRejectsGarbageSignature` (webhook); positive controls
  `TestMintVerifyRoundTrip`, `TestSignVerifyRoundTrip`.
- **Data safety**: Protected — rejection happens before any state change.

### 5. Rate-limit exhaustion → token-bucket state → reject, refill

- **Implementation**: `security/ratelimit/ratelimit.go:39` — `AllowN` refills
  the bucket from elapsed time at `rate` tokens/sec, caps at `capacity`, and
  denies when insufficient tokens remain. `Allow` consumes one token.
- **Tests**: `TestAllowConsumesTokens` (burst of 2, third denied),
  `TestAllowN` (bulk tokens), `TestRefill`.
- **Data safety**: Preserved — denied requests never reach the operation.

### 6. Audit-chain mismatch → hash verification → fail closed

- **Implementation**: `observability/audit/audit.go:47` — `Append` links each
  entry to the previous entry's hash; `Verify` (`audit.go:77`) recomputes every
  hash and re-links `PrevHash`, returning `false` on any mutation.
  `EnforceAppendOnly` (`audit.go:99`) additionally rejects non-monotonic
  sequences.
- **Tests**: `TestAppendAndVerify`, `TestTamperDetected`,
  `TestChainLinks`, `TestEnforceAppendOnly`.
- **Data safety**: Protected — a tampered trail is detectable, not silently
  trusted.

### 7. Webhook duplication → replay guard + signature → deduplicate, replay-safe

- **Implementation**: `webhook/webhook.go:17` authenticates the delivery
  (HMAC-SHA256); `reliability/replay/replay.go:53` — `Guard.Validate` rejects a
  replayed nonce (dedup by event identity) and out-of-window timestamps with
  `ErrReplay`. Together: a replayed delivery cannot carry a forged signature,
  and a captured delivery cannot be re-sent within the replay window.
- **Tests**: `TestSignVerifyRoundTrip`, `TestVerifyRejectsTamperedPayload`
  (webhook); `TestValidateRejectsDuplicateNonce`,
  `TestValidateRejectsExpired`, `TestValidateRejectsFutureSkew`,
  `TestValidateAllowsDifferentNoncesSameTime` (replay).
- **Data safety**: Protected — deduplicated and authenticated before any
  processing.

## Honesty boundary

- A row is only listed if `go test ./...` exercises its control (all evidence
  columns are real test names from this repository).
- The matrix asserts **what the mechanism does**, not that every deployment
  configures it. Wiring (`reliability/replay`) and key distribution are the
  operator's responsibility.
- Recovery is stated as implemented: `replay` rejects and requires a fresh
  nonce; it does not retry on its own.

## Maintaining this matrix

- To add a failure mode: implement the control, add the test, then add the row
  with the exact test name and `package/file.go:line` of the implementation.
- To change behavior: update the test first, then the row — the matrix must
  never describe behavior the tests do not prove.

## See also

- `docs/architecture.md` — flow and state-machine diagrams for the core controls.
- `README.md` — package-by-package failure-mode/defence index.
