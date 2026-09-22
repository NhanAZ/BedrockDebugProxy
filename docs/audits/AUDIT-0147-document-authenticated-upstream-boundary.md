# Audit 0147 - Document authenticated upstream identity boundary

## Intent

Record a confirmed local XUID rejection when an authenticated downstream client is paired with an unauthenticated upstream dial, so future parity work does not treat the identity-boundary failure as permission to forge or bypass authentication.

## Scope

- Updated `docs/research/local-identity-parity.md` with the sanitized evidence and confirmed, inferred, and unknown classifications.
- Added FIX-0113 to `docs/fix-history.md`.
- No source, capture schema, protocol behavior, or runtime code changed.

## Evidence and provenance

- Capture: `captures/session-20260921T121717Z`.
- Capture binary revision: `d272805fd571525cca6f7473b166fb1129bafd48`.
- The capture records `auth_mode=none`, an upstream `XUID does not match (possible impersonation attempt)` error, and no completed `session.negotiated` event.
- The source-backed authenticated boundary and offline identity tests remain documented in `docs/research/local-identity-parity.md`.

## Validation

- Documentation-only change. No runtime validation was required.
- `git diff --check` passed before commit.
- The existing quality gate and local tests remain the authoritative runtime validation for the unchanged source tree.

## Live validation status

No public-server or external-account validation was performed.

## Known limitations and follow-up

An authenticated upstream requires an explicitly authorized authenticated test. This clarification does not establish any public-server enforcement cause and does not change the proxy's authentication behavior.
