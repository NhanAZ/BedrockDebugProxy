# StartGame parity research

## Question

Does the terminating bridge preserve the complete world-start metadata that the upstream server sent, or does it reconstruct a smaller approximation for the downstream client?

## Confirmed observation

The vendored gophertunnel `GameData` type is intentionally a loose wrapper around only part of `StartGame`. Its `Conn.handleStartGame` implementation copied selected fields into `GameData`, while the proxy later called the listener-side `StartGame` serializer. That serializer supplied independent defaults for fields such as generator, command enablement, education features, broadcast flags, level identifiers, and telemetry metadata.

This was an observable protocol-boundary discrepancy even when gameplay packets were otherwise raw-forwarded. A server can use those fields to configure client behavior, so replacing them is not equivalent to a vanilla client receiving the server's packet.

## Narrow fix

`GameData.StartGamePacket` now retains a copy of the complete decoded packet on dialled connections. The listener-side serializer uses that packet when it is present. Callers that construct `GameData` directly retain the previous compact behavior. The change does not mutate the upstream packet or introduce packet filtering, injection, replay, or detector-specific logic.

The same capture review found that packet ID 313 (`JigsawStructureData`) and packet ID 337 (`VoxelShapes`) were decoded during upstream login but were not retained by the connection state. The listener therefore generated empty versions and the deferred upstream packets were forwarded later, producing two downstream copies. The login state now accepts and retains both packet types after resource-pack negotiation, and the listener reuses them when present.

The local capture [captures/session-20260921T003441Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T003441Z/releases/tag/capture-archive) also contained an `ItemRegistry` packet with equal upstream and downstream sizes but different wire hashes. Decoding both packets produced 2,076 entries with equal semantic values. The first difference was inside an NBT compound, which shows that typed re-encoding can change map-entry order even when the decoded registry is unchanged. A focused probe of the same capture found the same wire instability for `StartGame` and `JigsawStructureData`; `VoxelShapes` happened to round-trip byte-for-byte in that capture. The bridge now retains the complete raw `DimensionData`, `JigsawStructureData`, `StartGame`, `VoxelShapes`, and `ItemRegistry` packets and writes them unchanged when the downstream protocol matches.

