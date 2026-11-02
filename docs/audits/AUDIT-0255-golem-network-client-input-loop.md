# AUDIT-0255 - Add the Go network-client post-spawn input loop

- Commit subject: `Keep the Go network-client Golem session alive after spawn`
- Change type: optional agent-controlled protocol client
- Parent decision: [0017 Go network-client Golem](../decisions/0017-go-network-client-golem.md)

## Intent and scope

The first Go network-client Golem implementation could complete login,
resource-pack exchange, and spawn, but it did not emit any post-spawn client
input. This increment adds the smallest useful client behavior for stability
investigation: a bounded stationary `PlayerAuthInput` heartbeat at one packet
per Bedrock tick after spawn. The loop follows the active connection after a
Transfer and stops with the session.

The CLI also exposes `stability`, which waits for a bounded window and requires
at least one emitted client-input packet before returning the supplemental
report. The report records the configured interval, packet count, and last
send time.

## Protocol basis

The vendored gophertunnel packet definition documents `PlayerAuthInput` as the
client packet sent once per tick for server-authoritative movement. The client
uses the post-`DoSpawn` `minecraft.GameData` position and rotation, empty input
flags, the negotiated client input mode, normal play mode, and a derived camera
orientation. It does not attempt movement, inventory prediction, block
interaction, packet injection, replay, or anti-cheat evasion.

## Affected files and interfaces

- `internal/golemnet/client.go` adds the input loop, packet construction,
  bounded interval validation, and report fields.
- `internal/golemnet/client_test.go` verifies a loopback server receives
  stationary `PlayerAuthInput` packets with monotonic ticks and spawn
  position/rotation.
- `cmd/bedrock-debug-proxy/golem_network.go` adds the `stability` operation.
- `docs/golem-network.md`, the decision record, and `CHANGELOG.md` document the
  new boundary and its limitations.

## Security and evidence

The loop writes only the supported client packet through gophertunnel. It does
not weaken authentication, alter transport policy, or make server-specific
claims. Server data remains untrusted and no credentials, captures, or
decrypted resource packs are added. The report remains supplemental and cannot
replace a canonical capture or release validation report.

## Validation

The focused loopback test must observe three input packets, verify monotonic
ticks and the configured spawn state, and close cleanly. The full project
quality gate, race-enabled tests, build, and exact-revision CI remain required
before publication.

No public server was contacted by this implementation step. No retail client
was opened. A later live validation should use the Go Golem CLI against an
owner-controlled disposable server and then a target public server category,
checking the input packet count, server disconnect behavior, and any capture
or log evidence separately from the supplemental report.
