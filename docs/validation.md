# Real-world validation workflow

Automated tests establish that the implementation matches local expectations. They do not establish that those expectations match live Minecraft Bedrock servers. Runtime, networking, protocol, authentication, session, resource-pack, capture, encoding, and decoding changes need the scoped development evidence below. Official releases accept full authenticated CLI sessions or retail Minecraft client sessions. CLI evidence does not establish rendering or visible gameplay.

Changes that can affect the local web viewer have an additional same-change requirement. The commit must include the viewer asset and HTTP integration tests and the embedded JavaScript syntax check. Presentation or interaction changes must also be smoke-tested in a browser from a clean committed build. Use [`web-validation.md`](web-validation.md) for the exact scope and checklist.

For the end-to-end pull request checklist, use [`../CONTRIBUTING.md`](../CONTRIBUTING.md). For the ordered six-server release checklist, use [`releasing.md`](releasing.md). See [`server-targets.md`](server-targets.md) for endpoint references and the server-policy disclaimer. This document defines how to create and judge the validation evidence used by both workflows.

The automated suite includes a loopback integration session over RakNet. It exercises offline login, resource-pack negotiation with no packs, StartGame and spawn, typed packet forwarding in both directions, clean shutdown, raw packet capture, decoded events, connection metadata, and GameData snapshots. This catches local forwarding and capture regressions without accounts or public infrastructure. It does not exercise Microsoft authentication, the retail Minecraft client, public-server routing, transfers, live resource packs, or server-specific behavior and therefore does not replace the gates below.

## Development and pull requests

The Hive minimum baseline does not override server policy. As of 2026-09-20, Hive's [VPN and proxy policy](https://support.playhive.com/vpn/) says proxy use is not allowed, most proxy logins are rejected, and blocklists are updated continually. The maintainer's direct request to run an owner-controlled validation session is sufficient authorization for this internal workflow. Do not request or copy private messages into project artifacts. Do not use a punished account or a substitute account to investigate enforcement. Keep the report focused on observed technical evidence.

For a routine pull request with one focused runtime commit, `run --packs-only` is an acceptable live smoke check when the changed behavior occurs during upstream login or resource-pack exchange. Use The Hive as the general baseline, but choose a target that actually exercises the affected transport and flow for protocol-specific changes. The mode authenticates and connects directly upstream, records the resource-pack exchange, and closes without a local listener or retail client. Use a stamped binary for the exact revision and inspect the closed capture. This check does not request world spawn, test gameplay, or guarantee the pack offer a particular Minecraft client would receive.

A packs-only report is expected to have `result=incomplete` and `normal_session=not_observed`; it is scoped evidence for development pull requests and is never a release pass. Mark `packs_exchange=pass` only when the capture shows that the offered packs completed the supported exchange. If no pack is offered or the connection stops before exchange completes, record the relevant check as `not_observed` or `fail`. A general successful connection is not proof of a flow that the capture did not observe.

Changes outside the packs-only boundary still need focused automated, fixture, or loopback coverage for the affected flow. For example, a transfer-following change must test the `Transfer` path and next-hop behavior without treating a first-hop packs-only login as coverage. Resource-pack decryption changes require evidence of the supported decryption path. Record remaining live behavior as unobserved and complete the full CLI or retail session at the release gate.

For transfer-following changes, focused automated or loopback coverage must exercise the `Transfer` rewrite, a successful next-hop transport and upstream connection, and the next hop's resource-pack exchange when offered. When the route probe exposes it, inspect `client_guid_preserved: true` as evidence that the RakNet identity was carried into the next-hop handshake. A first-hop packs-only connection alone is not sufficient evidence for a transfer fix. The full client flow through a later spawn remains part of Enchanted's release validation.

The pre-reconnect route probe is a best-effort route hint. Its ping response or failure remains in the raw capture and analysis, and a failure marked `best_effort=true` is non-blocking only when the normal next-hop session still connects and reaches the required flow. The optional upstream address-discovery ping is bounded to five seconds for an initial dial and 750 ms for a transfer hop so a lost UDP ping cannot consume the full login and resource-pack deadline. Status-ping errors without the `best_effort` marker remain blocking.

The Enchanted release report has an additional mandatory flow. The server sends its form when the player joins. Select a known pack-bearing child, such as OP SMP, in that server-sent form and wait for the server to send `Transfer`. The report passes only when the proxy follows that transfer, captures the next-hop login and resource-pack exchange when offered, and reaches the minigame spawn. Connecting to the Enchanted hub without selecting a minigame is not sufficient. Record `transfer_following=pass` and `resource_pack_transfer=pass` for this flow.

CubeCraft has an additional maintainer risk gate. Before connecting, the maintainer must explicitly accept the possibility of rejection, kick, or ban for the exact diagnostic session. CubeCraft's public rules prohibit Bedrock proxies, so a risk acknowledgement must not be described as permission or a policy exception. If the acknowledgement is unavailable, do not connect and record `incomplete` with the relevant `risk_acknowledged=not_observed` check. The session must use the normal read-only proxy path and must not add bypass, mutation, replay, or anti-cheat evasion behavior.