In the affected local capture [captures/session-20260921T005313Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T005313Z/releases/tag/capture-archive), the upstream packet IDs occurred at sequences 40 and 41, while the downstream stream contained generated copies at 254 and 255 and later forwarded copies at 1019 and 1022. The affected Hive capture [captures/session-20260920T135412Z](https://github.com/BedrockDebugProxy-Captures/session-20260920T135412Z/releases/tag/capture-archive) showed the same shape with upstream packets at 983 and 985 and downstream copies at 1397, 1401, 4017, and 4023. These are capture observations of startup duplication, not an inference about any server-side policy.

The same local capture exposed a separate CreativeContent boundary. Packet ID 145 arrived from the upstream server at sequence 65 with a 49,283-byte raw payload. The downstream stream then contained a generated 2-byte packet at sequence 455 followed by the original 49,283-byte payload at sequence 1077. The 2-byte packet is gophertunnel's listener compatibility response from `handleRequestChunkRadius`, while the later packet is the upstream server packet forwarded after spawn. This is confirmed proxy-generated duplication, not a server-policy signal.

The same capture also shows why the spawn acknowledgements must not be compared as one continuous stream. The upstream connection consumed `ChunkRadiusUpdated` (ID 70) at sequence 70 and `PlayStatusPlayerSpawn` (ID 2) at sequence 175 while completing its own dialled login. The downstream listener then emitted ID 70 and ID 2 at sequences 622 and 623 in response to the downstream client's independent `RequestChunkRadius`. The ID 70 payload hash is `36a9e7f1c95b82ffb99743e0c5c4ce95d83c9a430aac59f84ef3cbfab6145068` on both channels. The spawn-status ID 2 payload hash is `88185d128d9922e0e6bcd32b07b6c7f20f27968eab447a1d8d1cdf250f79f7d3` on both channels. The upstream login handler consumes these packets before the forwarding loops start, so forwarding the upstream copies would either be impossible at that boundary or would add a second downstream response. This is a confirmed independent-session boundary, not a lost packet or an extra response.

The source path is `minecraft/conn.go`: `handleChunkRadiusUpdated` and `handlePlayStatus` complete the dialled connection's spawn state, while `handleRequestChunkRadius` writes the listener-side ID 70 and ID 2 response. `internal/proxy/proxy.go:startGame` starts both state machines before the ordinary forwarding loops. The local capture therefore has different sequence positions but matching response bytes. Any future change to this boundary must first prove that it can retain and substitute the consumed packets without delaying the downstream client or changing the required spawn handshake.

## Post-fix vanilla-client loopback

On 2026-09-21, a real Windows Bedrock 1.26.51 client completed a controlled loopback session through the binary built from commit `d272805fd571525cca6f7473b166fb1129bafd48`. The upstream was the owner-controlled local PocketMine server with Xbox authentication temporarily disabled for this test and restored afterward. The client completed login, resource-pack negotiation, spawn, normal play, and quit.

The closed capture [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive) contains 10,645 events and 4,222 blobs. It has zero dropped events, write errors, truncation, or decode errors. The analysis view reports protocol 2193 and game version 1.26.51 on both connections, matching ClientData game version and device OS, and a complete negotiated, spawned, and closed lifecycle. Raw packet counts match for every packet ID and direction between upstream and downstream. The retained startup payloads are byte-identical across the two channels for JigsawStructureData (ID 313, 209,307 bytes), VoxelShapes (ID 337, 23,178 bytes), StartGame (ID 11, 73,081 bytes), ItemRegistry (ID 162, 113,448 bytes), and CreativeContent (ID 145, 49,283 bytes).

This is confirmed local fidelity evidence for the current same-protocol terminating path. The different sequence numbers reflect the two independent login chains and the listener's bounded login buffering; they are not evidence that the proxy changed the startup payloads. The run does not establish authenticated public-server behavior, raw UDP or RakNet parity, or mixed URL resource-pack offer behavior.

## Evidence and provenance

- Vendored implementation: `third_party/gophertunnel/minecraft/game_data.go`, `conn.go`, and `internal/proxy/proxy.go` at the project revision that introduced FIX-0059.
- Upstream gophertunnel `GameData` remains a partial wrapper at commit [`7a556a07335b663744b50d38062636ad8283f314`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/game_data.go), reviewed 2026-09-21.
- The packet definition and serialization order are in the same upstream revision's [`StartGame`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/protocol/packet/start_game.go) type.
- The raw startup retention path is implemented in the vendored `minecraft.GameData`, `Conn` login handlers, and listener-side startup writer. It validates the packet ID and negotiated protocol before writing retained bytes. Typed reconstruction remains the fallback when the raw packet is absent or the protocol differs.
- For the CreativeContent boundary, the vendored listener retains its normal empty compatibility response for ordinary callers. The proxy marks only its terminating downstream connection to suppress that generated response, because the upstream packet remains available to the forwarding loop. The option is not a packet filter and does not alter an upstream CreativeContent packet.

## Validation

- `go test ./...` in the nested gophertunnel module passed.
- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed.
- The proxy integration test sends non-default `StartGame` fields from an owner-controlled loopback server and verifies them through the downstream decoded packet.
- The same integration test sends non-empty Jigsaw and VoxelShapes data and verifies that the downstream client retains one decoded copy of each.
- The same integration test supplies non-default raw `JigsawStructureData`, `VoxelShapes`, `StartGame`, and `ItemRegistry` packets and verifies that the downstream connection retains the exact startup wire bytes.
- The same integration test supplies a non-empty CreativeContent packet after spawn and verifies that the downstream client receives its decoded contents once and that the upstream and downstream raw hashes match.
- No public server was contacted. No account, server policy, or enforcement behavior was changed or tested.

## Limitations

This preserves raw `DimensionData`, JigsawStructureData, `StartGame`, `VoxelShapes`, and `ItemRegistry` bytes through the terminating bridge when protocol IDs match and the packets arrive in the supported pre-spawn boundary. It does not make the two authenticated login chains identical, and it does not change the ordering state machine. Packets that first arrive after the connection is fully spawned remain ordinary forwarded traffic. When a raw packet cannot be validated, the existing typed path remains active and the write warning is observable in the operator log.

The CreativeContent suppression applies only to the generated listener compatibility packet on the terminating proxy path. If an upstream server sends its own CreativeContent packet, the proxy still forwards and captures it as ordinary server-to-client traffic. The owner-controlled real-client loopback covers this distinction; public-server validation remains pending.
