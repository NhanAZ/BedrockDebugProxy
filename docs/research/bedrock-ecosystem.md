# Bedrock networking ecosystem research

This document records the initial source review used to choose the BedrockDebugProxy architecture. The review was performed on 2026-08-30. Commit identifiers matter because Bedrock libraries and packet definitions change frequently.

## Conclusions

The first implementation will use Go and gophertunnel as a replaceable session and protocol adapter. The project will own its capture model, storage, observation hooks, decoded representation, resource-pack artifacts, CLI, and analysis workflow.

Gophertunnel exposes two useful observation boundaries without requiring a source fork.

1. `minecraft.Network` creates the reliable transport connection used by the Minecraft packet encoder and decoder. Wrapping its `net.Conn` preserves each application batch as it enters or leaves gophertunnel. These bytes are still compressed or encrypted when those features are active. They are after RakNet reassembly on reads and before RakNet fragmentation on writes.
2. `ListenConfig.PacketFunc` and `Dialer.PacketFunc` receive a packet header and raw packet payload. This hook runs after inbound batch decryption, decompression, and framing, and before high-level decode. It also observes login-sequence packets that normal proxy forwarding code does not receive.

The two boundaries complement each other. Neither is a substitute for raw UDP or PCAP capture. A later transport adapter or OS-level capture can add RakNet datagram, acknowledgement, fragmentation, loss, and retransmission evidence without changing the canonical event model.

The canonical capture will be an append-only directory containing a manifest, an ordered JSON Lines event stream, and content-addressed raw blobs. Decoded packet fields are a derived view. Raw observations, decode failures, unknown packets, session state, resource packs, and completeness counters remain first-class data.

## Source review

### Sandertv gophertunnel

