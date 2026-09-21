# Audit AUDIT-0143

## Intent

Preserve Transfer rewrite error evidence without dereferencing an absent raw packet read.

## Scope

- Updated `internal/proxy/proxy.go` so the Transfer rewrite failure path releases a retained raw packet conditionally and preserves the wrapped operation and capture errors.
- Added `TestFinishTransferRewriteErrorAtHopAllowsDecodedPacketWithoutRawRead` in `internal/proxy/transfer_test.go`.
- Updated `docs/research/capture-observability.md`, `docs/fix-history.md`, and the Unreleased changelog.

## Evidence and provenance

The forwarding loop sets `rawRead` to nil after a raw read already exposes a decoded packet and never sets it on typed forwarding. The previous Transfer rewrite error branch unconditionally called `rawRead.Release()`, so an invalid listener address could panic before the bridge error was captured. The focused helper keeps the existing error event and cleanup semantics while making the nil case explicit. No external source code was copied.

## Automated validation

- Focused proxy transfer tests, including the no-raw-read rewrite failure regression, are required before commit.
- Full `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit.

## Live validation

No live server, public service, or external account was contacted. Live Transfer rewrite failures remain pending.

## Known limitations and follow-up

This change only fixes error-path safety and provenance. It does not change Transfer routing, packet rewriting, or the opt-in follow-transfers policy.
