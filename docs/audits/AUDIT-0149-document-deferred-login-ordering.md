# Audit 0149 - Document deferred-login raw and decoded ordering

## Intent

Clarify why a deferred login packet can have an early `packet.raw` event and a later `packet.decoded` event without indicating packet loss or wire reordering.

## Scope

- Updated `docs/research/capture-observability.md` with local capture evidence and the raw-hook, deferred-queue, and bridge-consumption boundaries.
- Added FIX-0115 to `docs/fix-history.md`.
- No runtime, capture schema, packet, or forwarding behavior changed.

## Evidence and provenance

- Owner-controlled capture [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive) at revision `d272805fd571525cca6f7473b166fb1129bafd48` records raw login packets before `session.spawned` and their decoded views after the bridge consumes the deferred queue.
- The in-tree gophertunnel `packet.go`, `conn.go`, and `raw_packet.go` source establishes that the packet hook runs before login decode and that deferred packets are returned later by `ReadPacketDataWithTime`.
- The recorder assigns one ordered submission sequence across concurrent observer and bridge callbacks, so adjacent sequence numbers are not a raw-to-decoded pairing contract.

## Automated validation

- Documentation consistency and `git diff --check` are required before commit.
- No runtime tests were changed because this is a research clarification only.

## Live validation status

No live or public-server validation was performed. The evidence is from an owner-controlled local capture and source inspection.

## Known limitations and follow-up

The current capture does not add an explicit raw-event foreign key to decoded events. Analysis must use packet identity, direction, connection, and raw blob evidence rather than adjacent sequence numbers. Future protocol conversion paths may have additional typed packets without a raw representation at the bridge boundary.
