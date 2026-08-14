# Architecture

Each package owns one failure mode. There are no cross-package imports between
concerns, so each can be reasoned about and tested in isolation.

## Idempotency flow

```
request (key, payload)
  → hash payload (sha256)
  → TryClaim(key, hash)
      ├─ already claimed, different payload → 409 conflict
      ├─ already claimed, completed        → return stored response
      └─ new key                            → run op, MarkComplete, return
```

Key semantics: one logical operation = one key; the *payload hash* prevents a
key being silently reused for a different request.

## Circuit breaker state machine

```
CLOSED --failure>=threshold--> OPEN
OPEN   --cooldown elapsed-->   HALF-OPEN
HALF-OPEN --probe success>=n--> CLOSED
HALF-OPEN --probe failure-->    OPEN
```

## Audit chain

```
entry N = {seq, timestamp, actor, action, resource, prev_hash=hash(N-1), hash=hash(N)}
Verify() recomputes every hash and re-links prev_hash; any mutation breaks the chain.
```

## See also
- `docs/methodology.md`