Do not approve or merge a runtime-affecting pull request without exact-revision automated checks and applicable focused flow evidence. A packs-only report may be incomplete only because it ends before spawn; it does not waive a missing check for the changed flow. Documentation, comments, formatting, and changes proven not to affect runtime may be exempt. When impact is uncertain, require validation.

## Stamped test build

Commit the revision, ensure the working tree is clean, run the quality gate, and create a binary whose capture manifest records the exact Git commit.

```powershell
.\tools\quality.ps1
.\tools\build.ps1 -Version dev
.\bin\bedrock-debug-proxy.exe version
```

`tools/build.ps1` refuses a dirty tree. Do not use an unstamped `go build` binary for validation evidence because its capture cannot prove which revision was tested.

## Capture and report

For a development pull request, run the stamped binary in packs-only mode and generate the sanitized, scoped report with project tooling. There is no local listener, so no Minecraft client connection is needed.

```powershell
.\bin\bedrock-debug-proxy.exe `
    run `
    --upstream "experience:The Hive" `
    --auth device `
    --packs-only
.\tools\create-validation-report.ps1 `
    -Binary .\bin\bedrock-debug-proxy.exe `
    -Capture .\captures\<SESSION> `
    -Server "The Hive" `
    -Result incomplete `
    -Check @("normal_session=not_observed", "packs_exchange=pass") `
    -Output .\validation\local\<REVISION>\the-hive.json
```

Device authentication is cached in the selected default account after the first successful login. Use the account workflow in [`authentication.md`](authentication.md) when validation needs a different local account. Replace the target when the changed flow requires another server or transport, and name that target in the report. Do not combine `--packs-only` with `--listen` or `--follow-transfers`.

Use public server labels rather than addresses in `-Server`. A manual check uses `lowercase_name=pass`, `lowercase_name=fail`, or `lowercase_name=not_observed`. Add one for each flow reviewed, such as `resource_pack_transfer=pass` or `resource_pack_decryption=pass`.

A full-session report requires exactly one accepted client mode. Retail reports use `normal_session=pass`. CLI reports use `cli_session=pass`, `authenticated_upstream=pass`, and `post_spawn_stability=pass`. For CLI stability, review at least 30 seconds of post-spawn traffic without an unexplained disconnect or protocol error. Every other supplied review check must pass, and capture-derived evidence must prove all of the following facts.

- The capture is closed and free of verifier issues, write errors, dropped events, and truncated events.
- The upstream connected and the session negotiated, spawned, and closed.
- Client-to-server and server-to-client events were observed.
- Decoded packet events were observed.
- No packet decode, structured snapshot, resource-pack decryption, or other blocking error event was recorded.
- Sessions declaring automatic artifact folders have an error-free completed `artifacts/status.json` matching the capture ID, build revision, and final event sequence. See [session artifacts](session-artifacts.md) for the output and its limits.

The manifest `complete` flag remains in the report but is not by itself a pass gate. The current capture schema sets it to `false` whenever a known observation boundary is declared, including the intentional absence of raw UDP and RakNet acknowledgement frames. The stricter pass gate uses the verifier plus explicit write, dropped, and truncated counters so a documented scope boundary is not confused with data loss inside the supported boundary.

Normal connection termination produces read-side `transport.payload` and `bridge.read_error` evidence when a read loop ends. The first forwarding loop to terminate claims coordinated shutdown before its error is classified, so the peer loop cannot race the shutdown marker. A peer loop that is awakened by that close may return `context.Canceled`, `net.ErrClosed`, or gophertunnel's typed `minecraft.DisconnectError` for a received Bedrock `Disconnect`; those expected bridge errors remain in the `session.close` diagnostics but are not emitted as `bridge.*_error` events. A first write-side close error can be the loop that observes the disconnect, so `bridge.write_error` and `bridge.flush_error` with those termination errors are treated as the shutdown signal and omitted as well. The downstream `RequestChunkRadius` callback also omits a write error caused by that typed remote disconnect. The raw `Disconnect` packet remains in the capture. Other writes before a close, library errors, resource-pack derivation errors, and all other error kinds remain recorded and block an automatic pass. The report preserves termination counts without automatically treating them as protocol failures. The developer must still inspect them and may mark the session `fail` when their timing or cause is abnormal.

