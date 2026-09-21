# Chunk-radius parity

## Scope

This note records the initial `RequestChunkRadius` difference observed in an owner-controlled local capture and the narrow forwarding change that now preserves the downstream request for the upstream connection. It does not claim that the proxy can determine a client setting before the client sends the packet, and it does not change gameplay packets or choose a radius from the device platform.

## Confirmed capture evidence

The closed capture `captures/session-20260921T122228Z` contains packet ID 69 (`RequestChunkRadius`) on both bridge channels.

| Channel | Payload | Decoded request |
| --- | --- | --- |
| downstream | `301c` | `ChunkRadius=24`, `MaxChunkRadius=28` |
| upstream | `2010` | `ChunkRadius=16`, `MaxChunkRadius=16` |

The payloads are the complete packet bodies after the packet ID. The capture also records downstream `session.spawned.chunk_radius=24` and upstream `session.spawned.chunk_radius=16`. This is a confirmed parity difference in the observed session, not evidence that 16/16 is invalid for every client.

## Source and implementation boundary

The in-tree gophertunnel copy at commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee` sends a default `RequestChunkRadius{16,16}` from [`minecraft/conn.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/conn.go) after it receives `ItemRegistry`. The local copy now also exposes an opt-in `Dialer.RequestChunkRadiusProvider` for deterministic client-side loopback tests. The provider keeps the original 16/16 fallback when it is absent, returns false, or returns a non-positive radius.

The proxy cannot wait for the real downstream request while dialing the upstream server. The upstream login and resource-pack exchange happen before the terminating bridge starts the downstream `StartGame` sequence. Waiting at that boundary caused a five-second timeout and preserved the old 16/16 request. The runtime fix therefore leaves the normal initial request in place and, when the downstream connection emits packet 69, decodes the request and writes the same typed values to the paired upstream connection. The downstream raw packet remains captured before this forwarding step. The focused loopback test verifies that the upstream observes the original 16/16 request followed by the downstream 24/28 request.

Forwarding failures are now recorded as `bridge.write_error` events with the
operation `forward RequestChunkRadius`. This includes the race where the
paired upstream connection is unavailable or closes between the downstream
packet observation and the typed write. The original downstream raw packet is
still retained, so a failed forward is visible rather than silently discarded.

If the observed body cannot be decoded as a complete `RequestChunkRadius`, the
proxy now records a `packet.decode_error` event with the packet ID and typed
context, including the packet header's sub-client routing bytes, before leaving
the raw packet untouched. This keeps strict forwarding validation visible
without treating an untrusted malformed body as a reason to rewrite or suppress
the packet.

## Classification

- Confirmed: packet ID 69 and the two payloads above differ in the owner-controlled capture.
- Confirmed: the current gophertunnel default is 16/16, and the listener can observe the downstream request before the bridge forwards ordinary gameplay packets.
- Confirmed: a failed later forward is represented in the capture as a bridge write error instead of being ignored by the listener packet callback.
- Confirmed: a malformed or trailing-byte later request is represented as a packet decode error with its raw packet still retained.
- Confirmed: the decode-error event preserves the request header's sender and target sub-client values alongside the raw packet.
- Inferred: forwarding a later `RequestChunkRadius` is the safest generic parity improvement available without redesigning the login/resource-pack barrier. Bedrock servers may retain their previously approved radius when handling a later request.
- Unknown: whether a particular public server uses the initial request, a later request, or an internal policy when selecting the approved radius. A public-server result requires a fresh stamped capture and must be reported separately from this local evidence.

## Validation boundary

Automated validation covers canonical encoding, malformed and trailing-byte rejection, the gophertunnel provider hook, and a loopback proxy session. No public-server validation is claimed by this note. The change does not mutate, drop, inject, replay, or suppress gameplay traffic.
