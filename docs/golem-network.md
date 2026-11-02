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
{"op":"observe","duration":"30s"}
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
- `observe` waits and returns the current report snapshot.
- `report` returns the current report snapshot immediately.
- `quit` closes the session and stops the scenario.

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

The report schema is `bedrockdebugproxy.golem.network.v1`. It contains bounded
event counts, packet type counts, forms, transfers, text messages, resource-pack
count, hop count, spawn state, and errors. It intentionally does not contain
raw packet bytes, authentication tokens, or decrypted pack contents.

This report is not a release validation report. The current network bot does
not yet send the full stream of post-spawn client input and gameplay packets,
does not write canonical BedrockDebugProxy captures, and does not derive
decrypted resource-pack archives. It is therefore useful for login, form,
transfer, pack-count, and spawn investigations, but it cannot by itself satisfy
`cli_session=pass`, post-spawn stability, or the six-server release gate.
The existing proxy capture workflow and its reviewed CLI protocol driver remain
the release evidence path until those boundaries are implemented and verified.

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
