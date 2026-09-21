# Audit AUDIT-0134 - Document typed additional-packet boundary

- Commit subject: `Document typed additional-packet boundary`
- Intent: record the raw-evidence boundary when gophertunnel returns a typed packet without raw bytes after login-state consumption or protocol conversion.
- Scope: update `docs/research/capture-observability.md` and `docs/fix-history.md`. No runtime code, capture schema, packet bytes, authentication behavior, or forwarding policy changed.
- Evidence and provenance: reviewed the in-tree gophertunnel `minecraft/raw_packet.go`, `minecraft/packet.go`, `minecraft/conn.go`, and `minecraft/protocol.go`. Protocol 2193 currently uses identity conversion. The packet hook remains the earlier raw observation boundary.
- Automated validation: `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit. Documentation-only change has no runtime test delta.
- Live validation: not performed. No public server or external account was contacted.
- Known limitations: a future protocol adapter may produce multiple typed packets from one input. The project does not fabricate raw bytes for those values and still needs a fixture before adding derived annotations or forwarding changes.
- Follow-up: if a supported protocol implementation begins returning additional packets, add a deterministic capture fixture and explicit provenance fields before changing the bridge.
