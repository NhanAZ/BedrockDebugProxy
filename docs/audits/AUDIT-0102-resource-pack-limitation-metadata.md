# Audit 0102 - Correct resource-pack limitation metadata

## Intent

Correct the capture manifest limitation that still described upstream optional resource-pack policy as unsupported after the bridge began retaining `ResourcePacksInfo` metadata.

## Scope

- Updated `captureLimitations` in `cmd/bedrock-debug-proxy/main.go`.
- Added regression assertions in `cmd/bedrock-debug-proxy/main_test.go`.
- Updated `CHANGELOG.md` and `docs/fix-history.md`.

This change affects capture metadata only. It does not alter packet forwarding, resource-pack retrieval, ordering, or raw evidence.

## Evidence and provenance

The listener callback in `internal/proxy/proxy.go` supplies retained upstream `ResourcePacksInfo` metadata. `docs/resource-packs.md` records that metadata is preserved while URL entries are replaced with archive-backed entries that the proxy can serve.

## Automated validation

- `go test ./cmd/bedrock-debug-proxy`
- `git diff --check`
- Full `tools/quality.ps1 -LowMemory` passed.

## Live validation

Not applicable to wire behavior. A mixed URL and RakNet offer still requires owner-controlled real-client validation as documented elsewhere.

## Known limitations and follow-up

The corrected wording does not claim byte-identical `ResourcePacksInfo` offers. URL fields are intentionally removed from downstream copies because the proxy serves the reconstructed archive over its own connection. Failed retrieval remains part of the capture evidence.