- Source reviewed at [`283a5a97dfe65da94bcc0b401807f6aefa9e72ee`](https://github.com/Sandertv/gophertunnel/tree/283a5a97dfe65da94bcc0b401807f6aefa9e72ee), tagged `v1.61.0`
- License is MIT
- Current packet constants at this revision are protocol `2169` and game version `1.26.45`
- The module requires Go 1.25 or newer

Gophertunnel performs Bedrock login, Xbox authentication, encryption, compression, batching, packet framing, protocol conversion, resource-pack download, and spawn sequencing. Its included proxy demonstrates the standard pair of terminating connections and two forwarding loops.

The public `Protocol` interface separates packet pools, readers, writers, and conversion from the current protocol. The project README states that one protocol is shipped at a time, while the API can host more. BedrockDebugProxy must therefore record the concrete protocol ID and version and must not assume a capture can always be decoded by a future default protocol.

`PacketFunc` receives packet headers and raw payloads on reads and writes. Unknown packet IDs can be represented by `packet.Unknown` when the listener and dialer are configured not to disconnect. Invalid packet decode errors are logged and the read loop may skip the packet. BedrockDebugProxy must record the pre-decode hook independently so skipped packets remain present.

The current resource-pack path downloads upstream packs, retains the advertised content key, supports a cache interface, and sends listener packs sequentially. `FetchResourcePacks` can choose packs after the connecting client's identity and client data are known. This enables a fixed-target proxy to establish the upstream connection during downstream login, capture the upstream packs, and then offer those packs to the client through public APIs.

The `Network` interface warns that wrappers must preserve optional packet transport capabilities. The first observing network supports RakNet only. NetherNet support must use a capability-preserving wrapper or a dedicated adapter rather than assuming ordinary `net.Conn` framing.

### Sandertv go-raknet

- Source reviewed at [`ea813dc668b5a2a2767cc5577bf77869c965f27a`](https://github.com/Sandertv/go-raknet/tree/ea813dc668b5a2a2767cc5577bf77869c965f27a)
- License is MIT

Go-raknet provides the reliable ordered connection below gophertunnel. Using it through the gophertunnel `Network` boundary keeps the dependency replaceable. Its connection exposes useful latency information, but the ordinary `net.Conn` stream is already above UDP datagrams and RakNet recovery behavior.

BedrockDebugProxy will record application-batch bytes at this boundary in the first version. Detailed RakNet frame capture is a separate layer and must not be implied by the initial event names or documentation.

### NhanAZ-Tools bedrocktool

- Private owner-controlled source reviewed locally at `c947abe1dbb334b27466da51642d9d4e7b6f88f5`
- Upstream project license is GPL-3.0

The fork demonstrates practical requirements that are easy to miss in a minimal proxy. It keeps hooks for pre-play packets, resource-pack handling, packet timing, transfers, and raw capture. Its resource-pack work shows that downstream pack announcements must remain sequential and that large packs need bounded queues and explicit completion handling.

The fork also keeps a detailed resource-pack decryption investigation. BedrockDebugProxy will independently implement and test the protocol behavior it needs. No source from the fork was copied into the initial implementation. Any future reuse must be deliberate, compatible with `GPL-3.0-or-later`, and recorded with exact provenance.

The local fork pins a patched Go toolchain for a Windows asynchronous `WSARecvFrom` stability issue. BedrockDebugProxy will use `toolchain go1.26.6` while the development host reports Go 1.26.1, then re-evaluate the pin when the project has its own long-running Windows tests.

### bedrock-tool bedrocktool

- Repository reviewed at the current owner fork lineage and upstream metadata
- License is GPL-3.0

The upstream project is a broad Bedrock tooling suite rather than a narrow capture engine. Its breadth is useful for use-case discovery, especially authentication, transfer handling, downloads, and world data. Its architecture and scope are not a fit for direct reuse in the initial core. Its GPL-3.0 license is compatible with the selected project direction, but no source was copied.

### PrismarineJS bedrock-protocol

- Source reviewed at [`6011e261c1b028f92d732348dc91339fb12275dd`](https://github.com/PrismarineJS/bedrock-protocol/tree/6011e261c1b028f92d732348dc91339fb12275dd)
- License is MIT

This project separates RakNet, framing, compression, encryption, generated protocol codecs, client and server sessions, and relay behavior. Its generated schemas and multi-version ecosystem are valuable comparison points for packet definitions and differential decode tests.

The JavaScript runtime is not required for the initial proxy. BedrockDebugProxy should later use it as an independent decoder in compatibility tests rather than coupling the main capture path to two protocol stacks.

### Kas-tle ProxyPass and CloudburstMC ProxyPass

- Kas-tle source reviewed at [`baa6d9a565c58f5f2306c5646fc833704253109f`](https://github.com/Kas-tle/ProxyPass/tree/baa6d9a565c58f5f2306c5646fc833704253109f)
- CloudburstMC repository metadata reviewed on 2026-08-30
- Both repositories are AGPL-3.0

ProxyPass shows a mature terminating-proxy feature surface that includes online authentication, transfer following, featured experiences, Realms, friend sessions, RakNet, NetherNet, packet inspection, protocol testing, and resource-pack download and decryption.

The fork maintains modified protocol and network submodules. That improves its ability to debug those libraries but also demonstrates the maintenance cost of deep stack forks. BedrockDebugProxy will begin with public gophertunnel extension points and add a narrow adapter or fork only when a documented fidelity requirement cannot be met otherwise.

No ProxyPass source will be copied without an explicit decision to accept the additional AGPL network-use obligations and record the resulting provenance. The current implementation uses ProxyPass only as a research reference.

### EndstoneMC spyglass

- Source reviewed at [`b4f5f7f879ec0089d353093ee15e6770caf7c1b0`](https://github.com/EndstoneMC/spyglass/tree/b4f5f7f879ec0089d353093ee15e6770caf7c1b0)
- License is MIT

Spyglass hooks the Bedrock client at packet send and packet read boundaries. It retains raw packet bodies, unread byte counts, decode success, detailed error trees, packet names, sub-client information, and timing. Its UI is backed by a disk stream plus an index rather than an unbounded in-memory packet list.

The bounded writer queue and visible rejected or dropped counters are strong design references. BedrockDebugProxy will make any loss visible in the manifest and event stream. Its default capture path will favor blocking writes over silent loss until measured performance justifies a more complex queue.

### EndstoneMC protocol-dumper

- Source reviewed at [`ea87a290a8154d850f41fef8aaffa0bbe7ebfbd4`](https://github.com/EndstoneMC/protocol-dumper/tree/ea87a290a8154d850f41fef8aaffa0bbe7ebfbd4)
- GitHub reports no repository-level license at this revision

The project extracts packet, struct, enum, and serialization schemas from a running Bedrock Dedicated Server. Runtime schema extraction is a valuable future input for update validation and field-name enrichment.

Because the repository-level license is absent, BedrockDebugProxy will use only the observed concept and public output where permitted. It will not copy source.

### MrSterdy bedrock-packet-interceptor

- Source reviewed at [`fd933b85c3672b368df6fd674f5842d7507d338a`](https://github.com/MrSterdy/bedrock-packet-interceptor/tree/fd933b85c3672b368df6fd674f5842d7507d338a)
- GitHub reports no license

This project combines a PrismarineJS relay with a Svelte web interface and streams packet events to the UI. It is useful evidence that interactive filtering and readable decoded fields matter, but an in-memory event emitter and browser view are not a durable canonical capture.

No source will be copied without a compatible license.

### Endermanbugzjfc PacketLoggerGophertunnel

- Repository metadata and source layout reviewed on 2026-08-30
- License is Apache-2.0
- The last source push reported by GitHub was 2023-08-29

This project is a small example of logging packets around a gophertunnel proxy. It validates the basic forwarding approach, but its age and log-oriented output do not satisfy current protocol support or lossless capture requirements.

### Dragonfly

- Source reviewed at [`a36ed0edb548298ab482939e1c653e39f9683719`](https://github.com/df-mc/dragonfly/tree/a36ed0edb548298ab482939e1c653e39f9683719), tagged `v0.11.4`
- License is MIT

Dragonfly is a server implementation built around gophertunnel rather than a packet proxy. It is useful for local integration fixtures and for understanding how the protocol library is exercised by a complete server. BedrockDebugProxy does not need the full server dependency in its capture core.

## Architecture alternatives

| Approach | Strengths | Costs and risks | Decision |
| --- | --- | --- | --- |
| Build every layer from UDP upward | Maximum control and theoretical fidelity | Authentication, encryption, compression, RakNet, protocol updates, and resource packs would delay useful captures and duplicate mature work | Rejected for the first implementation |
| Fork a complete existing proxy | Fast access to mature features | Inherits architecture, license, update burden, and unrelated behavior | Rejected as the default |
| Use gophertunnel only and print decoded packets | Small implementation | Loses durable raw evidence, failures, provenance, and analysis structure | Rejected |
| Own the capture architecture and use gophertunnel through adapters | Early working proxy, public hooks, raw packet payloads, replaceable protocol engine | Does not initially expose raw UDP or every internal codec transition | Accepted |

## Open research items

- Preserve NetherNet transport capabilities while observing its message boundary.
- Add optional raw UDP and PCAPNG capture without requiring elevated privileges for normal operation.
- Compare decoded fixtures against PrismarineJS and future runtime schemas.
- Determine how Transfer packets should start a new hop while keeping one logical session timeline.
- Measure disk-write backpressure during chunk-heavy sessions and choose an explicit overflow policy.
- Verify encrypted resource-pack variants with synthetic fixtures and owner-authorized live captures.
- Decide whether old protocol adapters belong in this repository or separate versioned modules.
