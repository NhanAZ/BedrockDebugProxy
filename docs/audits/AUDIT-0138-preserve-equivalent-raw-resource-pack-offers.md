# Audit AUDIT-0138 - Preserve equivalent raw resource-pack offers

- Commit subject: `Preserve equivalent raw resource-pack offers`
- Intent: close the confirmed same-protocol wire-preservation gap for `ResourcePacksInfo` offers that do not require URL or list transformation.
- Scope: retain raw offer bytes and source protocol in the vendored connection state, expose a narrow listener callback, select the raw path only after payload equivalence, and add nested and root loopback coverage. No authentication, transport, capture schema, or server-policy behavior changed.
- Evidence and provenance: reviewed `docs/research/resource-pack-stack-parity.md`, ADR 0004, the gophertunnel `ResourcePacksInfo` login and listener paths, and `internal/proxy/resource_pack_url.go` URL replacement behavior.
- Automated validation: focused nested gophertunnel tests and the root proxy loopback integration pass. `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit.
- Live validation: not performed. The user ended the local proxy session. No public server or external account was contacted.
- Known limitations: URL entries, changed offers, protocol mismatches, and write failures use typed serialization. Mixed URL, encrypted, and real-client offer validation remains pending.
- Follow-up: compare a fresh local capture with and without URL packs to confirm raw reuse and typed fallback in the same stamped binary.
