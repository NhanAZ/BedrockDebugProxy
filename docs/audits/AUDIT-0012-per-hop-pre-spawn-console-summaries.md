# AUDIT-0012 - Restore pre-spawn summaries for transfer hops

## Intent

Show login and resource-pack packet summaries for every transfer hop while preserving the existing low-volume console cadence and capture-first behavior.

## Scope

The change is limited to live reporter hop state in `internal/proxy/live_reporter.go`, hop synchronization in `internal/proxy/proxy.go`, the live reporter regression test, and the operator documentation and fix-history entry.

## Affected interfaces

- `liveReporter` tracks which hop has spawned instead of using one process-wide flag.
- `Runner.Run` marks the active hop before accepting its physical client connection.
- `SetSpawned` continues to suppress raw gameplay summaries after spawn, but only for the current hop.
- `SetHop` flushes the previous bucket before a new transfer hop begins.

## Evidence and provenance

Enchanted capture `captures/session-20260916T044421Z` recorded hop 2 resource-pack packets and `session.spawned`, while the corresponding terminal run did not print pre-spawn packet summaries. The capture observer already retained these raw events, so this is an operator-observability fix and does not change packet bytes, ordering, resource-pack delivery, or transfer routing. No external source code was copied.

## Automated validation

- The live reporter regression test covers pre-spawn summaries on two hops and confirms that post-spawn raw packet summaries remain suppressed within each hop.
- `tools/format.ps1`, `tools/quality.ps1`, and `git diff --check` are required before commit.

## Live validation status

Pending. Build a stamped binary from this commit, run The Hive as the minimum baseline, and repeat the Enchanted transfer flow. The terminal should show hop 2 login and resource-pack summaries before its spawn, while the capture remains complete and packet forwarding is unchanged.

## Known limitations

This change does not prevent an upstream server timeout after spawn and does not add per-packet logging outside the existing three-second buckets. The canonical capture remains the authoritative source for complete packet evidence.

## Follow-up work

If a later hop still lacks summaries, inspect the hop transition state and capture observer callback before changing forwarding or protocol handling.
