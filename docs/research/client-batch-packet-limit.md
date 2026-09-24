# Client batch packet-count limit

## Observation

The closed CubeCraft session on 2026-09-23 used stamped revision `bfc19102a312cbe5b0a18c9432c37afcbb446b24`. It completed resource-pack transfer and reached `session.spawned`, then the local downstream listener recorded `decode batch: number of packets exceeds max=812` and closed the client connection. The capture contains no server `Disconnect` packet at this point. Its final failing batch came from the client, and the batch payload was encrypted, so the exact packet count and member types are not available from the retained observation.

The confirmed immediate cause is a local decoder limit rejection, not a demonstrated CubeCraft kick or ban. The reason the client emitted a batch larger than 812 packets is unknown. A client-side backlog after a stall is a possible explanation, not an established fact.

## Source trace

In the pinned gophertunnel source, `packet.NewDecoder` applies a default maximum of 812 packets per batch. `minecraft.Listener.createConn` creates a connection with packet limits enabled, while `minecraft.Dialer` creates the upstream connection without those limits. The API provides `DisableBatchPacketLimit`, but removing the limit on the proxy's LAN-facing listener would leave packet-slice allocation unbounded by packet count.

- Vendored base revision: gophertunnel `v1.61.0`, commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`.
- Primary source: [pinned gophertunnel decoder](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/protocol/packet/decoder.go) and [pinned gophertunnel listener](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/listener.go).
- Local provenance: `third_party/gophertunnel/BEDROCKDEBUGPROXY_PATCH.md`.

## Implementation boundary

Keep 812 as the default for all existing decoder and listener users. Add an explicit positive per-listener override and set the proxy's client-facing listener to a finite 16,384 packets per batch. This caps the number of packet slice entries while leaving the existing payload/decompression limits and raw forwarding path unchanged. No packet is edited, skipped, or replayed. Values not explicitly set continue to use 812; upstream dialer behavior remains unchanged.

The 16,384 value is an implementation safety ceiling, not a claimed Bedrock protocol maximum. The capture proves only that the previous 812 ceiling was insufficient for this client batch; it does not reveal the exact count.

## Validation and follow-up

Synthetic decoder tests verify accepting a configured count, rejecting one packet above it, preserving the 812 default, and preserving the explicit disable API. Focused nested gophertunnel tests and `tools/quality.ps1 -LowMemory` passed, including root tests, build, lint, and vulnerability scanning. Live validation remains pending. Use a fresh stamped binary on The Hive, then repeat the same CubeCraft flow, confirming that no batch-limit decode error occurs and that the session remains active. Re-run the complete six-server matrix before release approval.
