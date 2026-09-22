# Capture observability review

## Scope

This review was performed on 2026-08-30 to answer one narrow question. Does the current one-session proxy discard useful evidence that its existing connection and capture architecture can already observe?

The review covered BedrockDebugProxy call sites, gophertunnel `v1.61.0` at commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`, and the original [`bedrock-tool/bedrocktool`](https://github.com/bedrock-tool/bedrocktool/tree/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0) repository at commit `d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0`. No source was copied from bedrocktool or any external project for this observability change.

## Gophertunnel observation boundaries

The pinned gophertunnel source establishes the following boundaries.

- [`minecraft/dial.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/dial.go) and [`minecraft/listener.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/listener.go) document that `PacketFunc` observes reads and writes, including packets consumed by the connection sequence such as Login.
- [`minecraft/packet.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/packet.go) calls the hook after reading the packet header and before high-level decode. Unknown or invalid packets therefore remain capturable even when later decoding fails.
- [`minecraft/conn.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/conn.go) exposes `IdentityData`, `ClientData`, authentication state, `GameData`, latency, client-cache state, and chunk radius. Login and StartGame are partly consumed before normal forwarding, so these decoded values were observable but not previously indexed as structured session evidence.
- [`minecraft/network.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/network.go) confirms that the current network wrapper is above the reliable transport connection. It cannot prove UDP datagrams, RakNet acknowledgements, fragmentation, loss, or retransmission behavior.

### Protocol conversion and pre-bridge packets

The current protocol implementation in the in-tree gophertunnel composite uses an
identity `ConvertToLatest` and `ConvertFromLatest` path for protocol 2193. A
normal same-protocol packet therefore remains eligible for the raw read and raw
write boundary described below. The connection API still permits a protocol
adapter to return more than one latest-protocol packet for one input packet.

`ReadPacketDataWithTime` returns a typed `PacketRead.Packet` without `Data` when
the connection has already decoded a packet during its login state machine or
has queued an additional packet produced by protocol conversion. Such a packet
cannot be reconstructed as raw bytes at that later bridge read. This does not
erase the canonical evidence: the gophertunnel packet hook runs after the packet
header is read and before decode, so BedrockDebugProxy has already recorded the
original raw packet when the hook was available. The bridge forwards the typed
value in this case and the capture keeps the decoded view separate from the raw
event.

Confirmed for the current tree: protocol 2193 does not currently produce
additional packets through conversion. Confirmed as an API boundary: a future
version adapter may produce them, and the dependency associates the original raw
packet only with the first converted value. Unknown: whether any supported
future protocol will use that multi-packet conversion path. The project must not
fabricate raw bytes for a typed additional packet; a future implementation that
uses this path should add an explicit derived-view annotation and regression
fixture before changing forwarding behavior.

### Deferred-login ordering evidence

The owner-controlled capture `captures/session-20260921T122228Z` at stamped
revision `d272805fd571525cca6f7473b166fb1129bafd48` confirms the temporal
separation between raw observation and bridge decoding. The downstream packet
hook recorded `ServerBoundLoadingScreen` (ID 312), `ClientCacheStatus` (ID
113), and several `PlayerAuthInput` packets (ID 144) before
`session.spawned`. The bridge later emitted a decoded `ServerBoundLoadingScreen`
event after `session.spawned`, followed by decoded gameplay packets.

This is expected behavior, not evidence that the raw packet was reordered or
lost. `parseData` invokes the packet hook before the login state machine
decodes a packet. The login state machine can defer that packet until the
current expected state accepts it, after which `ReadPacketDataWithTime` returns
the retained raw packet to the bridge. The recorder's sequence is therefore the
ordered submission stream of concurrent observer and bridge callbacks, not a
promise that every raw event is immediately adjacent to its later decoded view.
The raw blob and decoded event remain separately authoritative for their
respective boundaries.

