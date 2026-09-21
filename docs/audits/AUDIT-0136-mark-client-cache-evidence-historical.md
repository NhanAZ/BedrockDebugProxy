# Audit AUDIT-0136 - Mark client-cache mismatch evidence as historical

- Commit subject: `Mark client-cache mismatch evidence as historical`
- Intent: distinguish the pre-FIX-0091 client-cache mismatch capture from the current deferred-offer implementation so future analysis does not treat resolved evidence as a live regression.
- Scope: update `docs/research/client-cache-parity.md` and `docs/fix-history.md`. No runtime code, capture schema, packet bytes, authentication behavior, or forwarding policy changed.
- Evidence and provenance: reviewed FIX-0091 and FIX-0092, the client-cache research note, the current `internal/proxy/proxy.go` capability handoff, and the deterministic loopback assertions in `internal/proxy/proxy_test.go`.
- Automated validation: `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit. Documentation-only change has no runtime test delta.
- Live validation: not performed. The user has ended the local proxy session. No public server or external account was contacted.
- Known limitations: the historical capture still documents the original wire difference. Real vanilla-client cache storage and reuse remain unverified.
- Follow-up: if a fresh capture reports mismatched cache metadata, compare its binary revision with FIX-0091 before investigating a new runtime defect.
