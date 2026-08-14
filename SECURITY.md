# Security Policy

## Supported Versions

| Version | Supported |
| --- | --- |
| `main` | ✅ Supported |
| `v0.1.0` | ✅ Supported |

## Reporting a Vulnerability

Do **not** disclose security vulnerabilities publicly. Report privately through
a **GitHub Security Advisory** on this repository.

Include the affected file, a description, reproduction steps, and a suggested
fix if possible.

## Design notes

- The approval gate is the security boundary for automation: irreversible
  actions require explicit human approval.
- Provider credentials are never committed; they are injected at runtime.
- Audit log is append-only by design so post-hoc tampering is detectable.