Confirmed for this capture: raw hook arrival preceded bridge consumption for
the deferred login packets. Inferred: the ID 312 raw event before
`session.spawned` is the source of the later decoded ID 312 event, based on
packet identity, connection, direction, and the gophertunnel deferred queue.
Unknown: whether a future protocol conversion or server ordering will produce
the same separation for other packet families. Analysis must not pair raw and
decoded events solely by adjacent sequence numbers or fabricate a new raw
event at the decode boundary.

## Public bedrocktool comparison

The original public repository was used for use-case comparison, not protocol authority. Its [`handlers/capture.go`](https://github.com/bedrock-tool/bedrocktool/blob/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0/handlers/capture.go) retains raw packet records and downloaded resource-pack archives, while [`utils/proxy/session.go`](https://github.com/bedrock-tool/bedrocktool/blob/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0/utils/proxy/session.go) handles pre-play packets, StartGame state, cache blobs, transfers, and gameplay packets. It also contains specialized world, skin, and utility handlers.

BedrockDebugProxy retains the raw packet payload for every packet visible through the gophertunnel hook, including traffic that broad tools may filter as noisy. Packets that reach the forwarding bridge also receive decoded views. Packets consumed by gophertunnel during connection setup remain available as raw evidence, with decoded connection metadata and GameData recorded as selected session snapshots. The capture also retains transport payloads, resource-pack archives, cache-related packets, errors, unknown packets, and Transfer packets. Copying the specialized command surface would not improve the canonical session and would conflict with the project's narrow CLI.

## Forward-path timing

The bridge now records `bridge.forward_timing` for packets that can affect world collision and spawn state: `LevelChunk`, block updates, `NetworkChunkPublisherUpdate`, `ChunkRadiusUpdated`, `MovePlayer`, `SetActorMotion`, and `Transfer`. Each timing event references its decoded packet with `parent_sequence` and measures the raw or typed read, decoded-view capture submission, and write calls. Capture submission includes queue backpressure but not the later writer commit. Its `write_mode` identifies whether the bridge used the byte-preserving raw path or typed serialization. A same-protocol packet normally uses the raw path so the exact packet header and payload are preserved; a transfer uses the typed path only when `--follow-transfers` must rewrite its address. A transfer also measures its explicit flush. The normal gophertunnel flush runs asynchronously at the configured flush rate, so its completion is represented by the subsequent `transport.payload` event rather than by forcing a per-packet flush that would change the wire behavior under investigation.

These timings distinguish proxy-side read, capture submission, and forwarding delay from client-side chunk application. Closed-capture manifest values expose the writer queue peak, and the event stream remains the source of truth for commit order. They do not expose raw UDP datagrams, RakNet acknowledgements, or a client acknowledgement that collision data was applied. A suspected missing cage or void spawn still requires comparison with a direct-client capture and inspection of the corresponding `LevelChunk` and block-update payloads.

The clear local capture gaps were decoded connection metadata and upstream GameData. The implementation now records them as bounded derived views while preserving the exact Login and StartGame packet payloads as authoritative raw blobs. Large collections may be truncated according to the existing decoded-view limit. Spawn-time latency, cache state, and chunk radius are also retained because the current connections already expose them.

Transport payload events now supplement the application-batch blob with per-connection Bedrock boundary state as soon as the connection object is created, including during login. `raknet_client_guid` is retained when the active connection exposes a non-zero RakNet client identity. `bedrock_compression_state` is `negotiated` or `not_negotiated`, and negotiated events include the numeric compression threshold and algorithm. `bedrock_encryption_state` is `enabled`, `disabled`, or `not_enabled`. These fields describe the connection state observed by the proxy. The GUID is sensitive transport identity metadata and remains in the local capture. These fields do not decode RakNet datagrams, identify a cipher outside the connection API, or replace the raw transport payload.

Transport lifecycle events also retain the optional RakNet GUID and public-key capability at connection open or close when the underlying `net.Conn` exposes them. Only public-key capability and presence booleans are stored. This records the identity boundary needed for local parity analysis without persisting the key itself.

The observing transport wrapper preserves error provenance at the read and write boundary. If the underlying transport returns a partial transfer or operation error while the recorder also fails, the wrapper retains the transferred byte count and joins the transport and capture errors. The event still records the original operation error and raw bytes that were available. This keeps a recorder failure from masking a transport failure or changing a partial-read result.

The same rule applies when a connection-open event cannot be recorded. The wrapper still closes the newly opened transport and joins any close error with the recorder failure, so a capture failure does not hide a lower-layer shutdown problem.

Transfer rewrite failures follow the same error-preservation rule. The bridge releases a retained raw packet only when one exists, records `bridge.transfer_rewrite_error`, and returns the wrapped rewrite failure. Typed forwarding and raw reads that already exposed a decoded packet have no retained `PacketRead`; they now terminate through the same captured error path without a nil dereference.

The same conditional cleanup applies when recording the derived `packet.transfer_rewrite` view fails before the typed Transfer write. A closed or failing recorder must not turn that capture failure into a secondary nil dereference when the forwarding path has no retained raw packet.

Listener callbacks that cannot return errors also forward recorder failures to the run's `FailureSink`. The `RequestChunkRadius` decode and typed-forwarding hooks therefore stop the session when their error event cannot be persisted, rather than continuing with an incomplete canonical capture.

The same boundary applies to lifecycle failures. Spawn failures, upstream dial failures, and transfer route-probe outcomes now join any error returned while recording their diagnostic event and propagate that capture failure through the run sink. A route-probe network failure itself remains a non-fatal routing hint, but a missing route-probe event is no longer silently accepted.

Upstream setup cleanup also preserves provenance. If resource-pack recording, downstream pack preparation, or the `upstream.connected` event fails after a connection has opened, the returned error now includes any failure from closing that observed connection. Cleanup still occurs before the setup path returns.

## Representative packet-family coverage

The loopback regression test `TestRunnerForwardsBidirectionalPacketsAndCapturesSession` sends representative packets through both sides of the proxy and checks both forwarding and capture. It covers `PlayerSkin` with skin and cape bytes, `AddActor`, `LevelChunk`, `SubChunkRequest`, `UpdateBlock`, `InventoryContent`, `InventoryTransaction`, `MovePlayer`, and `Emote`.

| Observable data | Current preservation |
| --- | --- |
| Skin and cape | Login client metadata and `PlayerSkin` are captured. Binary fields in the structured view have size, SHA-256, and preview metadata; the complete encoded packet remains in the raw payload blob. |
| Entities | Entity spawn and movement packets are retained as raw payloads and decoded views, including metadata exposed by the active protocol model. |
| Chunks and world data | `LevelChunk` and sub-chunk request traffic is retained, including the complete chunk payload bytes. The view does not reconstruct a semantic world model. |
| Block updates | `UpdateBlock` and related packet traffic use the same raw plus decoded capture path. |
| Inventory | Inventory content, slot, and transaction packets retain raw payloads and decoded item fields, including fields exposed by the active protocol model. |

This coverage does not add specialized commands or a second capture subsystem. It verifies the existing canonical session path. Packet families that gophertunnel does not decode still retain their raw payload when the packet header is available.

## Remaining limits

- Transfer packets are retained. Automatic hop following is opt-in through `--follow-transfers`; the original transfer remains authoritative and each followed hop is captured with its own hop and connection identifiers.
- Raw UDP and RakNet recovery evidence requires a lower observation boundary.
- The first transport payload can legitimately report `not_negotiated` before a NetworkSettings packet has been processed. The event still does not expose raw UDP or RakNet framing.
- Decoded snapshots reflect gophertunnel's active protocol model. Future analysis must prefer raw packet blobs when checking a changed or disputed definition.
- Full login metadata is sensitive. Structured snapshots make selected identifiers easier to search even though the raw Login packet was already present.

These are explicit limits, not a reason to add specialized download commands or duplicate an external proxy architecture.
