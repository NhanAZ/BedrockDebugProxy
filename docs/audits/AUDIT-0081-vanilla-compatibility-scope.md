# Audit AUDIT-0081 - Define vanilla compatibility scope

- Commit subject: `Define vanilla compatibility scope`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Record that vanilla compatibility means protocol, transport, metadata, and capture parity on an authorized test server. It does not mean implementing or evading anti-cheat behavior.

## Scope

- Added `docs/decisions/0010-vanilla-compatibility-scope.md`.
- No runtime code, packet handling, authentication behavior, capture format, or CLI behavior changed.

## Evidence and provenance

- Existing capture-first and raw-forwarding decisions.
- Local identity parity evidence and its documented validation limits.
- The unresolved Hive incident research note, which separates observable capture facts from public-server policy hypotheses.

## Validation

- Automated checks: `git diff --check` and the project quality gate.
- Live validation: `not_observed`.
- Capture or report: `none`.

## Limitations and follow-up

This is a documentation-only scope clarification. No runtime impact. Authenticated two-session parity still needs an authorized owner-controlled test if it is to be investigated further.
