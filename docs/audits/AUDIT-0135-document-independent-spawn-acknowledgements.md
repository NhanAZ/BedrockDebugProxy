# Audit AUDIT-0135 - Document independent spawn acknowledgements

- Commit subject: `Document independent spawn acknowledgements`
- Intent: explain the apparent startup ordering difference between the upstream and downstream streams without misclassifying consumed spawn acknowledgements as packet loss or duplication.
- Scope: update `docs/research/start-game-parity.md` and `docs/fix-history.md`. No runtime code, capture schema, packet bytes, authentication behavior, or forwarding policy changed.
- Evidence and provenance: owner-controlled capture [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive); packet ID 70 and spawn-status packet ID 2 have matching raw hashes across upstream and downstream. Reviewed `third_party/gophertunnel/minecraft/conn.go` handlers and `internal/proxy/proxy.go:startGame` at the current revision.
- Automated validation: `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit. Documentation-only change has no runtime test delta.
- Live validation: not performed. The proxy and Minecraft client were already stopped, and no public server or external account was contacted.
- Known limitations: the note does not make the two authenticated login chains identical and does not claim raw UDP or RakNet parity. It does not establish any server-policy or enforcement signal.
- Follow-up: if a future state-machine change substitutes consumed acknowledgements, add a deterministic fixture that verifies downstream spawn timing and prevents duplicate responses before changing runtime code.
