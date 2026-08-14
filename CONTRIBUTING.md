# Contributing

Thanks for your interest.

## Ground rules

- **Provider-agnostic core.** `core/`, `workflow/`, `knowledge/`, and
  `observability/` must not import any specific LLM SDK.
- **Offline tests.** New features ship with tests that run without network or
  API keys (use a fake provider).
- **Small PRs.** One logical change per pull request.

## Getting started

1. Fork and clone.
2. `pip install -e ".[test]"`.
3. `python -m pytest tests/`.

## Pull requests

- Add or update a test with every change.
- Keep `examples/` runnable with zero configuration.
- Update `CHANGELOG.md`.

## Code of conduct

Be respectful and constructive. See `CODE_OF_CONDUCT.md`.
