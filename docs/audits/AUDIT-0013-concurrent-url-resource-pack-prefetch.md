# AUDIT-0013 - Reduce upstream URL resource-pack blocking

- Commit subject: `Prefetch URL resource packs concurrently`
- Change type: `runtime`
- Runtime impact: `yes`

## Intent

Reduce the time a featured-experience backend is held in the existing upstream-first resource-pack barrier. A live
Enchanted Factions capture reached hop 2 spawn and then received an upstream `Timed out` kick after seven URL packs were
prefetched serially.

## Scope

The URL resource-pack adapter now uses a bounded worker pool with three workers and publishes successful packs in the
order advertised by `ResourcePacksInfo`. The affected code is `internal/proxy/resource_pack_url.go` and its regression
test. `ResourcePackCache` validation, duplicate suppression, downstream URL removal, sequential listener delivery,
transfer routing, packet bytes, and capture schema are intentionally unchanged. Documentation and fix-history entries
describe the observed timing and the compatibility boundary.

## Evidence and provenance

Capture `captures/session-20260916T053620Z` recorded seven hop-2 URL pack archives for
`factions-spawn.factions.connect.enchanted.gg:19132`, then `session.spawned`, followed immediately by the upstream
message `You were kicked: Timed out`. The manifest has zero dropped, truncated, decode, or capture write errors. The
serial prefetch timing is the root-cause hypothesis; the capture does not include RakNet acknowledgements or prove the
backend's internal timeout rule. No external source code was copied.

## Automated validation

- `go test ./internal/proxy -run 'TestURLResourcePackCache' -count=1`
- `go test -race ./internal/proxy`
- The new barrier-backed HTTP fixture verifies concurrent starts and original pack order.
- Full `tools/quality.ps1`, formatting, build, and `git diff --check` remain required before commit.

## Live validation status

Pending. Build a stamped binary from this change and run the Enchanted transfer flow plus The Hive minimum baseline.
Retain the complete captures and classify any upstream timeout separately from capture-integrity failures.

## Limitations and follow-up

The worker pool reduces, but does not remove, the upstream-first barrier. `resource.ReadURL` still loads and parses each
archive in memory, and the capture boundary still omits raw RakNet acknowledgement, retransmission, and fragmentation
details. If a stamped live run still times out after the shorter prefetch window, compare direct and proxy sessions and
inspect the upstream packet sequence before changing transport or authentication behavior.
