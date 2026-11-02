# AUDIT-0254 - Add the Go network-client Golem foundation

- Commit subject: `Add the Go network-client Golem foundation`
- Change type: optional agent-controlled protocol client
- Runtime impact: adds a new `golem network` CLI mode. Existing proxy forwarding and PHP mailbox behavior remain available.

## Intent and scope

After the PHP GolemAgent companion moved to `NhanAZ-Plugins/GolemAgent`, the
maintainer requested the Go implementation for BedrockDebugProxy. This change
adds the first network-client foundation instead of copying the server-side
PHP simulation into Go.

The client resolves direct endpoints and experience selectors, uses the existing
gophertunnel login and resource-pack exchange, spawns into the world, reads
bounded server events, inventories forms, submits form JSON, sends chat, follows
RakNet and supported NetherNet Transfer hops, and writes a private local report.
The old mailbox command remains transitional for compatibility.

## Affected files and interfaces

- `internal/golemnet` contains the network session, bounded event/report model,
  form parser, Transfer state machine, and loopback coverage.
- `cmd/bedrock-debug-proxy/golem_network.go` adds `golem network` and its JSONL
  scenario operations.
- Root help, `docs/golem.md`, `docs/golem-network.md`, the decision record, and
  the changelog describe the new interface and evidence boundary.

## Provenance and validation

The implementation uses the pinned in-tree gophertunnel and go-raknet modules,
the existing `internal/experience` resolver, and the existing authentication
cache. No source is copied from EasySpecter, FakePlayer, Specter, or another
client implementation. No credentials, captures, or decrypted packs are
included.

Focused loopback tests cover form parsing and selection, chat delivery, bounded
reports, Transfer following, and the next-hop spawn event. The full project
quality gate and exact-revision CI are required before this commit is published.

## Live status and limitations

No public server was contacted by this implementation step. No retail client
was opened. The network report is not release evidence because the client does
not yet emit the full post-spawn client input stream, canonical BedrockDebugProxy
captures, or decrypted pack artifacts. Follow-up work must add those boundaries
and then validate the client against the required server categories.
