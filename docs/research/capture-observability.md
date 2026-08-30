# Capture observability review

## Scope

This review was performed on 2026-08-30 to answer one narrow question. Does the current one-session proxy discard useful evidence that its existing connection and capture architecture can already observe?

The review covered BedrockDebugProxy call sites, gophertunnel `v1.61.0` at commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`, and an owner-controlled local bedrocktool checkout at commit `c947abe1dbb334b27466da51642d9d4e7b6f88f5`. No source was copied from bedrocktool or any external project.

## Gophertunnel observation boundaries

The pinned gophertunnel source establishes the following boundaries.

- [`minecraft/dial.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/dial.go) and [`minecraft/listener.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/listener.go) document that `PacketFunc` observes reads and writes, including packets consumed by the connection sequence such as Login.
- [`minecraft/packet.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/packet.go) calls the hook after reading the packet header and before high-level decode. Unknown or invalid packets therefore remain capturable even when later decoding fails.
- [`minecraft/conn.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/conn.go) exposes `IdentityData`, `ClientData`, authentication state, `GameData`, latency, client-cache state, and chunk radius. Login and StartGame are partly consumed before normal forwarding, so these decoded values were observable but not previously indexed as structured session evidence.
- [`minecraft/network.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/network.go) confirms that the current network wrapper is above the reliable transport connection. It cannot prove UDP datagrams, RakNet acknowledgements, fragmentation, loss, or retransmission behavior.

## Local bedrocktool comparison

The local checkout was used for use-case comparison, not protocol authority. Its `handlers/capture.go` retains raw packet records and downloaded resource-pack archives, while `utils/proxy/session.go` handles pre-play packets, StartGame state, cache blobs, transfers, and gameplay packets. It also contains specialized world, skin, and utility handlers.

BedrockDebugProxy already retains the raw packet payload and decoded view for every packet visible through the gophertunnel hook, including traffic that broad tools may filter as noisy. It also retains transport payloads, resource-pack archives, cache-related packets, errors, unknown packets, and Transfer packets. Copying the specialized command surface would not improve the canonical session and would conflict with the project's narrow CLI.

The clear local capture gaps were decoded connection metadata and upstream GameData. The implementation now records them as bounded derived views while preserving the exact Login and StartGame packet payloads as authoritative raw blobs. Large collections may be truncated according to the existing decoded-view limit. Spawn-time latency, cache state, and chunk radius are also retained because the current connections already expose them.

## Remaining limits

- Transfer packets are retained but automatic hop following is not implemented.
- Raw UDP and RakNet recovery evidence requires a lower observation boundary.
- The transport payload event does not yet classify encryption or compression state per event.
- Decoded snapshots reflect gophertunnel's active protocol model. Future analysis must prefer raw packet blobs when checking a changed or disputed definition.
- Full login metadata is sensitive. Structured snapshots make selected identifiers easier to search even though the raw Login packet was already present.

These are explicit limits, not a reason to add specialized download commands or duplicate an external proxy architecture.
