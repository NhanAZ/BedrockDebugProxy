# Audit AUDIT-0080 - Record authenticated upstream boundary

## Intent

Document why the authenticated upstream login retains gophertunnel's Android metadata normalization instead of copying downstream device fields into a separately signed chain.

## Scope

- Updated `docs/research/local-identity-parity.md` with the immutable upstream source revision and its authenticated-login behavior.
- No runtime code, packet mutation, or authentication behavior changed.

## Evidence and provenance

- Sandertv/gophertunnel commit `7a556a07335b663744b50d38062636ad8283f314` calls `setAndroidData` before `login.Encode` in the authenticated path and documents that the chain title ID cannot be edited.
- The same source creates a new upstream key and signed login request, so downstream and upstream authenticated sessions cannot share one official client chain through this relay.

## Automated validation

- Documentation-only change. `git diff --check` and the project quality gate are required before commit.

## Live validation

No public server was contacted. No anti-cheat, authentication, or policy control was bypassed.

## Known limitations and follow-up

Authenticated two-session parity remains an architectural limitation. Any further change would require an authorized server-side test and a protocol-supported identity design, not copying unsigned metadata fields.
