# ADR 0006 - Keep login-order tolerance bounded by expected packet state

## Status

Accepted on 2026-09-14.

## Context

Featured Experiences do not all use one observable login sequence. The project has observed an early
`ResourcePacksInfo` on Enchanted and has documented historical Hive and CubeCraft differences in
[`docs/research/cubecraft-gophertunnel-discord-audit.md`](../research/cubecraft-gophertunnel-discord-audit.md). A fix for one sequence must not become a generic packet reorderer.

The in-tree gophertunnel connection defers packets whose IDs are not valid for the current login state. The
state transition must revisit already-received packets or a valid early packet can leave the dialer waiting
forever. At the same time, processing every deferred packet as soon as any state changes could consume a
later-phase packet too early and hide an ordering regression.

## Decision

Keep `expectedIDs` as the sole acceptance boundary during login. When a state transition changes that set,
scan the deferred queue for the earliest packet that is currently expected. Remove and handle only that packet,
then re-evaluate the queue after the handler advances the state again. Packets that are not currently expected
remain in arrival order in the queue. A decode or handler error closes the connection instead of leaving the
upstream dial blocked.

This policy tolerates observed server-specific order differences without changing packet bytes, inventing a
packet, dropping an unknown packet, or allowing a packet from an unverified later phase to skip the state
machine. Same-protocol forwarding remains raw, so the compatibility patch is limited to the terminating login
boundary where gophertunnel must consume packets to establish the connection.

## Consequences

- Enchanted's early `ResourcePacksInfo` can advance after `PlayStatus` without a new network read.
- A queued `ResourcePackStack` or `ItemRegistry` cannot block the packet that advances the current state, but it is not handled until its own expected phase.
- The queue algorithm is deterministic and retains raw evidence. It does not claim that all servers use one order.
- New ordering behavior requires a focused packet-order test and live validation on The Hive plus the affected server.

## Evidence and validation

- Failed Enchanted capture `captures/session-20260914T105050Z` at revision `96d8a01201d48b4a069c0e649374452a12b9f207`.
- Historical login-order evidence in [`docs/research/cubecraft-gophertunnel-discord-audit.md`](../research/cubecraft-gophertunnel-discord-audit.md).
- Regression tests in `third_party/gophertunnel/minecraft/conn_featured_experience_test.go` cover all resource-pack login permutations and preserve later-phase packets until their state.
- Exact-revision live validation for the latest queue implementation remains pending. The required next baseline is The Hive, followed by the affected Enchanted flow and the release matrix.

## Limitations

This boundary does not solve authenticated proxy identity differences, anti-cheat policy, invalid packet order,
resource-pack semantics that differ between hops, or server-specific transport failures. Such findings require a
separate capture-backed investigation rather than a wider expected-ID set.
