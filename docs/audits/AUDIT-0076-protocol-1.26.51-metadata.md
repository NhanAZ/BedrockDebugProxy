# Audit AUDIT-0076 - Align Bedrock 1.26.51 metadata

## Intent

Align authenticated upstream `ClientData.GameVersion` with a 1.26.51 downstream client when both connections use protocol 2193, without changing packet definitions or forwarded bytes.

## Scope

- Updated `third_party/gophertunnel/minecraft/protocol/info.go` from game version 1.26.50 to 1.26.51 while retaining protocol 2193.
- Updated focused protocol and proxy integration expectations and added a direct regression test for the authenticated Android metadata helper.
- Added the 1.26.51 research note and updated the protocol baseline, provenance, incident research, fix history, changelog, and third-party patch description.

## Evidence and provenance

- Mojang [`bedrock-protocol-docs` release `v1.26.51`](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.51), commit `8667debe0adad27de6c692c88942c489a1193206`, reports network protocol 2193.
- Mojang's [protocol changelog](https://mojang.github.io/bedrock-protocol-docs/changelog/) reports zero packet changes for stable 1.26.51.
- The closed local Hive capture `captures/session-20260920T135412Z` records downstream ClientData 1.26.51 and authenticated upstream ClientData 1.26.50. This is evidence of a local metadata mismatch, not evidence of the server's punishment rule.
- The in-tree gophertunnel authenticated Android path assigns `protocol.CurrentVersion` to upstream ClientData. The change keeps the existing authentication and identity flow intact.

## Automated validation

- `gofmt` passed for changed Go files.
- Focused protocol and proxy tests passed.
- `tools/quality.ps1` passed, including build, static analysis, vulnerability scan, viewer asset and HTTP tests, JavaScript syntax checks, tracked-artifact checks, and whitespace validation.
- `git diff --check` passed.

## Live validation

No public server or authenticated vanilla client was contacted. The local PocketMine `ProxyFingerprint` direct-versus-proxy evidence remains valid for the unauthenticated loopback path. Authenticated parity still requires a local vanilla-client or otherwise authorized test-server run.

## Known limitations and follow-up

This metadata correction does not identify the cause of the Hive `Unfair Advantage` response, make proxy use compliant with Hive policy, or bypass anti-cheat enforcement. A future authorized validation must use a stamped binary and inspect downstream and upstream ClientData versions in the closed capture.
