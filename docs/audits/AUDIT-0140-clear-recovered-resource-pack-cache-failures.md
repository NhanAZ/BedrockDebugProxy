# Audit AUDIT-0140

## Intent

Make a successful RakNet resource-pack fallback replace a stale non-fatal URL-prefetch error for the same cache key.

## Scope

- Updated `internal/proxy/resource_pack_url.go` to remove the error entry when `Store` accepts a matching pack.
- Added `TestURLResourcePackCacheStoreClearsPreviousPrefetchError` in `internal/proxy/resource_pack_url_test.go`.
- Updated `docs/resource-packs.md`, `docs/fix-history.md`, and the Unreleased changelog.

## Evidence and provenance

The vendored `ResourcePackCache` contract states that `Load` errors are non-fatal misses and that `Store` makes a pack available for a later load. The URL adapter recorded failed prefetches in a separate error map, but `Store` left that map entry in place. The focused change restores the cache contract without altering packet bytes, offer order, or the normal fallback path. No external source code was copied.

## Automated validation

- Focused proxy resource-pack tests, including the stale-error regression, are required before commit.
- Full `tools/quality.ps1 -LowMemory` and `git diff --check` are required before commit.

## Live validation

No live server, public service, or external account was contacted. Real-client URL-failure and RakNet-fallback validation remains pending.

## Known limitations and follow-up

The change applies to the per-connection URL cache and does not add offline pack retrieval or change resource-pack negotiation. A controlled server that first exposes an unreachable URL and then supplies the same pack over RakNet is still needed for end-to-end validation.
