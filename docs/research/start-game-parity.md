# StartGame parity research

## Question

Does the terminating bridge preserve the complete world-start metadata that the upstream server sent, or does it reconstruct a smaller approximation for the downstream client?

## Confirmed observation

The vendored gophertunnel `GameData` type is intentionally a loose wrapper around only part of `StartGame`. Its `Conn.handleStartGame` implementation copied selected fields into `GameData`, while the proxy later called the listener-side `StartGame` serializer. That serializer supplied independent defaults for fields such as generator, command enablement, education features, broadcast flags, level identifiers, and telemetry metadata.

This was an observable protocol-boundary discrepancy even when gameplay packets were otherwise raw-forwarded. A server can use those fields to configure client behavior, so replacing them is not equivalent to a vanilla client receiving the server's packet.

## Narrow fix

`GameData.StartGamePacket` now retains a copy of the complete decoded packet on dialled connections. The listener-side serializer uses that packet when it is present. Callers that construct `GameData` directly retain the previous compact behavior. The change does not mutate the upstream packet or introduce packet filtering, injection, replay, or detector-specific logic.

## Evidence and provenance

- Vendored implementation: `third_party/gophertunnel/minecraft/game_data.go`, `conn.go`, and `internal/proxy/proxy.go` at the project revision that introduced FIX-0059.
- Upstream gophertunnel `GameData` remains a partial wrapper at commit [`7a556a07335b663744b50d38062636ad8283f314`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/game_data.go), reviewed 2026-09-21.
- The packet definition and serialization order are in the same upstream revision's [`StartGame`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/protocol/packet/start_game.go) type.

## Validation

- `go test ./...` in the nested gophertunnel module passed.
- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed.
- The proxy integration test sends non-default `StartGame` fields from an owner-controlled loopback server and verifies them through the downstream decoded packet.
- No public server was contacted. No account, server policy, or enforcement behavior was changed or tested.

## Limitations

This preserves decoded `StartGame` fields through the terminating bridge. It does not make the two authenticated login chains identical, and it does not claim byte identity for the separate DimensionData, JigsawStructureData, VoxelShapes, ItemRegistry, or spawn-handshake packets.
