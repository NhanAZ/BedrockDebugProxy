# ADR 0011 - Forward the observed client cache capability

## Status

Accepted on 2026-09-22.

## Context

The terminating bridge has two Bedrock login sessions. The downstream client
sends `ClientCacheStatus` after `PlayStatus(LoginSuccess)`. The upstream Dialer
must send its own status before the upstream server sends `ResourcePacksInfo`.
The previous listener callback started the upstream login before it could read
the downstream status, so the Dialer used its false zero value even when a
vanilla Windows client had advertised `true`.

This was a confirmed local wire difference in [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive).
The value is a client capability, not a platform default, so unconditionally
enabling it would be incorrect for a client that reports `false`.

## Decision

Add an opt-in listener mode that sends `PlayStatus(LoginSuccess)` first,
waits for `ClientCacheStatus`, and then runs the existing dynamic
resource-pack callbacks. BedrockDebugProxy enables this mode and passes the
observed boolean to the upstream Dialer's `EnableClientCache` field. The
resource-pack offer, archive delivery, stack retention, and bridge forwarding
remain unchanged after the callback completes. Listeners that do not enable the
mode keep their previous ordering.

## Consequences

- The upstream connection advertises the same cache capability that the
  downstream client reported, without guessing from device metadata.
- The downstream client waits for the upstream resource-pack exchange after it
  sends its cache status. This is the existing upstream-first barrier with one
  earlier capability packet, not a second connection or a packet mutation.
- A client that closes before sending `ClientCacheStatus` does not complete the
  deferred offer, which is consistent with the login sequence requiring that
  status before the resource-pack response.
- The capture still retains both raw status packets and the per-connection
  `session.spawned` values.

## Evidence and provenance

- [`docs/research/client-cache-parity.md`](../research/client-cache-parity.md)
  records the packet evidence, protocol reference, and local PocketMine
  ordering.
- The root loopback integration test enables the cache on its vanilla-shaped
  client and verifies that the upstream connection and closed capture both
  report `true`.
- The change uses gophertunnel's existing `Dialer.EnableClientCache` and does
  not alter packet contents or gameplay forwarding.

## Limitations

The local integration fixture exercises the cache capability, login ordering,
`LevelChunk` cache metadata, and the miss-response exchange with raw payload
preservation. It does not prove that a real vanilla client stores or reuses
those blobs, nor how any public server uses the capability. A real-client
local capture remains useful for that behavior.
