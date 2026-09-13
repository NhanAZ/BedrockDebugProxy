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
- Exact implementation evidence is in [`minecraft/dial.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/dial.go), [`minecraft/listener.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/listener.go), [`minecraft/packet.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/packet.go), [`minecraft/conn.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/conn.go), and [`minecraft/network.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/network.go)
- The `Transfer` packet contract is defined in [`minecraft/protocol/packet/transfer.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/protocol/packet/transfer.go). It carries a hostname, UDP port, `ReloadWorld`, and optional gathering information; the client normally disconnects and joins the target.

Gophertunnel performs Bedrock login, Xbox authentication, encryption, compression, batching, packet framing, protocol conversion, resource-pack download, and spawn sequencing. Its included proxy demonstrates the standard pair of terminating connections and two forwarding loops.

Its [`minecraft/auth/live.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/auth/live.go) exposes `RefreshTokenSourceWriter`, which constructs a refresh-capable source from a previously issued Microsoft OAuth token and sends any interactive authentication instructions to the supplied writer. BedrockDebugProxy persists the token returned by that public API in its per-user cache and falls back to device authentication when refresh fails.

The public `Protocol` interface separates packet pools, readers, writers, and conversion from the current protocol. The project README states that one protocol is shipped at a time, while the API can host more. BedrockDebugProxy must therefore record the concrete protocol ID and version and must not assume a capture can always be decoded by a future default protocol.

`PacketFunc` receives packet headers and raw payloads on reads and writes. Unknown packet IDs can be represented by `packet.Unknown` when the listener and dialer are configured not to disconnect. Invalid packet decode errors are logged and the read loop may skip the packet. BedrockDebugProxy must record the pre-decode hook independently so skipped packets remain present.

The current resource-pack path downloads upstream packs, retains the advertised content key, supports a cache interface, and sends listener packs sequentially. `FetchResourcePacks` can choose packs after the connecting client's identity and client data are known. This enables a fixed-target proxy to establish the upstream connection during downstream login, capture the upstream packs, and then offer those packs to the client through public APIs.

The `Network` interface warns that wrappers must preserve optional packet transport capabilities. The first observing network supports RakNet only. NetherNet support must use a capability-preserving wrapper or a dedicated adapter rather than assuming ordinary `net.Conn` framing.

### Sandertv go-raknet

- Source reviewed at [`ea813dc668b5a2a2767cc5577bf77869c965f27a`](https://github.com/Sandertv/go-raknet/tree/ea813dc668b5a2a2767cc5577bf77869c965f27a)
- License is MIT
- Connection and latency evidence is in [`conn.go`](https://github.com/Sandertv/go-raknet/blob/ea813dc668b5a2a2767cc5577bf77869c965f27a/conn.go)

Go-raknet provides the reliable ordered connection below gophertunnel. Using it through the gophertunnel `Network` boundary keeps the dependency replaceable. Its connection exposes useful latency information, but the ordinary `net.Conn` stream is already above UDP datagrams and RakNet recovery behavior.

BedrockDebugProxy will record application-batch bytes at this boundary in the first version. Detailed RakNet frame capture is a separate layer and must not be implied by the initial event names or documentation.

### bedrock-tool/bedrocktool

- Original public repository reviewed at [`d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0`](https://github.com/bedrock-tool/bedrocktool/tree/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0)
- License is GPL-3.0

The upstream project is a broad Bedrock tooling suite rather than a narrow capture engine. Its breadth is useful for use-case discovery, especially authentication, Experience selection, transfer handling, downloads, and world data. Its architecture and scope are not a fit for direct reuse in the initial core. Its GPL-3.0 license is compatible with the selected project direction. No source was copied during the initial architecture implementation. Later adaptation of its legacy signaling code is recorded separately in `THIRD_PARTY_NOTICES.md` and `experience-routing.md`.

The repository persists an OAuth token and reconstructs a refresh source in [`utils/auth/auth.go`](https://github.com/bedrock-tool/bedrocktool/blob/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0/utils/auth/auth.go), [`utils/auth/account.go`](https://github.com/bedrock-tool/bedrocktool/blob/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0/utils/auth/account.go), and [`utils/auth/files.go`](https://github.com/bedrock-tool/bedrocktool/blob/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0/utils/auth/files.go). These files were reviewed on 2026-08-31 as evidence that persistent device authentication is established practice in this ecosystem. BedrockDebugProxy's cache was implemented independently around gophertunnel's public refresh API, with a project-specific path, size bound, replacement behavior, tests, and operator warnings.

### Follow-up authentication review

The current `bedrocktool` source was reviewed on 2026-09-13 at [`fcc057c418a28f1429411f0cd2d0c2a2bd3df7ba`](https://github.com/bedrock-tool/bedrocktool/tree/fcc057c418a28f1429411f0cd2d0c2a2bd3df7ba). Its GUI displays the verification URI and user code through an [`AuthCodeHandler`](https://github.com/bedrock-tool/bedrocktool/blob/fcc057c418a28f1429411f0cd2d0c2a2bd3df7ba/ui/messages/events.go), then opens `https://login.live.com/oauth20_remoteconnect.srf?otc=<code>` when the displayed URI is clicked in [`authpopup.go`](https://github.com/bedrock-tool/bedrocktool/blob/fcc057c418a28f1429411f0cd2d0c2a2bd3df7ba/ui/gui/popups/authpopup.go). That handler API comes from bedrocktool's pinned [`olebeck/gophertunnel` submodule at `b614cb65d41267941c29b2e0041e8cf6a7384271`](https://github.com/olebeck/gophertunnel/blob/b614cb65d41267941c29b2e0041e8cf6a7384271/minecraft/auth/live.go), not the Sandertv gophertunnel version used by this project.

BedrockDebugProxy keeps Sandertv gophertunnel's authentication and packet path unchanged and adapts its existing device-auth writer to print the same direct login URL, while retaining the original verification URI and code as a fallback. This is an independently implemented output adaptation, not copied bedrocktool source, and it does not alter the authentication protocol or packet path.

### Resource-pack URL behavior review

The current upstream `bedrocktool` source was reviewed on 2026-09-13 at [`85d5cfe1545c8d853be2859144a8357539ffc0f2`](https://github.com/bedrock-tool/bedrocktool/tree/85d5cfe1545c8d853be2859144a8357539ffc0f2), with the corresponding fork reviewed at [`6754869bac3214e7f7c30e86c382ecac57a516c1`](https://github.com/NhanAZ-Tools/bedrocktool/tree/6754869bac3214e7f7c30e86c382ecac57a516c1). Both revisions contain URL-aware resource-pack handling in [`utils/proxy/resourcepacks/resourcepacks.go`](https://github.com/bedrock-tool/bedrocktool/blob/85d5cfe1545c8d853be2859144a8357539ffc0f2/utils/proxy/resourcepacks/resourcepacks.go). The handler separates packs with `DownloadURL`, downloads them, applies the advertised content key, and sends `PackResponseAllPacksDownloaded` after URL and chunk downloads complete. This explains why bedrocktool can join Enchanted where a plain gophertunnel dialer requests chunks instead.

BedrockDebugProxy implements the same protocol-level outcome independently. Its per-connection `ResourcePackCache` observes the already-captured `ResourcePacksInfo` through gophertunnel's public `PacketFunc`, prefetches only the advertised HTTP(S) URL, validates UUID, version, and compressed size, and lets the gophertunnel state machine send its normal completion response. No bedrocktool source was copied or adapted. The URL remains in the reconstructed pack and in the raw capture, while failures remain visible through gophertunnel's cache warning and its normal chunk fallback. URL retrieval is limited to the current accepted connection and is not an offline pack downloader. The in-tree dependency copy carries one additional spawn-completion compatibility patch documented below.

### Featured-experience spawn compatibility

The fork's gophertunnel copy was reviewed on 2026-09-13 at [`cbeb266`](https://github.com/NhanAZ-Tools/bedrocktool/commit/cbeb266fbe9101f67d17840bb9b5e2749df4ace2). Its `minecraft/conn.go` marks `gameDataReceived` when `PlayStatusPlayerSpawn` arrives because Lifeboat and Enchanted may omit `ChunkRadiusUpdated`. Sandertv gophertunnel `v1.61.0` waits for both signals in `tryFinaliseClientConn`, so its dialer can otherwise remain blocked after the server has already sent `StartGame`, `ItemRegistry`, and `PlayStatusPlayerSpawn`. The current Enchanted capture `session-20260913T081031Z` records that exact sequence: no `ChunkRadiusUpdated` event appears between the upstream `RequestChunkRadius` and `PlayStatus`.

BedrockDebugProxy carries this three-line compatibility patch in the in-tree MIT-licensed gophertunnel copy under `third_party/gophertunnel`. It does not alter packet bytes, reorder traffic, or suppress unknown packets. The patch only permits the existing spawn acknowledgement and dial completion for a valid server sequence that omits the optional radius response. The regression test is `minecraft/conn_featured_experience_test.go` in that dependency copy.

### Featured-experience transfer route behavior

The current `NhanAZ-Tools/bedrocktool` fork was also reviewed at [`6754869bac3214e7f7c30e86c382ecac57a516c1`](https://github.com/NhanAZ-Tools/bedrocktool/tree/6754869bac3214e7f7c30e86c382ecac57a516c1) on 2026-09-13. Its transfer path preserves the previous RakNet UDP source address and bounds the status ping before dialing the next target. This is relevant for featured-experience frontends that assign a backend using UDP flow state. Starting a fresh generic dial only after the downstream reconnect can otherwise leave the client waiting for resource packs while the proxy is still blocked before the next upstream connection opens.

The current proxy now records the upstream source address on `Transfer`, probes the next RakNet target with that source address before waiting for the client reconnect, and reuses the source address for the next dial while skipping a second status ping. This is an independently implemented, narrow adaptation around the public gophertunnel and go-raknet APIs. The selected go-raknet revision does not expose the client GUID, so GUID preservation is not claimed here. Route-probe success or failure is recorded as `transfer.route_probe`; a failed probe does not discard the normal dial attempt. This behavior is limited to the default RakNet transport.

The Enchanted capture `session-20260913T143415Z` demonstrated the failure that motivated this change: hop 1 completed `Transfer`, but hop 2 had no upstream transport-open or Bedrock packet before the downstream resource-pack login waited and disconnected. A future live validation must confirm `upstream-2` transport-open, `upstream.connected`, resource-pack exchange, and spawn after a server transfer.

### PrismarineJS bedrock-protocol

- Source reviewed at [`6011e261c1b028f92d732348dc91339fb12275dd`](https://github.com/PrismarineJS/bedrock-protocol/tree/6011e261c1b028f92d732348dc91339fb12275dd)
- License is MIT
- Layering evidence is in [`src/rak.js`](https://github.com/PrismarineJS/bedrock-protocol/blob/6011e261c1b028f92d732348dc91339fb12275dd/src/rak.js), [`src/transforms/framer.js`](https://github.com/PrismarineJS/bedrock-protocol/blob/6011e261c1b028f92d732348dc91339fb12275dd/src/transforms/framer.js), [`src/transforms/encryption.js`](https://github.com/PrismarineJS/bedrock-protocol/blob/6011e261c1b028f92d732348dc91339fb12275dd/src/transforms/encryption.js), and [`src/relay.js`](https://github.com/PrismarineJS/bedrock-protocol/blob/6011e261c1b028f92d732348dc91339fb12275dd/src/relay.js)

This project separates RakNet, framing, compression, encryption, generated protocol codecs, client and server sessions, and relay behavior. Its generated schemas and multi-version ecosystem are valuable comparison points for packet definitions and differential decode tests.

The JavaScript runtime is not required for the initial proxy. BedrockDebugProxy should later use it as an independent decoder in compatibility tests rather than coupling the main capture path to two protocol stacks.

### Kas-tle ProxyPass

- Kas-tle source reviewed at [`baa6d9a565c58f5f2306c5646fc833704253109f`](https://github.com/Kas-tle/ProxyPass/tree/baa6d9a565c58f5f2306c5646fc833704253109f)
- License is AGPL-3.0

This fork shows a mature terminating-proxy feature surface that includes online authentication, transfer following, featured experiences, Realms, friend sessions, RakNet, NetherNet, packet inspection, protocol testing, and resource-pack download and decryption. Its README and [`PackDownloader.java`](https://github.com/Kas-tle/ProxyPass/blob/baa6d9a565c58f5f2306c5646fc833704253109f/src/main/java/org/cloudburstmc/proxypass/network/bedrock/session/PackDownloader.java) support those specific research claims.

The fork maintains modified protocol and network submodules. That improves its ability to debug those libraries but also demonstrates the maintenance cost of deep stack forks. BedrockDebugProxy will begin with public gophertunnel extension points and add a narrow adapter or fork only when a documented fidelity requirement cannot be met otherwise.

No ProxyPass source will be copied without an explicit decision to accept the additional AGPL network-use obligations and record the resulting provenance. The current implementation uses ProxyPass only as a research reference.

### CloudburstMC ProxyPass

- Current source reviewed at [`b92c4d4f88ee18643df2e5df26a15f8ef4a7db00`](https://github.com/CloudburstMC/ProxyPass/tree/b92c4d4f88ee18643df2e5df26a15f8ef4a7db00)
- License is AGPL-3.0

The original CloudburstMC project remains evidence for the basic terminating-proxy lineage. Its current tree does not contain the Kas-tle `PackDownloader.java` path and is not used as evidence for resource-pack decryption behavior. Treat the two repositories as separate revisions with different feature surfaces.

### EndstoneMC spyglass

- Source reviewed at [`b4f5f7f879ec0089d353093ee15e6770caf7c1b0`](https://github.com/EndstoneMC/spyglass/tree/b4f5f7f879ec0089d353093ee15e6770caf7c1b0)
- License is MIT
- Capture evidence is in [`network.cpp`](https://github.com/EndstoneMC/spyglass/blob/b4f5f7f879ec0089d353093ee15e6770caf7c1b0/src/spyglass/network.cpp), [`capture.cpp`](https://github.com/EndstoneMC/spyglass/blob/b4f5f7f879ec0089d353093ee15e6770caf7c1b0/src/spyglass/overlay/capture.cpp), [`store.cpp`](https://github.com/EndstoneMC/spyglass/blob/b4f5f7f879ec0089d353093ee15e6770caf7c1b0/src/spyglass/overlay/store.cpp), and [`packet_details.cpp`](https://github.com/EndstoneMC/spyglass/blob/b4f5f7f879ec0089d353093ee15e6770caf7c1b0/src/spyglass/overlay/pane/packet_details.cpp)

Spyglass hooks the Bedrock client at packet send and packet read boundaries. It retains raw packet bodies, unread byte counts, decode success, detailed error trees, packet names, sub-client information, and timing. Its UI is backed by a disk stream plus an index rather than an unbounded in-memory packet list.

The bounded writer queue and visible rejected or dropped counters are strong design references. BedrockDebugProxy will make any loss visible in the manifest and event stream. Its default capture path will favor blocking writes over silent loss until measured performance justifies a more complex queue.

### EndstoneMC protocol-dumper

- Source reviewed at [`ea87a290a8154d850f41fef8aaffa0bbe7ebfbd4`](https://github.com/EndstoneMC/protocol-dumper/tree/ea87a290a8154d850f41fef8aaffa0bbe7ebfbd4)
- GitHub reports no repository-level license at this revision
- Schema extraction evidence is in [`src/main.cpp`](https://github.com/EndstoneMC/protocol-dumper/blob/ea87a290a8154d850f41fef8aaffa0bbe7ebfbd4/src/main.cpp), [`src/visitor.cpp`](https://github.com/EndstoneMC/protocol-dumper/blob/ea87a290a8154d850f41fef8aaffa0bbe7ebfbd4/src/visitor.cpp), and [`src/models.h`](https://github.com/EndstoneMC/protocol-dumper/blob/ea87a290a8154d850f41fef8aaffa0bbe7ebfbd4/src/models.h)

The project extracts packet, struct, enum, and serialization schemas from a running Bedrock Dedicated Server. Runtime schema extraction is a valuable future input for update validation and field-name enrichment.

Because the repository-level license is absent, BedrockDebugProxy will use only the observed concept and public output where permitted. It will not copy source.

### MrSterdy bedrock-packet-interceptor

- Source reviewed at [`fd933b85c3672b368df6fd674f5842d7507d338a`](https://github.com/MrSterdy/bedrock-packet-interceptor/tree/fd933b85c3672b368df6fd674f5842d7507d338a)
- GitHub reports no license
- Relay and UI event evidence is in [`src/lib/server/proxy.ts`](https://github.com/MrSterdy/bedrock-packet-interceptor/blob/fd933b85c3672b368df6fd674f5842d7507d338a/src/lib/server/proxy.ts), [`src/lib/server/emitter.ts`](https://github.com/MrSterdy/bedrock-packet-interceptor/blob/fd933b85c3672b368df6fd674f5842d7507d338a/src/lib/server/emitter.ts), and [`src/routes/api/events/+server.ts`](https://github.com/MrSterdy/bedrock-packet-interceptor/blob/fd933b85c3672b368df6fd674f5842d7507d338a/src/routes/api/events/%2Bserver.ts)

This project combines a PrismarineJS relay with a Svelte web interface and streams packet events to the UI. It is useful evidence that interactive filtering and readable decoded fields matter, but an in-memory event emitter and browser view are not a durable canonical capture.

No source will be copied without a compatible license.

### Endermanbugzjfc PacketLoggerGophertunnel

- Source reviewed at [`e718c1432397bb66fa95c5188ca239617a2c7298`](https://github.com/Endermanbugzjfc/PacketLoggerGophertunnel/tree/e718c1432397bb66fa95c5188ca239617a2c7298)
- License is Apache-2.0
- The last source push reported by GitHub was 2023-08-29

This project is a small example of logging packets around a gophertunnel proxy. It validates the basic forwarding approach, but its age and log-oriented output do not satisfy current protocol support or lossless capture requirements.

### Dragonfly

- Source reviewed at [`a36ed0edb548298ab482939e1c653e39f9683719`](https://github.com/df-mc/dragonfly/tree/a36ed0edb548298ab482939e1c653e39f9683719), tagged `v0.11.4`
- License is MIT
- Server fixture evidence is in [`server/server.go`](https://github.com/df-mc/dragonfly/blob/a36ed0edb548298ab482939e1c653e39f9683719/server/server.go) and [`server/player/player.go`](https://github.com/df-mc/dragonfly/blob/a36ed0edb548298ab482939e1c653e39f9683719/server/player/player.go)

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
- `--follow-transfers` starts a new hop after a server `Transfer` by rewriting the client destination to the local listener. The original target is then dialled after client reconnect, while the capture keeps one logical session timeline. See ADR 0003 for the trust boundary and concrete-listener requirement.
- Continue measuring disk-write backpressure during chunk-heavy sessions before considering any asynchronous queue or overflow policy. The current recorder uses blocking, lossless writes and avoids per-event durable flushes by default.
- Verify encrypted resource-pack variants with synthetic fixtures and owner-authorized live captures.
- Decide whether old protocol adapters belong in this repository or separate versioned modules.

The external source audit and the correction of the stale CloudburstMC reference are recorded in [`reference-audit.md`](reference-audit.md).
