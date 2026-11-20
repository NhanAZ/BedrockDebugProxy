# Go Golem server teleport acknowledgement

## Finding

The Go network-client Golem previously emitted its stationary `PlayerAuthInput`
heartbeat but had no response path for a server `MovePlayer` teleport or reset
after spawn. The existing BedrockPackArchive Mineville route client already
acknowledges the current player's server-directed region reset. This was a
missing action in BedrockDebugProxy's client, not evidence that the server or
protocol definition was faulty.

The Golem reader now acknowledges only packets addressed to its active
connection's player runtime ID and whose gophertunnel mode is `Teleport` or
`Reset`. It uses the packet's position and rotation, a strictly increasing
input tick, `HandledTeleport` and `PersistSneak` flags, and
`VerticalCollision` when `OnGround` is true. It ignores other entity IDs,
ordinary movement, paused input, and stale connections.

## Protocol evidence

The reviewed source is Mojang's `bedrock-protocol-docs` revision
`f8d1496b5efc95105bd9ae1f2f2f076a43f71d82`, the `v1.26.62` schema release for
protocol 2225, reviewed on 2026-11-20.

- [`PlayerAuthInputData.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82/json/PlayerAuthInputData.json)
  lists `HandledTeleport`, `PersistSneak`, and `VerticalCollision` as input
  flags.
- [`PlayerAuthInputPacketPayload.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82/json/PlayerAuthInputPacketPayload.json)
  defines the input flag set, predicted position, rotation, camera orientation,
  and client tick fields used by the response.
- [`MovePlayerPacketPayload.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82/json/MovePlayerPacketPayload.json)
  defines the target runtime ID, position mode, position, rotation, ground
  state, and teleport data. Its wire enum calls value 1 `Respawn` and value 2
  `Teleport`. The selected gophertunnel packet API names value 1
  `MoveModeReset` and value 2 `MoveModeTeleport`.

These schemas establish the available fields and flag names. They do not state
the required packet order for every server or explicitly specify the response
for every respawn/reset flow. Supporting both local mode constants follows the
existing BedrockPackArchive route helper and is covered by a loopback test. No
Mojang or gophertunnel code was copied, so this does not change source-code
provenance or the protocol baseline.

## Independent implementation cross-check

Dragonfly revision
[`0f9a3a5736267db60bf46e92d408f81be1de9a64`](https://github.com/df-mc/dragonfly/commit/0f9a3a5736267db60bf46e92d408f81be1de9a64)
is MIT licensed. Its session sends the player's server-directed teleport as a
`MovePlayer` packet in teleport mode and stores the expected position. Its
`PlayerAuthInput` handler ignores movement that is still more than one block
from that target, then clears the pending target after movement returns close
enough. This independently confirms that server-driven position changes need
client/server state synchronization. Dragonfly does not check the
`HandledTeleport` flag in this handler, so this comparison does not prove that
every server requires that flag or that every reset mode has identical rules.

- [`server/session/world.go` at the reviewed revision](https://github.com/df-mc/dragonfly/blob/0f9a3a5736267db60bf46e92d408f81be1de9a64/server/session/world.go#L246-L276)
- [`server/session/handler_player_auth_input.go` at the reviewed revision](https://github.com/df-mc/dragonfly/blob/0f9a3a5736267db60bf46e92d408f81be1de9a64/server/session/handler_player_auth_input.go#L43-L60)

## Validation and limits

`TestClientAcknowledgesServerTeleport` uses a synthetic loopback server. It
sends a teleport for another entity before teleports for the Golem's own
entity, then checks both the Teleport and Reset responses for the server's
position and rotation, correct ground state, `PersistSneak`, and increasing
ticks. No private capture, account data, or pack contents are used in the
fixture.

An exploratory Mineville session was run with an unstamped working-tree build,
but it is not cited here because its capture was not archived. It cannot serve
as exact-revision evidence, release validation, or proof that a previous
server failure had this cause. It does not establish behavior for other routes
or servers.

The change does not modify resource-pack offers, archive handling, capture
fields, or the BedrockPackArchive importer contract. No paired importer code
change is required. The producer and consumer checks still need to be run at
the final commit before publication.
