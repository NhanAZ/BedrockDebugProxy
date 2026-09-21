# Audit AUDIT-0113

## Intent

Record the confirmed difference between downstream and upstream
`ClientCacheStatus` values without changing packet forwarding or startup order.

## Scope

- Added `docs/research/client-cache-parity.md`.
- Added FIX-0081 to `docs/fix-history.md`.
- No runtime code, capture schema, dependency, or generated artifact changed.

## Evidence and provenance

- `captures/session-20260921T122228Z` contains downstream packet ID 129 with
  payload `01` and upstream packet ID 129 with payload `00`.
- The same closed capture records the two values in `session.spawned`.
- The source path was reviewed in the vendored gophertunnel copy at commit
  `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`.
- The note classifies confirmed, inferred, and unknown behavior separately and
  explains why a generic normalization is not currently safe.

## Validation

- Automated validation: documentation-only. `tools/quality.ps1 -LowMemory` and
  `git diff --check` are required before commit.
- Local validation: existing closed-capture evidence was inspected directly.
- Live validation: not performed. No server or account was contacted.

## Limitations and follow-up

The note does not claim that the cache mismatch causes enforcement or that a
future redesign should force the value true. A controlled local cache-enabled
server test is required before changing the startup architecture.
