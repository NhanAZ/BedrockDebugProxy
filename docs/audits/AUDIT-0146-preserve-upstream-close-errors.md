# Audit AUDIT-0146

## Intent

Preserve upstream connection close errors when setup or capture preparation fails after dialing.

## Scope

- Updated `internal/proxy/proxy.go` so resource-pack recording, downstream pack preparation, and `upstream.connected` cleanup paths join operation and close errors.
- Added `TestCloseAfterUpstreamErrorPreservesCloseError` in `internal/proxy/proxy_test.go`.
- Updated `docs/research/capture-observability.md`, `docs/fix-history.md`, and the Unreleased changelog.

## Evidence and provenance

`connectUpstream` owns the opened connection before returning it and previously discarded `Close()` results in three setup-failure branches. The helper keeps cleanup order and returns both causes without changing packet or transport setup behavior. No external source code was copied.

## Automated validation

- Focused upstream cleanup, route-probe, and loopback forwarding tests are required before commit.
- Full `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit.

## Live validation

No live server, public service, or external account was contacted. Live upstream cleanup failures remain pending.

## Known limitations and follow-up

The change covers the three post-dial setup branches. It does not change resource-pack delivery, routing, or shutdown policy.