The report derives revision, timestamps, client and upstream versions, protocol IDs, traffic counts, resource-pack counts, and safe error-kind counts from the capture. It omits addresses, packet payloads, error messages, player identifiers, content keys, resource-pack metadata, decrypted assets, and the text of capture limitations because a dynamic limitation may contain a pack identifier. It retains only the limitation count. Review the JSON before uploading it. Keep the working capture local and sensitive by default. If a project document needs to cite a raw capture, archive that capture unchanged in its own private repository in [BedrockDebugProxy-Captures](https://github.com/BedrockDebugProxy-Captures), attach it to the `capture-archive` release, verify the repository privacy and archive SHA-256, then cite the release URL. Never use a local capture path as an evidence citation. Retain the local copy until the archive has been verified. Follow the full [capture citation workflow](../AGENTS.md#capture-evidence-citations).

If an anomaly exists, use `-Result fail`. If a server or required flow could not be tested, use `-Result incomplete` and the applicable `not_observed` check. Never omit a failed or unavailable baseline silently. A report is evidence of one observed session, not a guarantee for other servers or later revisions.

Generate candidate reports under the ignored `validation/local/<REVISION>/` directory. Upload the sanitized JSON as pull request or release evidence without adding it to the candidate commit. Committing a report would change the pull request head or release revision after the tested binary was stamped. Keep any long-term archive as a release asset or in a deliberate later documentation-only record that preserves the original `tested_revision`.

Do not commit `.bdpcap`, raw captures, authentication state, resource-pack keys, or third-party assets.

Add `--decrypt-resource-packs` only when the change or release test explicitly needs to validate the authorized resource-pack decryption flow. It is not part of the default connection sanity test.

## Release gate

Every official release requires a full authenticated CLI session or a retail Minecraft client session and a passing report tied to a validated runtime revision for each baseline server. A retail application is optional. See [the maintainer decision](decisions/0015-cli-release-validation.md).

- The Hive
- Galaxite
- Lifeboat
- Mineville Zeqa (`experience:Mineville Zeqa`)
- Enchanted (`experience:Enchanted`)
- CubeCraft (`experience:CubeCraft`)

For each server, check connection, authentication, resource packs when offered, spawn, both traffic directions, packet decoding, stability, and unexplained protocol or decode errors. Do not assume the six servers use the same packet order, timing, optional packets, software, or infrastructure.

For CubeCraft, also require `risk_acknowledged=pass` before accepting `result=pass`. The public policy warning remains in force, and the acknowledgement does not imply that CubeCraft has granted an exception.

A baseline failure starts an investigation. Determine whether the cause is a proxy bug, protocol misunderstanding, valid server behavior, or an external outage before changing code. Do not add a server-specific workaround merely to make the release matrix green. If an external condition prevents testing, retain an `incomplete` report and keep the release not ready until the maintainer explicitly resolves the gate.

After generating all six reports for one runtime revision, run
`tools/check-release-readiness.ps1 -Revision <40-character-commit>`. It checks that every required label has a
passing report with automatic and manual gates for that exact revision. It does not replace review of the local captures by the operator or agent.

### CLI session workflow

Start the stamped proxy with device authentication, a concrete loopback listener, and the existing self-signed downstream opt-in. Connect a reviewed CLI protocol client to that listener, accept offered resource packs, complete spawn, consume packets, and perform ordinary required protocol responses for at least 30 seconds after spawn. Retain the driver revision or exact local source with the private evidence. Close the client and let the proxy finalize artifacts and verification. Use a single approved account session at a time.

```powershell
.\bin\bedrock-debug-proxy.exe run `
    --listen "127.0.0.1:19134" `
    --upstream "experience:The Hive" `
    --auth device `
    --allow-unauthenticated-client `
    --follow-transfers `
    --decrypt-resource-packs
```

The self-signed downstream is confined to loopback. Upstream Microsoft authentication remains required. This is an existing diagnostic opt-in and does not change the default listener authentication. Packs-only and server-side Golem simulations cannot satisfy this workflow. For a retail client, use the normal authenticated downstream path instead.

Generate a new report after reviewing the closed capture. Keep older incomplete reports unchanged.

```powershell
.\tools\create-validation-report.ps1 `
    -Binary .\bin\bedrock-debug-proxy.exe `
    -Capture .\captures\<CLI-SESSION> `
    -Server "The Hive" `
    -Result pass `
    -Check @("cli_session=pass", "authenticated_upstream=pass", "post_spawn_stability=pass", "experience_resolution=pass") `
    -Output ".\validation\local\<REVISION>\the-hive.json"
```

For Enchanted, also supply transfer_following=pass and resource_pack_transfer=pass after the known child reaches spawn and the stability interval. For CubeCraft, also supply risk_acknowledged=pass after the maintainer risk gate. CLI reports establish the observed protocol flow, not retail rendering, input handling, or gameplay. Release notes must identify this evidence type.

### Documentation-only release delta

If a candidate is created after the live matrix and the committed diff from the validated runtime revision to
the candidate contains only Markdown documentation, the existing reports may be reused without changing their
`tested_revision`. This is an explicit evidence-preserving exception, not permission to edit report facts. The
validated revision must be an ancestor of the candidate, and the checker must verify that the diff contains only
Markdown files in the root release documents or `docs/`, plus the release-readiness checker itself when its
allowlist logic is the only non-documentation change.

Run the gate with both revisions:

```powershell
.\tools\check-release-readiness.ps1 `
    -Revision <candidate-40-character-commit> `
    -ValidatedRevision <runtime-40-character-commit>
```

The release notes and audit record must identify the candidate and validated runtime revisions. Any change to
product code, tools other than the release checker, workflows, configuration, generated files, or another
non-approved path requires a fresh exact-revision matrix. The exception still requires a clean tree, quality gate,
build, CI, documentation review, and
the normal release and backup checks.
