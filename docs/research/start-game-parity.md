# StartGame parity research

## Question

Does the terminating bridge preserve the complete world-start metadata that the upstream server sent, or does it reconstruct a smaller approximation for the downstream client?

## Confirmed observation

The vendored gophertunnel `GameData` type is intentionally a loose wrapper around only part of `StartGame`. Its `Conn.handleStartGame` implementation copied selected fields into `GameData`, while the proxy later called the listener-side `StartGame` serializer. That serializer supplied independent defaults for fields such as generator, command enablement, education features, broadcast flags, level identifiers, and telemetry metadata.

This was an observable protocol-boundary discrepancy even when gameplay packets were otherwise raw-forwarded. A server can use those fields to configure client behavior, so replacing them is not equivalent to a vanilla client receiving the server's packet.

## Narrow fix

`GameData.StartGamePacket` now retains a copy of the complete decoded packet on dialled connections. The listener-side serializer uses that packet when it is present. Callers that construct `GameData` directly retain the previous compact behavior. The change does not mutate the upstream packet or introduce packet filtering, injection, replay, or detector-specific logic.

The same capture review found that packet ID 313 (`JigsawStructureData`) and packet ID 337 (`VoxelShapes`) were decoded during upstream login but were not retained by the connection state. The listener therefore generated empty versions and the deferred upstream packets were forwarded later, producing two downstream copies. The login state now accepts and retains both packet types after resource-pack negotiation, and the listener reuses them when present.

The local capture `captures/session-20260921T003441Z` also contained an `ItemRegistry` packet with equal upstream and downstream sizes but different wire hashes. Decoding both packets produced 2,076 entries with equal semantic values. The first difference was inside an NBT compound, which shows that typed re-encoding can change map-entry order even when the decoded registry is unchanged. A focused probe of the same capture found the same wire instability for `StartGame` and `JigsawStructureData`; `VoxelShapes` happened to round-trip byte-for-byte in that capture. The bridge now retains the complete raw `DimensionData`, `JigsawStructureData`, `StartGame`, `VoxelShapes`, and `ItemRegistry` packets and writes them unchanged when the downstream protocol matches.

In the affected local capture `captures/session-20260921T005313Z`, the upstream packet IDs occurred at sequences 40 and 41, while the downstream stream contained generated copies at 254 and 255 and later forwarded copies at 1019 and 1022. The affected Hive capture `captures/session-20260920T135412Z` showed the same shape with upstream packets at 983 and 985 and downstream copies at 1397, 1401, 4017, and 4023. These are capture observations of startup duplication, not an inference about any server-side policy.

## Evidence and provenance

- Vendored implementation: `third_party/gophertunnel/minecraft/game_data.go`, `conn.go`, and `internal/proxy/proxy.go` at the project revision that introduced FIX-0059.
- Upstream gophertunnel `GameData` remains a partial wrapper at commit [`7a556a07335b663744b50d38062636ad8283f314`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/game_data.go), reviewed 2026-09-21.
- The packet definition and serialization order are in the same upstream revision's [`StartGame`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/protocol/packet/start_game.go) type.
- The raw startup retention path is implemented in the vendored `minecraft.GameData`, `Conn` login handlers, and listener-side startup writer. It validates the packet ID and negotiated protocol before writing retained bytes. Typed reconstruction remains the fallback when the raw packet is absent or the protocol differs.

## Validation

- `go test ./...` in the nested gophertunnel module passed.
- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed.
- The proxy integration test sends non-default `StartGame` fields from an owner-controlled loopback server and verifies them through the downstream decoded packet.
- The same integration test sends non-empty Jigsaw and VoxelShapes data and verifies that the downstream client retains one decoded copy of each.
- The same integration test supplies non-default raw `JigsawStructureData`, `VoxelShapes`, `StartGame`, and `ItemRegistry` packets and verifies that the downstream connection retains the exact startup wire bytes.
- No public server was contacted. No account, server policy, or enforcement behavior was changed or tested.

## Limitations

This preserves raw `DimensionData`, JigsawStructureData, `StartGame`, `VoxelShapes`, and `ItemRegistry` bytes through the terminating bridge when protocol IDs match and the packets arrive in the supported pre-spawn boundary. It does not make the two authenticated login chains identical, and it does not change the ordering state machine. Packets that first arrive after the connection is fully spawned remain ordinary forwarded traffic. When a raw packet cannot be validated, the existing typed path remains active and the write warning is observable in the operator log.
