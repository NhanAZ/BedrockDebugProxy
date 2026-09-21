# Audit AUDIT-0137 - Mark chunk-radius mismatch evidence as historical

- Commit subject: `Mark chunk-radius mismatch evidence as historical`
- Intent: distinguish the pre-FIX-0083 `RequestChunkRadius` capture from the current typed-forwarding implementation so future analysis does not mistake an old initial-request mismatch for a current forwarding failure.
- Scope: update `docs/research/chunk-radius-parity.md` and `docs/fix-history.md`. No runtime code, capture schema, packet bytes, authentication behavior, or forwarding policy changed.
- Evidence and provenance: reviewed FIX-0083 through FIX-0086, the current downstream packet callback in `internal/proxy/proxy.go`, and the deterministic 16/16-then-24/28 assertions in `internal/proxy/proxy_test.go`.
- Automated validation: `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit. Documentation-only change has no runtime test delta.
- Live validation: not performed. The user ended the local proxy session. No public server or external account was contacted.
- Known limitations: a later server response or approved radius is not inferred from the typed forward. Public-server behavior remains unknown.
- Follow-up: if a fresh capture reports only the initial 16/16 request, compare its binary revision with FIX-0083 before diagnosing a regression.
