# ADR 0001 - Capture-first terminating proxy

## Status

Accepted on 2026-08-30.

## Context

BedrockDebugProxy must authenticate to online Bedrock servers, terminate the client and server sessions, inspect decoded packets, preserve raw evidence, collect resource packs, and survive frequent protocol changes.

Implementing RakNet, Xbox authentication, Bedrock encryption, compression, framing, packet definitions, and resource-pack negotiation from scratch would postpone useful capture work. Treating a high-level packet logger as the whole architecture would make failures and lower-layer evidence impossible to recover.

## Decision

The project owns a protocol-neutral capture model and uses gophertunnel `v1.61.0` as its first Bedrock adapter.

The proxy terminates two independent sessions.

```text
Minecraft client
  <-> downstream RakNet and Bedrock session
BedrockDebugProxy bridge
  <-> upstream RakNet and Bedrock session
Destination server
```

Each observable representation is recorded as a separate ordered event.

- Transport payload events hold application-batch bytes at the `minecraft.Network` boundary. They are after RakNet reassembly on reads and before RakNet fragmentation on writes.
- Raw packet events hold the packet header and exact payload exposed by gophertunnel after inbound decrypt, decompress, and frame processing or before outbound batching, compression, and encryption.
- Decoded packet events hold the concrete Go packet type and a JSON-safe field tree.
- Lifecycle, error, resource-pack, limit, and annotation events share the same timeline.

Raw bytes are stored in content-addressed blobs. Events refer to blobs by SHA-256 digest and size. The event sequence is authoritative when timestamps are equal.

The first capture form is a directory instead of a single archive. Append-only files and independently durable blobs are easier to recover after a crash. A later export command may create an immutable `.bdpcap` archive without changing the logical schema.

Gophertunnel is isolated in the Bedrock adapter and proxy packages. The capture package does not import it. A future decoder, protocol version, replay tool, or non-Go analyzer can consume the same capture without loading gophertunnel.

## Fidelity rules

Decoded data never replaces raw data. A missing decode creates an explicit error or unknown event and does not remove the raw event.

The recorder uses one bounded ordered asynchronous writer by default. Capture submission blocks only when the queue reaches its configured capacity. The writer never drops, samples, truncates, or reorders accepted events. Queue capacity, blocking policy, loss policy, and peak occupancy are recorded in manifest values. A writer failure marks the capture incomplete and is surfaced to the forwarding path.

Every event receives a process-monotonic elapsed time, UTC wall time, and monotonically increasing sequence number. Connection, hop, channel, endpoints, and logical direction remain explicit.

Transport payloads are not called raw UDP or raw RakNet. Packet payloads are not called encrypted batches. Event stages describe only the boundary that actually produced the bytes.

## Resource packs

For a fixed upstream target, the listener's `FetchResourcePacks` hook may establish the upstream connection while downstream login is waiting. The downloaded upstream packs can then be recorded and offered to the client using public gophertunnel APIs.

The original archive, advertised metadata, content key, chunk observations, reconstruction result, hashes, and any derived decrypted tree are separate artifacts. Decryption never overwrites the captured archive.

## Consequences

The first useful version can support current Bedrock authentication and packets while the project concentrates on capture integrity and analysis.

The first version cannot claim raw UDP or complete RakNet diagnostics. It also relies on the protocol version shipped by its selected gophertunnel adapter. Both limitations are recorded in capture metadata and documentation.

Deep instrumentation remains possible through a narrow future fork, custom network adapter, or OS capture source. Those changes do not require replacing the event model.

The default proxy is intentionally observational. Packet dropping, injection, and replay do not belong in the bridge. The separately documented `--follow-transfers` mode is the narrow exception for routing a Bedrock `Transfer` back through the listener so that the next hop can remain observable; it preserves the original packet and is disabled by default.
