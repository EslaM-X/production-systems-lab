# Roadmap

## North star

> A stranger clones this repo, runs one example against a real service, and
> sees the correctness loops (idempotency, retries, circuit breaking)
> working — within minutes.

Signal that matters, in order:

1. A newcomer runs the payment-gateway demo and understands the flow.
2. External users run the examples against their own services.
3. External contributors open PRs with new examples or patterns.

## Sequenced next

1. **More real demos** — add a second scenario (e.g. order processing with
   outbox pattern) that mirrors the depth of `payment-gateway`.
2. **Verification harness** — a script that runs every example and asserts
   its expected output, so README demos never drift from reality.
3. **Pattern library** — extract the reusable correctness patterns
   (idempotency, retry, circuit breaker) into documented reference sections.
4. **Adoption funnel** — CONTRIBUTING guide, good-first-issue labels, and a
   "run in minutes" quickstart at the top of the README.

## Explicitly out of scope (for now)

- New languages or frameworks without a demonstrated need.
- Cosmetic screenshots that don't show real behavior.
- Framework-ification before real usage exists.
