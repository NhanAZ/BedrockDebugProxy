# ADR 0003 - Opt-in transfer following

## Status

Accepted on 2026-09-13.

## Context

Featured Experiences and network hubs may send a Bedrock `Transfer` packet after login or during play. The client normally leaves the current connection and connects directly to the target. A terminating debug proxy that only forwards the current connection therefore stops observing at the first hop.

## Decision

`--follow-transfers` enables an explicit compatibility mode. When an upstream `Transfer` is observed, the proxy preserves the original packet in the capture, records a structured rewrite event, and sends a copy whose address and port point to the local listener. The proxy keeps the listener open, waits for the client to reconnect, and establishes the original target as the next upstream hop. Each hop receives its own connection IDs and hop number while remaining in one logical capture session.

The default remains non-mutating. Without the flag, the original `Transfer` is forwarded unchanged, recorded, and the current process ends when that connection closes. Transfer following is not a cheat or bypass feature and must not be used to evade a server's rules or anti-cheat controls.

## Constraints

- The local listener address sent to the client must be a concrete, reachable host and port. Wildcard addresses such as `0.0.0.0:19132` are rejected for the rewritten packet.
- The old client connection is retained only long enough for the reliable transfer packet and reconnect handshake. A bounded grace period prevents a stalled client from holding the single-player listener forever.
- A session follows at most 64 hops. The limit prevents an accidental transfer loop from creating unbounded connections and capture output.
- The original upstream packet and payload remain authoritative evidence. The rewritten downstream packet is separately observable through the downstream packet hook.
- The feature does not follow arbitrary redirects outside the Bedrock `Transfer` packet or change any other gameplay packet.
