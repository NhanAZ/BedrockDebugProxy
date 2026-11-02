# Go network-client Golem

`golem network` is the Go network client for agent-controlled Bedrock server
scenarios. It uses the pinned gophertunnel and go-raknet implementations already
used by BedrockDebugProxy. It does not open Minecraft for Windows and it does
not use the PHP GolemAgent plugin.

The client resolves a literal `HOST:PORT` or an `experience:NAME` selector,
authenticates when requested, completes the upstream login and resource-pack
exchange, calls `DoSpawn`, and then reads server packets into a bounded event
stream. It recognizes server text, forms, disconnects, and Transfer packets.
Transfer following is enabled by default. RakNet hops preserve the previous
source address and client GUID when the transport exposes them. A NetherNet
transfer first probes the HTTPS endpoint used by the proxy and falls back to
RakNet when the endpoint does not advertise that transport.

## Command

```powershell
.\bin\bedrock-debug-proxy.exe golem network `
    --server experience:Enchanted `
    --auth device `
    --report .local\golem-network\enchanted.json `
    --capture .local\golem-network\enchanted-capture `
    --decrypt-resource-packs `
    --validation-report validation\Enchanted.golem.json `
    --server-label Enchanted `
    --validation-result incomplete `
    --input .local\golem-network\enchanted.jsonl
```

Use `--auth none` only for an operator-controlled local server with
authentication disabled. The default `device` mode reads the existing
BedrockDebugProxy account cache. The cache remains a credential and is never
included in stdout or the report.

The report path must not already exist. Reports are created with mode `0600`
and include server-sent form JSON, so keep them private. Do not upload a report
or paste its form data into a public issue without reviewing it first.

## Scenario JSON Lines

The command writes one JSON response for the initial `connect` operation and
one response for each scenario line. Blank lines are ignored. Unknown fields,
oversized fields, invalid JSON, and unsupported operations fail before the
operation is executed.

```jsonl
{"op":"wait_form","id":"hub"}
{"op":"select_form","id":"choose-mines","index":0}
{"op":"wait_form","id":"nested"}
{"op":"select_form","id":"choose-hop","index":0}
{"op":"wait","duration":"5s"}
{"op":"stability","duration":"30s"}
{"op":"report"}
{"op":"quit"}
```

Supported operations are:

- `wait_form` waits for the next server-sent form and returns its bounded
  metadata, buttons, and raw JSON.
- `current_form` returns the latest form without consuming an event.
- `select_form` submits a button index from the latest form. It is intended
  for menu forms and rejects indexes outside the observed button list.
- `submit_form` submits a caller-supplied JSON response for the latest form.
  This is needed for custom and modal forms. The response is validated as JSON
  and capped at 1 MiB.
- `chat` sends one client chat message. Server replies appear as later text
  events.
- `wait` lets the reader continue collecting events for a bounded duration.
- `stability` waits for a bounded post-spawn window while the client sends a
  stationary `PlayerAuthInput` heartbeat, then returns the report and fails if
  no client input packet was emitted.
- `observe` waits and returns the current report snapshot.
- `report` returns the current report snapshot immediately.
- `quit` closes the session and stops the scenario.

When `--capture` is provided, the command writes the canonical
`bedrockdebugproxy.capture.v1` directory. It retains post-transport Bedrock
packet payloads and decoded packet views in both directions, connection and
spawn lifecycle evidence, bounded GameData, and resource-pack archives. Add
`--decrypt-resource-packs` for release work so encrypted offers also produce
the derived plaintext archive and `contents.json` evidence. The original
encrypted archive remains the authoritative local evidence.

`--validation-report` runs the repository's existing
`tools/create-validation-report.ps1` after the capture closes. The command
automatically supplies `cli_session`, `authenticated_upstream`, and
`post_spawn_stability` from observed evidence. Use repeated `--validation-check`
options for review facts that only an agent can establish, such as
`enchanted_form_review=pass` and `enchanted_route_matrix=pass`. The default
result is `incomplete`; `--validation-result pass` is rejected by the existing
report generator unless every automatic and manual gate passes. The binary
must be built with `tools/build.ps1` so the capture contains the exact Git
revision required by the report generator.

## Enchanted workflow

The network client does not assume that Enchanted has a fixed child count or a
fixed hop count. A scenario should inspect the form returned after Hub spawn,
select an observed top-level entry, and inspect the next form if the selection
returns another form. When the server emits `Transfer`, the default client
follows it, completes the next login and resource-pack exchange, and emits a
new `session.spawned` event with the next hop number. A route-specific scenario
can then wait for another form or observe the terminal server.

The route matrix remains an agent responsibility. Return to Hub by starting a
fresh session before testing the next top-level entry. A maintenance/status
entry, a navigation entry, a direct route, and a selection without Transfer
are valid observed outcomes and must be classified from the returned events.

## Report boundary

The supplemental report schema is `bedrockdebugproxy.golem.network.v1`. It
contains bounded event counts, packet type counts, forms, transfers, text
messages, resource-pack count, hop count, spawn state, stationary
client-input count and rate, and errors. It intentionally does not contain
raw packet bytes, authentication tokens, or decrypted pack contents.

The canonical capture and validation-report options now make the Go network
client usable as a release-report producer. The release gate still reviews
the report and its capture: a normal `stability` operation must cover at least
30 seconds, encrypted packs must be decrypted successfully, and server-specific
workflow checks remain agent responsibilities. The client currently provides
a stationary `PlayerAuthInput` heartbeat and scenario-driven forms, chat, and
transfer following. It does not yet claim full retail-client movement,
inventory, combat, block interaction, or rendering gameplay.

## Security and limits

The client treats server form data, text, packet types, transfer targets, and
scenario JSON as untrusted input. It validates transfer host and port, limits
forms and messages, bounds retained events, refuses an event-buffer overflow,
limits transfer hops to 64, and never constructs a shell command from scenario
data. A transfer is followed only after the target has passed the address
validator. The client does not mutate, replay, inject, or drop packets.

Use a disposable account and keep reports and local captures separate from
public source. `--auth none` is for local development only. Public-server
validation remains subject to the existing server policies and release
workflow.
